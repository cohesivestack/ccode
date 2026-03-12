package ccode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
)

var runnerEntryPointPattern = regexp.MustCompile(`(?s)export\s+default\s+function(?:\s+\w+)?\s*\(\s*\w+\s*:\s*Context\s*\)`)

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

	return ctx.executeRunnerBundle(result.OutputFiles)
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

func (ctx *Context) executeRunnerBundle(outputFiles []api.OutputFile) error {
	if len(outputFiles) == 0 {
		return fmt.Errorf("runner build produced no output files")
	}

	runtime := goja.New()
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
	if err := object.Set("templateToString", ctx.TemplateToString); err != nil {
		return nil, fmt.Errorf("set runner context functions: %w", err)
	}
	if err := object.Set("templateToFile", ctx.TemplateToFile); err != nil {
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
	return object, nil
}

func (ctx *RunnerContext) Println(args ...any) {
	target := ctx.stdout
	if target == nil {
		target = os.Stdout
	}
	fmt.Fprintln(target, args...)
}

func (ctx *RunnerContext) TemplateToString(templatePath string, data map[string]any) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return "", fmt.Errorf("runner context is not initialized")
	}
	return ctx.ccodeContext.TemplateToString(templatePath, data)
}

func (ctx *RunnerContext) TemplateToFile(templatePath string, filePath string, data map[string]any) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}
	return ctx.ccodeContext.TemplateToFile(templatePath, filePath, data)
}
