package ccode

import (
	"fmt"
	"os"

	v "github.com/cohesivestack/valgo"
	"gopkg.in/yaml.v3"
)

// Config represents the root cohesive code configuration.
type Config struct {
	Path       string `yaml:"path"`
	OutputPath string `yaml:"output_path"`
	HiddenPath string `yaml:"hidden_path"`
}

func (c *Config) validate() error {
	val := v.Is(
		v.String(c.Path, "path").Not().Blank(),
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

	config, err = NewConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

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
	if isStringBlank(config.Path) {
		config.Path = "ccode"
	}
	if isStringBlank(config.OutputPath) {
		config.OutputPath = "."
	}
	if isStringBlank(config.HiddenPath) {
		config.HiddenPath = DefaultHiddenFolderName
	}
}
