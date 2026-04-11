package ccode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

func (ctx *RunnerContext) ParseJSONFromBytes(jsonBytes []byte) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	return ctx.parseJSON(string(jsonBytes))
}

func (ctx *RunnerContext) ParseJSONFromString(jsonString string) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	return ctx.parseJSON(jsonString)
}

func (ctx *RunnerContext) ParseJSONFromFile(filePath string) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	_filePath := filepath.Join(ctx.ccodeContext.config.CCodePath, filePath)

	if !fileExists(_filePath) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	data, err := os.ReadFile(_filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}

	return ctx.parseJSON(string(data))
}

func (ctx *RunnerContext) parseJSON(input string) (goja.Value, error) {
	if ctx == nil || ctx.runtime == nil || ctx.jsonParse == nil || ctx.jsonObject == nil {
		return nil, fmt.Errorf("runner context JSON parser is not initialized")
	}

	value, err := ctx.jsonParse(ctx.jsonObject, ctx.runtime.ToValue(input))
	if err != nil {
		return nil, fmt.Errorf("parse JSON: %w", err)
	}

	object, ok := value.(*goja.Object)
	if !ok || goja.IsNull(value) || object.ClassName() != "Object" {
		return nil, fmt.Errorf("JSON root must be an object")
	}

	return value, nil
}
