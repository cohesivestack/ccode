package ccode

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	texttemplate "text/template"

	templateassets "github.com/cohesivestack/ccode/template"
	v "github.com/cohesivestack/valgo"
)

type initOptions struct {
	ProjectPath string
	ConfigPath  string
}

func (o *initOptions) setDefaults() {
	if isStringBlank(o.ProjectPath) {
		o.ProjectPath = "ccode"
	} else {
		o.ProjectPath = filepath.Clean(o.ProjectPath)
	}

	if isStringBlank(o.ConfigPath) {
		o.ConfigPath = DefaultConfigFileName
	} else {
		o.ConfigPath = filepath.Clean(o.ConfigPath)
	}
}

func (o *initOptions) validate() error {
	val := v.Is(
		v.String(o.ProjectPath, "project_path").Not().Blank(),
		v.String(o.ConfigPath, "config_path").Not().Blank(),
	)
	if val.Valid() {
		return nil
	}
	return val.ToValgoError()
}

func newInitOptions(projectPath string, configPath string) (*initOptions, error) {
	options := &initOptions{
		ProjectPath: projectPath,
		ConfigPath:  configPath,
	}
	options.setDefaults()

	if err := options.validate(); err != nil {
		switch typed := err.(type) {
		case *v.Error:
			out, _ := typed.MarshalJSONPretty()
			return nil, fmt.Errorf("init validation failed: %s", string(out))
		default:
			return nil, fmt.Errorf("init validation failed: %w", err)
		}
	}

	return options, nil
}

func Init(projectPath string, configPath string) error {
	options, err := newInitOptions(projectPath, configPath)
	if err != nil {
		return err
	}

	hiddenPath := filepath.Join(options.ProjectPath, DefaultHiddenFolderName)
	hiddenLibPath := filepath.Join(hiddenPath, "lib")
	buildPath := filepath.Join(hiddenPath, "build")

	for _, path := range []string{options.ProjectPath, hiddenLibPath, buildPath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
	}

	if err := writeConfigIfMissing(options.ConfigPath, options.ProjectPath); err != nil {
		return err
	}
	if err := writeFileIfMissing(filepath.Join(hiddenLibPath, "context.ts"), templateassets.ContextTemplate, "context template"); err != nil {
		return err
	}
	if err := writeFileIfMissing(filepath.Join(options.ProjectPath, "tsconfig.json"), templateassets.TSConfigTemplate, "tsconfig"); err != nil {
		return err
	}

	return nil
}

func writeConfigIfMissing(configPath string, projectPath string) error {
	rendered, err := renderConfigTemplate(projectPath)
	if err != nil {
		return err
	}

	return writeFileIfMissing(configPath, rendered, "config file")
}

func renderConfigTemplate(projectPath string) (string, error) {
	tmpl, err := texttemplate.New(DefaultConfigFileName).Parse(templateassets.ConfigTemplate)
	if err != nil {
		return "", fmt.Errorf("parse config template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, struct {
		ProjectPath string
	}{
		ProjectPath: projectPath,
	}); err != nil {
		return "", fmt.Errorf("render config template: %w", err)
	}

	return buffer.String(), nil
}

func writeFileIfMissing(path string, content string, assetName string) error {
	if fileExists(path) {
		slog.Warn(fmt.Sprintf("%s already exists; not overwriting", assetName), "path", path)
		return nil
	}

	parentDir := filepath.Dir(path)
	if parentDir != "." {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("create parent directory %s: %w", parentDir, err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s at %s: %w", assetName, path, err)
	}

	return nil
}
