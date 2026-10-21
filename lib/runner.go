package ccode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	templateassets "github.com/cohesivestack/ccode/template"
	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

type RunnerContext struct {
	ccodeContext *Context
	runtime      *goja.Runtime
	jsonObject   *goja.Object
	jsonParse    goja.Callable
	stdout       io.Writer
}

func Run(config *Config, process string) error {
	return NewContext(config).Run(process)
}

func (ctx *Context) Run(process string) error {
	if err := ctx.validateWorkspaceSupportFiles(); err != nil {
		return err
	}

	entryPointPath, sourcePath, err := ctx.resolveProcessEntryPoint(process)
	if err != nil {
		return err
	}
	if ctx.activeProcesses == nil {
		ctx.activeProcesses = map[string]struct{}{}
	}
	processKey := filepath.Clean(sourcePath)
	if _, running := ctx.activeProcesses[processKey]; running {
		return fmt.Errorf("circular process execution detected: %s is already running", entryPointPath)
	}
	ctx.activeProcesses[processKey] = struct{}{}
	defer delete(ctx.activeProcesses, processKey)

	result, err := ctx.compileTypescriptForRunner(entryPointPath)
	if err != nil {
		return err
	}

	return ctx.executeRunnerBundle(result.OutputFiles)
}

func (ctx *Context) RunProcess(process string) error {
	return ctx.Run(process)
}

func (ctx *Context) validateWorkspaceSupportFiles() error {
	if ctx == nil || ctx.config == nil {
		return fmt.Errorf("context config is required")
	}

	requiredPaths := []string{
		ctx.config.CCodePath,
		filepath.Join(ctx.config.CCodePath, "tsconfig.json"),
	}
	supportFilePaths, err := templateassets.SupportFilePaths()
	if err != nil {
		return fmt.Errorf("list required support files: %w", err)
	}
	for _, path := range supportFilePaths {
		requiredPaths = append(
			requiredPaths,
			filepath.Join(ctx.config.HiddenPath, "lib", filepath.FromSlash(path)),
		)
	}

	missingPaths := []string{}
	for _, path := range requiredPaths {
		if !fileExists(path) {
			missingPaths = append(missingPaths, path)
		}
	}
	if len(missingPaths) == 0 {
		return nil
	}

	return fmt.Errorf("ccode workspace is not initialized or support files are missing. Run: ccode init. Missing: %s", strings.Join(missingPaths, ", "))
}

func (ctx *Context) resolveProcessEntryPoint(process string) (string, string, error) {
	if isStringBlank(process) {
		return "", "", fmt.Errorf("process name cannot be blank")
	}

	cleaned := filepath.Clean(filepath.FromSlash(process))
	if filepath.IsAbs(cleaned) || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("invalid process path %q", process)
	}

	entryPointPath := cleaned
	if strings.HasSuffix(entryPointPath, ".ts") && !strings.HasSuffix(entryPointPath, ProcessFileSuffix) {
		return "", "", fmt.Errorf("invalid process path %q: process files must end in %s or omit the suffix", process, ProcessFileSuffix)
	}
	if !strings.HasSuffix(entryPointPath, ProcessFileSuffix) {
		entryPointPath += ProcessFileSuffix
	}

	sourcePath := filepath.Join(ctx.config.CCodePath, entryPointPath)
	if !fileExists(sourcePath) {
		return "", "", fmt.Errorf("process %q not found: expected %s", process, sourcePath)
	}

	return entryPointPath, sourcePath, nil
}

func (ctx *Context) executeRunnerBundle(outputFiles []api.OutputFile) error {
	if len(outputFiles) == 0 {
		return fmt.Errorf("runner build produced no output files")
	}

	runtime := goja.New()
	if err := installRunnerNativeUtilities(runtime); err != nil {
		return err
	}

	module := runtime.NewObject()
	exports := runtime.NewObject()
	if err := module.Set("exports", exports); err != nil {
		return fmt.Errorf("prepare CommonJS module: %w", err)
	}
	runtime.Set("module", module)
	runtime.Set("exports", exports)

	program, err := goja.Compile(outputFiles[0].Path, string(outputFiles[0].Contents), false)
	if err != nil {
		return fmt.Errorf("compile runner bundle: %w", err)
	}

	if _, err := runtime.RunProgram(program); err != nil {
		return fmt.Errorf("execute runner bundle: %w", err)
	}

	exportedValue := module.Get("exports")
	exportedObject := exportedValue.ToObject(runtime)
	defaultExport := exportedObject.Get("default")
	defaultFunction, ok := goja.AssertFunction(defaultExport)
	if !ok {
		return fmt.Errorf("runner bundle must export a default function")
	}

	runnerContext := &RunnerContext{
		ccodeContext: ctx,
		runtime:      runtime,
		stdout:       ctx.stdout,
	}
	if err := runnerContext.initializeJSONParser(); err != nil {
		return err
	}
	jsContext, err := runnerContext.toValue(runtime)
	if err != nil {
		return err
	}

	if _, err := defaultFunction(goja.Undefined(), jsContext); err != nil {
		return fmt.Errorf("run default export: %w", err)
	}

	return nil
}

func (ctx *RunnerContext) initializeJSONParser() error {
	if ctx == nil || ctx.runtime == nil {
		return fmt.Errorf("runner context runtime is not initialized")
	}

	jsonObject := ctx.runtime.Get("JSON").ToObject(ctx.runtime)
	jsonParse, ok := goja.AssertFunction(jsonObject.Get("parse"))
	if !ok {
		return fmt.Errorf("JSON.parse is not available")
	}

	ctx.jsonObject = jsonObject
	ctx.jsonParse = jsonParse
	return nil
}

func (ctx *RunnerContext) toValue(runtime *goja.Runtime) (goja.Value, error) {
	object := runtime.NewObject()
	if err := object.Set("println", ctx.Println); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("runProcess", ctx.ccodeContext.RunProcess); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("renderTemplate", ctx.RenderTemplate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("generate", ctx.Generate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("accelerate", ctx.Accelerate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseJSONFromBytes", ctx.ParseJSONFromBytes); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseJSONFromString", ctx.ParseJSONFromString); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseJSONFromFile", ctx.ParseJSONFromFile); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseOpenAPIFromBytes", ctx.ParseOpenAPIFromBytes); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseOpenAPIFromString", ctx.ParseOpenAPIFromString); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("parseOpenAPIFromFile", ctx.ParseOpenAPIFromFile); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("loadOpenAPIFromFile", ctx.LoadOpenAPIFromFile); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("loadOpenAPIFromString", ctx.LoadOpenAPIFromString); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("loadOpenAPIFromBytes", ctx.LoadOpenAPIFromBytes); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("inspectDatabase", ctx.InspectDatabase); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	return object, nil
}

func (ctx *RunnerContext) Println(args ...any) {
	target := ctx.stdout
	if target == nil {
		target = os.Stdout
	}
	fmt.Fprintln(target, args...)
}

func (ctx *RunnerContext) RenderTemplate(templatePath string, data goja.Value) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return "", fmt.Errorf("runner context is not initialized")
	}

	templateData, err := gojaValueToTemplateData(ctx.runtime, data)
	if err != nil {
		return "", fmt.Errorf("convert template data: %w", err)
	}

	return ctx.renderTemplate(templatePath, templateData)
}
