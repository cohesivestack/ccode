package ccode

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/dop251/goja"
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

func gojaValueToTemplateData(value goja.Value) (any, error) {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil, nil
	}

	object, ok := value.(*goja.Object)
	if !ok {
		return value.Export(), nil
	}

	switch object.ClassName() {
	case "Array":
		return gojaArrayToTemplateSlice(object)
	case "Object":
		return gojaObjectToTemplateDict(object)
	case "String", "Number", "Boolean", "Date":
		return object.Export(), nil
	default:
		return object.Export(), nil
	}
}

func gojaObjectToTemplateDict(object *goja.Object) (exec.Dict, error) {
	keys := object.Keys()
	dict := exec.Dict{
		Pairs: make([]*exec.Pair, 0, len(keys)),
	}

	for _, key := range keys {
		item, err := gojaValueToTemplateData(object.Get(key))
		if err != nil {
			return exec.Dict{}, fmt.Errorf("convert object key %q: %w", key, err)
		}

		dict.Pairs = append(dict.Pairs, &exec.Pair{
			Key:   exec.AsValue(key),
			Value: exec.AsValue(item),
		})
	}

	return dict, nil
}

func gojaArrayToTemplateSlice(object *goja.Object) ([]any, error) {
	length := object.Get("length").ToInteger()
	items := make([]any, 0, length)

	for i := int64(0); i < length; i++ {
		item, err := gojaValueToTemplateData(object.Get(strconv.FormatInt(i, 10)))
		if err != nil {
			return nil, fmt.Errorf("convert array index %d: %w", i, err)
		}
		items = append(items, item)
	}

	return items, nil
}
