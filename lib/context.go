package ccode

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

type Context struct {
	config *Config
	logger *log.Logger
	stdout io.Writer
}

func NewContext(config *Config) *Context {
	resolvedConfig := &Config{}
	if config != nil {
		*resolvedConfig = *config
	}

	if !filepath.IsAbs(resolvedConfig.HiddenPath) && !isStringBlank(resolvedConfig.HiddenPath) {
		resolvedConfig.HiddenPath = filepath.Join(resolvedConfig.Path, resolvedConfig.HiddenPath)
	}

	return &Context{
		config: resolvedConfig,
		stdout: os.Stdout,
	}
}
