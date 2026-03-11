package ccode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func (ctx *Context) ParseJSONFromBytes(jsonBytes []byte) (result map[string]any, err error) {

	err = json.Unmarshal(jsonBytes, &result)

	return result, err
}

func (ctx *Context) ParseJSONFromString(jsonString string) (result map[string]any, err error) {

	return ctx.ParseJSONFromBytes([]byte(jsonString))
}

func (ctx *Context) ParseJSONFromFile(filePath string) (result map[string]any, err error) {

	_filePath := filepath.Join(ctx.config.CCodePath, filePath)

	if !fileExists(_filePath) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	data, err := os.ReadFile(_filePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}

	return ctx.ParseJSONFromBytes(data)
}

func (ctx *RunnerContext) ParseJSONFromBytes(jsonBytes []byte) (map[string]any, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	return ctx.ccodeContext.ParseJSONFromBytes(jsonBytes)
}

func (ctx *RunnerContext) ParseJSONFromString(jsonString string) (map[string]any, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	return ctx.ccodeContext.ParseJSONFromString(jsonString)
}

func (ctx *RunnerContext) ParseJSONFromFile(filePath string) (map[string]any, error) {
	if ctx == nil || ctx.ccodeContext == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}
	return ctx.ccodeContext.ParseJSONFromFile(filePath)
}
