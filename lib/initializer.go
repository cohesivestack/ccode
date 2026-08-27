package ccode

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	texttemplate "text/template"

	templateassets "github.com/cohesivestack/ccode/template"
	v "github.com/cohesivestack/valgo"
	"gopkg.in/yaml.v3"
)

type initOptions struct {
	ProjectPath string
	ConfigPath  string
	Version     string
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
		v.String(o.Version, "version").Not().Blank(),
	)
	if val.Valid() {
		return nil
	}
	return val.ToValgoError()
}

func newInitOptions(projectPath string, configPath string, version string) (*initOptions, error) {
	options := &initOptions{
		ProjectPath: projectPath,
		ConfigPath:  configPath,
		Version:     version,
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

func Init(projectPath string, configPath string, version string) error {
	options, err := newInitOptions(projectPath, configPath, version)
	if err != nil {
		return err
	}

	workspacePath, hiddenPath, err := resolveInitWorkspacePaths(options)
	if err != nil {
		return err
	}
	hiddenLibPath := filepath.Join(hiddenPath, "lib")
	buildPath := filepath.Join(hiddenPath, "build")

	for _, path := range []string{workspacePath, hiddenLibPath} {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
	}
	if err := os.RemoveAll(buildPath); err != nil {
		return fmt.Errorf("clear build cache %s: %w", buildPath, err)
	}
	if err := os.MkdirAll(buildPath, 0755); err != nil {
		return fmt.Errorf("create %s: %w", buildPath, err)
	}

	if err := removeObsoleteSupportFiles(hiddenLibPath); err != nil {
		return err
	}
	if err := installSupportFiles(hiddenLibPath); err != nil {
		return err
	}

	configContents, err := renderConfigTemplate(options.ProjectPath, options.Version)
	if err != nil {
		return err
	}
	for _, asset := range []struct {
		path      string
		contents  string
		name      string
		overwrite bool
	}{
		{path: options.ConfigPath, contents: configContents, name: "config file"},
		{path: filepath.Join(hiddenPath, ".gitignore"), contents: templateassets.HiddenGitIgnoreTemplate, name: ".ccode gitignore"},
		{path: filepath.Join(workspacePath, "tsconfig.json"), contents: templateassets.TSConfigTemplate, name: "tsconfig"},
	} {
		writeAsset := writeFileIfMissing
		if asset.overwrite {
			writeAsset = writeFile
		}
		if err := writeAsset(asset.path, asset.contents, asset.name); err != nil {
			return err
		}
	}
	if err := updateConfigVersion(options.ConfigPath, options.Version); err != nil {
		return err
	}

	return nil
}

var obsoleteSupportFileNames = []string{
	"openapi.ts",
	"database.ts",
	"database-postgresql.ts",
	"database-mysql.ts",
	"database-mariadb.ts",
	"database-sqlite.ts",
}

func installSupportFiles(hiddenLibPath string) error {
	paths, err := templateassets.SupportFilePaths()
	if err != nil {
		return fmt.Errorf("list embedded support files: %w", err)
	}

	for _, path := range paths {
		contents, err := templateassets.SupportFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded support file %s: %w", path, err)
		}
		if err := writeFile(
			filepath.Join(hiddenLibPath, filepath.FromSlash(path)),
			string(contents),
			fmt.Sprintf("%s support file", path),
		); err != nil {
			return err
		}
	}

	return nil
}

func removeObsoleteSupportFiles(hiddenLibPath string) error {
	for _, name := range obsoleteSupportFileNames {
		path := filepath.Join(hiddenLibPath, name)
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect obsolete support file %s: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove obsolete support file %s: %w", path, err)
		}
	}
	return nil
}

func resolveInitWorkspacePaths(options *initOptions) (string, string, error) {
	workspacePath := options.ProjectPath
	hiddenPath := filepath.Join(workspacePath, DefaultHiddenFolderName)

	if !fileExists(options.ConfigPath) {
		return workspacePath, hiddenPath, nil
	}

	payload, err := os.ReadFile(options.ConfigPath)
	if err != nil {
		return "", "", fmt.Errorf("read config file %s: %w", options.ConfigPath, err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(payload, config); err != nil {
		return "", "", fmt.Errorf("parse config file %s: %w", options.ConfigPath, err)
	}
	if !isStringBlank(config.CCodePath) {
		workspacePath = config.CCodePath
	}
	if !filepath.IsAbs(workspacePath) {
		workspacePath = filepath.Clean(filepath.Join(filepath.Dir(options.ConfigPath), workspacePath))
	}

	hiddenPath = config.HiddenPath
	if isStringBlank(hiddenPath) {
		hiddenPath = DefaultHiddenFolderName
	}
	if !filepath.IsAbs(hiddenPath) {
		hiddenPath = filepath.Join(workspacePath, hiddenPath)
	}

	return workspacePath, hiddenPath, nil
}

func updateConfigVersion(configPath string, version string) error {
	payload, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read config file %s: %w", configPath, err)
	}

	content := string(payload)
	lines := strings.SplitAfter(content, "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(strings.TrimRight(line, "\n"))
		if !strings.HasPrefix(trimmed, "version:") {
			continue
		}

		prefixLength := strings.Index(line, "version:")
		prefix := line[:prefixLength]
		lineEnding := ""
		if strings.HasSuffix(line, "\n") {
			lineEnding = "\n"
		}
		comment := ""
		if commentIndex := strings.Index(line[prefixLength:], "#"); commentIndex >= 0 {
			comment = strings.TrimRight(line[prefixLength+commentIndex:], "\n")
			if !strings.HasPrefix(comment, " ") {
				comment = " " + comment
			}
		}
		lines[index] = fmt.Sprintf("%sversion: %s%s%s", prefix, version, comment, lineEnding)
		return os.WriteFile(configPath, []byte(strings.Join(lines, "")), 0644)
	}

	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += fmt.Sprintf("version: %s\n", version)

	return os.WriteFile(configPath, []byte(content), 0644)
}

func renderConfigTemplate(projectPath string, version string) (string, error) {
	tmpl, err := texttemplate.New(DefaultConfigFileName).Parse(templateassets.ConfigTemplate)
	if err != nil {
		return "", fmt.Errorf("parse config template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tmpl.Execute(&buffer, struct {
		ProjectPath string
		Version     string
	}{
		ProjectPath: projectPath,
		Version:     version,
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

	return writeFile(path, content, assetName)
}

func writeFile(path string, content string, assetName string) error {
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
