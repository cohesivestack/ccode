package ccode

import (
	"fmt"
	"slices"

	"github.com/dop251/goja"
)

func nativeTypeMappingTransformation(
	runtime *goja.Runtime,
	name string,
	transform func(string, typeMappingOptions) (string, error),
) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		input := nativeStringArgument(runtime, name, call.Argument(0))
		options, err := nativeTypeMappingOptions(call.Argument(1))
		if err != nil {
			panic(runtime.NewTypeError("native %s %s", name, err))
		}
		result, err := transform(input, options)
		if err != nil {
			panic(runtime.NewTypeError("native %s %s", name, err))
		}
		return runtime.ToValue(result)
	}
}

func nativeTypeMappingOptions(value goja.Value) (typeMappingOptions, error) {
	if goja.IsUndefined(value) {
		return typeMappingOptions{}, nil
	}
	object, ok := value.(*goja.Object)
	if !ok || object.ClassName() != "Object" {
		return typeMappingOptions{}, fmt.Errorf("options must be an object")
	}

	values := make(map[string]any)
	for _, key := range object.Keys() {
		item := object.Get(key)
		if goja.IsUndefined(item) && slices.Contains(typeMappingOptionNames, key) {
			continue
		}
		if key == "overrides" {
			overrides, ok := item.(*goja.Object)
			if !ok || overrides.ClassName() != "Object" {
				return typeMappingOptions{}, fmt.Errorf("overrides must be a mapping of non-blank strings to non-blank strings")
			}
			entries := make(map[string]any)
			for _, name := range overrides.Keys() {
				entries[name] = overrides.Get(name).Export()
			}
			values[key] = entries
		} else {
			values[key] = item.Export()
		}
	}
	return parseTypeMappingOptions(values)
}
