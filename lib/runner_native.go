package ccode

import (
	"fmt"
	"strconv"

	"github.com/dop251/goja"
)

type runnerNativeFunction struct {
	name      string
	transform func(goja.FunctionCall) goja.Value
}

func installRunnerNativeUtilities(runtime *goja.Runtime) error {
	if runtime == nil {
		return fmt.Errorf("runner runtime is required")
	}

	stringNamespace := runtime.NewObject()
	stringFunctions := []runnerNativeFunction{
		{name: "camelCase", transform: nativeInitialismTransformation(runtime, "string.camelCase", stringToCamelCase)},
		{name: "pascalCase", transform: nativeInitialismTransformation(runtime, "string.pascalCase", stringToPascalCase)},
		{name: "snakeCase", transform: nativeStringTransformation(runtime, "string.snakeCase", stringToSnakeCase)},
		{name: "kebabCase", transform: nativeStringTransformation(runtime, "string.kebabCase", stringToKebabCase)},
		{name: "constantCase", transform: nativeStringTransformation(runtime, "string.constantCase", stringToConstantCase)},
		{name: "dotCase", transform: nativeStringTransformation(runtime, "string.dotCase", stringToDotCase)},
		{name: "pathCase", transform: nativeStringTransformation(runtime, "string.pathCase", stringToPathCase)},
		{name: "titleCase", transform: nativeInitialismTransformation(runtime, "string.titleCase", stringToTitleCase)},
		{name: "sentenceCase", transform: nativeInitialismTransformation(runtime, "string.sentenceCase", stringToSentenceCase)},
		{name: "upperFirst", transform: nativeStringTransformation(runtime, "string.upperFirst", stringToUpperFirst)},
		{name: "lowerFirst", transform: nativeStringTransformation(runtime, "string.lowerFirst", stringToLowerFirst)},
		{name: "normalizeSpace", transform: nativeStringTransformation(runtime, "string.normalizeSpace", stringToNormalizeSpace)},
	}
	for _, function := range stringFunctions {
		if err := stringNamespace.Set(function.name, function.transform); err != nil {
			return fmt.Errorf("register native string.%s: %w", function.name, err)
		}
	}

	goNamespace := runtime.NewObject()
	goFunctions := []runnerNativeFunction{
		{name: "toExportedIdentifier", transform: nativeInitialismTransformation(runtime, "go.toExportedIdentifier", stringToGoExported)},
		{name: "toUnexportedIdentifier", transform: nativeInitialismTransformation(runtime, "go.toUnexportedIdentifier", stringToGoUnexported)},
		{name: "toPackageName", transform: nativeStringTransformation(runtime, "go.toPackageName", stringToGoPackage)},
	}
	for _, function := range goFunctions {
		if err := goNamespace.Set(function.name, function.transform); err != nil {
			return fmt.Errorf("register native go.%s: %w", function.name, err)
		}
	}

	native := runtime.NewObject()
	if err := native.Set("string", stringNamespace); err != nil {
		return fmt.Errorf("register native string namespace: %w", err)
	}
	if err := native.Set("go", goNamespace); err != nil {
		return fmt.Errorf("register native go namespace: %w", err)
	}

	if err := runtime.GlobalObject().DefineDataProperty(
		"__ccodeNative",
		native,
		goja.FLAG_FALSE,
		goja.FLAG_FALSE,
		goja.FLAG_FALSE,
	); err != nil {
		return fmt.Errorf("install native runner utilities: %w", err)
	}

	return nil
}

func nativeStringTransformation(
	runtime *goja.Runtime,
	name string,
	transform func(string) string,
) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		input := nativeStringArgument(runtime, name, call.Argument(0))
		return runtime.ToValue(transform(input))
	}
}

func nativeInitialismTransformation(
	runtime *goja.Runtime,
	name string,
	transform func(string, []string) string,
) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		input := nativeStringArgument(runtime, name, call.Argument(0))

		var initialisms []string
		if len(call.Arguments) >= 2 {
			var err error
			initialisms, err = nativeInitialisms(call.Arguments[1])
			if err != nil {
				panic(runtime.NewTypeError("native %s %s", name, err))
			}
		}

		return runtime.ToValue(transform(input, initialisms))
	}
}

func nativeStringArgument(runtime *goja.Runtime, name string, value goja.Value) string {
	if !goja.IsString(value) {
		panic(runtime.NewTypeError("native %s requires a string value", name))
	}

	return value.String()
}

func nativeInitialisms(value goja.Value) ([]string, error) {
	array, ok := value.(*goja.Object)
	if !ok || array.ClassName() != "Array" {
		return nil, fmt.Errorf("initialisms must be an array of strings")
	}

	length := int(array.Get("length").ToInteger())
	values := make([]initialismValue, 0, length)
	for index := range length {
		item := array.Get(strconv.Itoa(index))
		if !goja.IsString(item) {
			values = append(values, initialismValue{})
		} else {
			values = append(values, initialismValue{text: item.String(), isString: true})
		}
	}

	return parseInitialisms(values)
}
