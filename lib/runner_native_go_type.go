package ccode

import (
	"fmt"
	"slices"

	"github.com/dop251/goja"
)

func nativeGoTypeTransformation(runtime *goja.Runtime) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		input := nativeStringArgument(runtime, "go.toType", call.Argument(0))
		options, err := nativeGoTypeOptions(call.Argument(1))
		if err != nil {
			panic(runtime.NewTypeError("native go.toType %s", err))
		}
		result, err := stringToGoType(input, options)
		if err != nil {
			panic(runtime.NewTypeError("native go.toType %s", err))
		}
		return runtime.ToValue(result)
	}
}

func nativeGoTypeOptions(value goja.Value) (goTypeOptions, error) {
	if goja.IsUndefined(value) {
		return goTypeOptions{}, nil
	}
	object, ok := value.(*goja.Object)
	if !ok || object.ClassName() != "Object" {
		return goTypeOptions{}, fmt.Errorf("options must be an object")
	}

	values := make(map[string]any)
	for _, key := range object.Keys() {
		item := object.Get(key)
		if goja.IsUndefined(item) && slices.Contains(goTypeOptionNames, key) {
			continue
		}
		if key == "overrides" {
			overrides, ok := item.(*goja.Object)
			if !ok || overrides.ClassName() != "Object" {
				return goTypeOptions{}, fmt.Errorf("overrides must be a mapping of non-blank strings to non-blank strings")
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
	return parseGoTypeOptions(values)
}
