package ccode

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	openapiutils "github.com/pb33f/libopenapi/utils"
)

func (ctx *RunnerContext) ParseOpenAPIFromBytes(specBytes []byte) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.parseOpenAPIDocument(specBytes, datamodel.NewDocumentConfiguration(), "bytes", "")
}

func (ctx *RunnerContext) ParseOpenAPIFromString(spec string) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	return ctx.parseOpenAPIDocument([]byte(spec), datamodel.NewDocumentConfiguration(), "string", "")
}

func (ctx *RunnerContext) ParseOpenAPIFromFile(filePath string, optionValues ...goja.Value) (goja.Value, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.runtime == nil {
		return nil, fmt.Errorf("runner context is not initialized")
	}

	expectedVersion, err := ctx.parseOpenAPIFileOptions(optionValues)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(ctx.ccodeContext.config.CCodePath, filePath)
	if !fileExists(fullPath) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	specBytes, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read OpenAPI file %s: %w", filePath, err)
	}

	config := datamodel.NewDocumentConfiguration()
	config.BasePath = filepath.Dir(fullPath)
	config.SpecFilePath = filepath.Base(fullPath)
	config.AllowFileReferences = true

	return ctx.parseOpenAPIDocument(specBytes, config, filePath, expectedVersion)
}

func (ctx *RunnerContext) parseOpenAPIFileOptions(optionValues []goja.Value) (string, error) {
	if len(optionValues) == 0 {
		return "", nil
	}
	if len(optionValues) > 1 {
		return "", fmt.Errorf("parse OpenAPI file accepts at most one options argument")
	}

	optionsValue := optionValues[0]
	if optionsValue == nil || goja.IsUndefined(optionsValue) || goja.IsNull(optionsValue) {
		return "", fmt.Errorf("parse OpenAPI file options must be an object")
	}

	expectedVersionValue := optionsValue.ToObject(ctx.runtime).Get("expectedVersion")
	if expectedVersionValue == nil || goja.IsUndefined(expectedVersionValue) || goja.IsNull(expectedVersionValue) {
		return "", fmt.Errorf("parse OpenAPI file options must include expectedVersion")
	}

	expectedVersion, ok := expectedVersionValue.Export().(string)
	if !ok {
		return "", fmt.Errorf("parse OpenAPI file expectedVersion must be a string")
	}

	switch expectedVersion {
	case "3.0", "3.1":
		return expectedVersion, nil
	default:
		return "", fmt.Errorf("unsupported expected OpenAPI version %q", expectedVersion)
	}
}

func (ctx *RunnerContext) parseOpenAPIDocument(specBytes []byte, config *datamodel.DocumentConfiguration, source string, expectedVersion string) (goja.Value, error) {
	document, err := libopenapi.NewDocumentWithConfiguration(specBytes, config)
	if err != nil {
		return nil, fmt.Errorf("parse OpenAPI document %s: %w", source, err)
	}
	defer document.Release()

	specInfo := document.GetSpecInfo()
	if specInfo == nil {
		return nil, fmt.Errorf("parse OpenAPI document %s: missing spec info", source)
	}

	switch specInfo.SpecType {
	case openapiutils.OpenApi2:
		return nil, fmt.Errorf("swagger is not supported")
	case openapiutils.OpenApi3:
		// supported
	default:
		return nil, fmt.Errorf("unsupported OpenAPI spec type %q", specInfo.SpecType)
	}

	if expectedVersion != "" {
		actualVersion := openAPIMajorMinorVersion(specInfo.Version)
		if actualVersion != expectedVersion {
			return nil, fmt.Errorf("expected OpenAPI %s.x, but %s declares %s", expectedVersion, source, specInfo.Version)
		}
	}

	docModel, err := document.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build OpenAPI v3 model %s: %w", source, err)
	}

	renderedJSON, err := docModel.Model.RenderJSON("  ")
	if err != nil {
		return nil, fmt.Errorf("render OpenAPI v3 model %s as JSON: %w", source, err)
	}

	return ctx.parseJSON(string(renderedJSON))
}

func openAPIMajorMinorVersion(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return version
	}
	return parts[0] + "." + parts[1]
}
