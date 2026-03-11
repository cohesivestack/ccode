package ccode

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"
)

func (ctx *Context) TemplateToString(templatePath string, data map[string]any) (result string, err error) {
	result, err = ctx.renderTemplate(templatePath, data)
	if err != nil {
		return "", err
	}

	return
}

func (ctx *Context) TemplateToFile(templatePath, filePath string, data map[string]any) (err error) {
	if isStringBlank(filePath) {
		return fmt.Errorf("file path is required")
	}

	rendered, err := ctx.renderTemplate(templatePath, data)
	if err != nil {
		return err
	}

	outputFilePath := filePath
	if !filepath.IsAbs(outputFilePath) {
		outputFilePath = filepath.Join(ctx.config.OutputPath, outputFilePath)
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

func (ctx *Context) renderTemplate(templatePath string, data map[string]any) (result string, err error) {
	if ctx == nil || ctx.config == nil {
		return "", fmt.Errorf("context config is required")
	}
	if isStringBlank(templatePath) {
		return "", fmt.Errorf("template path is required")
	}

	loader, err := loaders.NewFileSystemLoader(ctx.config.CCodePath)
	if err != nil {
		return "", fmt.Errorf("create template loader: %w", err)
	}

	templateContext := exec.NewContext(map[string]any{
		"data": data,
	})

	tmpl, err := exec.NewTemplate(filepath.Clean(templatePath), gonja.NewConfig(), loader, gonja.DefaultEnvironment)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", templatePath, err)
	}

	result, err = tmpl.ExecuteToString(templateContext)
	if err != nil {
		return "", fmt.Errorf("render template %q: %w", templatePath, err)
	}

	return
}
