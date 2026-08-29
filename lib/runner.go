package ccode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	templateassets "github.com/cohesivestack/ccode/template"
	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

var runnerEntryPointPattern = regexp.MustCompile(`(?s)export\s+default\s+function(?:\s+\w+)?\s*\(\s*\w+\s*:\s*Context\s*\)`)

type RunnerContext struct {
	ccodeContext             *Context
	runtime                  *goja.Runtime
	jsonObject               *goja.Object
	jsonParse                goja.Callable
	stdout                   io.Writer
	scopeName                string
	trackedAcceleratorStates map[string]struct{}
	trackedAcceleratorScopes map[string]struct{}
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

	if err := validateRunnerSource(sourcePath); err != nil {
		return err
	}

	result, err := ctx.compileTypescriptForRunner(entryPointPath)
	if err != nil {
		return err
	}

	defaultScope := runnerScopeFromEntryPoint(entryPointPath)
	return ctx.executeRunnerBundle(result.OutputFiles, defaultScope)
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
	if filepath.Ext(entryPointPath) != ".ts" {
		entryPointPath += ".ts"
	}

	sourcePath := filepath.Join(ctx.config.CCodePath, entryPointPath)
	if !fileExists(sourcePath) {
		return "", "", fmt.Errorf("process %q not found: expected %s", process, sourcePath)
	}

	return entryPointPath, sourcePath, nil
}

func validateRunnerSource(sourcePath string) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read process source %s: %w", sourcePath, err)
	}

	if !runnerEntryPointPattern.Match(source) {
		return fmt.Errorf("process %s must export a default function with a single Context-typed parameter", sourcePath)
	}

	return nil
}

func (ctx *Context) executeRunnerBundle(outputFiles []api.OutputFile, defaultScope string) error {
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
		ccodeContext:             ctx,
		runtime:                  runtime,
		stdout:                   ctx.stdout,
		scopeName:                defaultScope,
		trackedAcceleratorStates: map[string]struct{}{},
		trackedAcceleratorScopes: map[string]struct{}{},
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

	return runnerContext.cleanupUntrackedAcceleratorStates()
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
	if err := object.Set("renderTemplate", ctx.RenderTemplate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("generate", ctx.Generate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("accelerate", ctx.Accelerate); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("setScope", ctx.SetScope); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("scope", ctx.Scope); err != nil {
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

	templateData, err := gojaValueToTemplateData(data)
	if err != nil {
		return "", fmt.Errorf("convert template data: %w", err)
	}

	return ctx.renderTemplate(templatePath, templateData)
}

func runnerScopeFromEntryPoint(entryPointPath string) string {
	fileName := filepath.Base(filepath.Clean(entryPointPath))
	extension := filepath.Ext(fileName)
	scopeName := strings.TrimSuffix(fileName, extension)
	if isStringBlank(scopeName) {
		return "default"
	}
	return scopeName
}

func (ctx *RunnerContext) SetScope(scopeName string) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}
	normalizedScopeName, err := normalizeAcceleratorPathPart(scopeName, "scope")
	if err != nil {
		return err
	}

	ctx.scopeName = normalizedScopeName
	return nil
}

func (ctx *RunnerContext) Scope() string {
	if ctx == nil || isStringBlank(ctx.scopeName) {
		return "default"
	}
	return ctx.scopeName
}
