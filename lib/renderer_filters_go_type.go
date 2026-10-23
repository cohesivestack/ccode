package ccode

import (
	"fmt"

	"github.com/nikolalohinski/gonja/v2/exec"
)

func goTypeTemplateFilter(_ *exec.Evaluator, in *exec.Value, params *exec.VarArgs) *exec.Value {
	if err := validateTemplateFilterArguments(params, typeMappingOptionNames...); err != nil {
		return templateFilterError("goType", err)
	}
	input, err := templateFilterString(in)
	if err != nil {
		return templateFilterError("goType", err)
	}
	values := make(map[string]any)
	if params != nil {
		for _, key := range typeMappingOptionNames {
			value, supplied := params.KwArgs[key]
			if !supplied {
				continue
			}
			if key == "overrides" {
				entries, err := templateGoTypeOverrides(value)
				if err != nil {
					return templateFilterError("goType", err)
				}
				values[key] = entries
			} else if value == nil || value.IsNil() {
				values[key] = nil
			} else if value.IsString() {
				values[key] = value.String()
			} else if value.IsBool() {
				values[key] = value.Bool()
			} else {
				values[key] = value.Interface()
			}
		}
	}
	options, err := parseTypeMappingOptions(values)
	if err != nil {
		return templateFilterError("goType", err)
	}
	result, err := stringToGoType(input, options)
	if err != nil {
		return templateFilterError("goType", err)
	}
	return exec.AsValue(result)
}

func templateGoTypeOverrides(value *exec.Value) (map[string]any, error) {
	if value == nil || !value.IsDict() {
		return nil, fmt.Errorf("overrides must be a mapping of non-blank strings to non-blank strings")
	}
	entries := make(map[string]any)
	var err error
	value.Iterate(func(_ int, _ int, key, item *exec.Value) bool {
		if key == nil || !key.IsString() || item == nil || !item.IsString() {
			err = fmt.Errorf("overrides must be a mapping of non-blank strings to non-blank strings")
			return false
		}
		entries[key.String()] = item.String()
		return true
	}, func() {})
	return entries, err
}
