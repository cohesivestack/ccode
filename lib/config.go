package ccode

import (
	"fmt"
	"os"
	"path/filepath"

	v "github.com/cohesivestack/valgo"
	"gopkg.in/yaml.v3"
)

// Config represents the root cohesive code configuration.
type Config struct {
	CCodePath  string `yaml:"ccode_path"`
	OutputPath string `yaml:"output_path"`
	HiddenPath string `yaml:"hidden_path"`
	Version    string `yaml:"version"`
}

func (c *Config) validate() error {
	val := v.Is(
		v.String(c.CCodePath, "ccode_path").Not().Blank(),
		v.String(c.OutputPath, "output_path").Not().Blank(),
		v.String(c.HiddenPath, "hidden_path").Not().Blank(),
	)

	if val.Valid() {
		return nil
	}
	return val.ToValgoError()
}

// LoadConfig loads a YAML config file and returns a validated Config with defaults.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	if isStringBlank(config.Version) {
		return nil, fmt.Errorf("config validation failed: version is required")
	}

	config, err = NewConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}
	config.resolveConfigRelativePaths(filepath.Dir(filename))

	return config, nil
}

// NewConfig applies defaults and validates a config.
func NewConfig(config *Config) (*Config, error) {
	if config == nil {
		config = &Config{}
	}

	config.setDefaults()

	if err := config.validate(); err != nil {
		switch err.(type) {
		case *v.Error:
			out, _ := err.(*v.Error).MarshalJSONPretty()
			return nil, fmt.Errorf("config validation failed: %s", string(out))
		default:
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return config, nil
}

func (config *Config) setDefaults() {
	if isStringBlank(config.CCodePath) {
		config.CCodePath = "ccode"
	}
	if isStringBlank(config.OutputPath) {
		config.OutputPath = "."
	}
	if isStringBlank(config.HiddenPath) {
		config.HiddenPath = DefaultHiddenFolderName
	}
}

func (config *Config) resolveConfigRelativePaths(baseDir string) {
	if isStringBlank(baseDir) {
		return
	}
	if !filepath.IsAbs(config.CCodePath) && !isStringBlank(config.CCodePath) {
		config.CCodePath = filepath.Clean(filepath.Join(baseDir, config.CCodePath))
	}
}

func (config *Config) UnmarshalYAML(node *yaml.Node) error {
	type configAlias struct {
		CCodePath  string `yaml:"ccode_path"`
		LegacyPath string `yaml:"path"`
		OutputPath string `yaml:"output_path"`
		HiddenPath string `yaml:"hidden_path"`
		Version    string `yaml:"version"`
	}

	var raw configAlias
	if err := node.Decode(&raw); err != nil {
		return err
	}

	config.CCodePath = raw.CCodePath
	if isStringBlank(config.CCodePath) {
		config.CCodePath = raw.LegacyPath
	}
	config.OutputPath = raw.OutputPath
	config.HiddenPath = raw.HiddenPath
	config.Version = raw.Version

	return nil
}
