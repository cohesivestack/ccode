package ccode

import "log"

type Context struct {
	config *Config
	logger *log.Logger
}

func NewContext(config *Config) *Context {
	return &Context{config: config}
}

type RunnerContext struct {
	context *Context
}
