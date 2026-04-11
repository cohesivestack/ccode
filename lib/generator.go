package ccode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
)

func (ctx *RunnerContext) Generate(templatePath string, filePath string, data goja.Value) error {
	if ctx == nil || ctx.ccodeContext == nil {
		return fmt.Errorf("runner context is not initialized")
	}

	templateData, err := gojaValueToTemplateData(data)
	if err != nil {
		return fmt.Errorf("convert template data: %w", err)
	}

	if isStringBlank(filePath) {
		return fmt.Errorf("file path is required")
	}

	rendered, err := ctx.renderTemplate(templatePath, templateData)
	if err != nil {
		return err
	}

	outputFilePath := filePath
	if !filepath.IsAbs(outputFilePath) {
		outputFilePath = filepath.Join(ctx.ccodeContext.config.OutputPath, outputFilePath)
	}
	outputFilePath = filepath.Clean(outputFilePath)

	if err := os.MkdirAll(filepath.Dir(outputFilePath), 0755); err != nil {
		return fmt.Errorf("create output directory for %q: %w", outputFilePath, err)
	}

	if err := os.WriteFile(outputFilePath, []byte(rendered), 0644); err != nil {
		return fmt.Errorf("write rendered template to %q: %w", outputFilePath, err)
	}

	return nil
}
