package ccode

import (
	"fmt"
	"path/filepath"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"
)

func (ctx *RunnerContext) renderTemplate(templatePath string, data any) (string, error) {
	if ctx == nil || ctx.ccodeContext == nil || ctx.ccodeContext.config == nil {
		return "", fmt.Errorf("context config is required")
	}
	if isStringBlank(templatePath) {
		return "", fmt.Errorf("template path is required")
	}

	loader, err := loaders.NewFileSystemLoader(ctx.ccodeContext.config.CCodePath)
	if err != nil {
		return "", fmt.Errorf("create template loader: %w", err)
	}

	templateContext := exec.NewContext(map[string]any{
		"data": data,
	})

	environment, err := cohesiveTemplateEnvironment()
	if err != nil {
		return "", fmt.Errorf("create template environment: %w", err)
	}

	tmpl, err := exec.NewTemplate(filepath.Clean(templatePath), gonja.NewConfig(), loader, environment)
	if err != nil {
		return "", fmt.Errorf("parse template %q: %w", templatePath, err)
	}

	result, err := tmpl.ExecuteToString(templateContext)
	if err != nil {
		return "", fmt.Errorf("render template %q: %w", templatePath, err)
	}

	return result, nil
}
