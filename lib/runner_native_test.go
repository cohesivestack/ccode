package ccode

import (
	"fmt"
	"sort"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallRunnerNativeUtilitiesRejectsNilRuntime(t *testing.T) {
	err := installRunnerNativeUtilities(nil)
	require.EqualError(t, err, "runner runtime is required")
}

func TestInstallRunnerNativeUtilitiesCreatesCompleteInternalAPI(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))

	nativeValue := runtime.Get("__ccodeNative")
	require.False(t, goja.IsUndefined(nativeValue))
	native := nativeValue.ToObject(runtime)
	rootKeys := native.Keys()
	sort.Strings(rootKeys)
	assert.Equal(t, []string{"go", "openapi", "string", "typescript"}, rootKeys)

	tests := []struct {
		name      string
		namespace string
		expected  []string
	}{
		{
			name:      "string namespace",
			namespace: "string",
			expected: []string{
				"camelCase",
				"pascalCase",
				"snakeCase",
				"kebabCase",
				"constantCase",
				"dotCase",
				"pathCase",
				"titleCase",
				"sentenceCase",
				"upperFirst",
				"lowerFirst",
				"normalizeSpace",
			},
		},
		{
			name:      "go namespace",
			namespace: "go",
			expected: []string{
				"toExportedIdentifier",
				"toUnexportedIdentifier",
				"toPackageName",
			},
		},
		{
			name:      "TypeScript namespace",
			namespace: "typescript",
			expected: []string{
				"toTypeIdentifier",
				"toValueIdentifier",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			namespaceValue := native.Get(test.namespace)
			require.False(t, goja.IsUndefined(namespaceValue))
			namespace := namespaceValue.ToObject(runtime)

			actual := namespace.Keys()
			sort.Strings(actual)
			expected := append([]string(nil), test.expected...)
			sort.Strings(expected)
			assert.Equal(t, expected, actual)

			for _, name := range test.expected {
				_, callable := goja.AssertFunction(namespace.Get(name))
				assert.True(t, callable, "%s.%s", test.namespace, name)
			}
		})
	}

	openAPIValue := native.Get("openapi")
	require.False(t, goja.IsUndefined(openAPIValue))
	openAPI := openAPIValue.ToObject(runtime)
	assert.Equal(t, []string{"path"}, openAPI.Keys())

	pathValue := openAPI.Get("path")
	require.False(t, goja.IsUndefined(pathValue))
	path := pathValue.ToObject(runtime)
	pathFunctions := []string{"toColon", "toSquareBrackets", "toAngleBrackets", "toDollar"}
	actualPathFunctions := path.Keys()
	sort.Strings(actualPathFunctions)
	sort.Strings(pathFunctions)
	assert.Equal(t, pathFunctions, actualPathFunctions)
	for _, name := range pathFunctions {
		_, callable := goja.AssertFunction(path.Get(name))
		assert.True(t, callable, "openapi.path.%s", name)
	}

	for _, name := range []string{"Go", "Strings", "TypeScript", "OpenAPI", "Path", "go", "string", "typescript", "openapi"} {
		value := runtime.Get(name)
		assert.True(t, value == nil || goja.IsUndefined(value), "global %s", name)
	}

	descriptor, err := runtime.RunString(`
        const descriptor = Object.getOwnPropertyDescriptor(globalThis, "__ccodeNative");
        JSON.stringify({
          enumerable: descriptor.enumerable,
          configurable: descriptor.configurable,
          writable: descriptor.writable,
        });
    `)
	require.NoError(t, err)
	assert.JSONEq(t, `{"enumerable":false,"configurable":false,"writable":false}`, descriptor.String())
}

func TestRunnerNativeOpenAPIPathUtilitiesDelegateToExistingTransformations(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))

	tests := []struct {
		name      string
		function  string
		transform func(string, bool) string
	}{
		{name: "colon", function: "toColon", transform: openAPIPathToColon},
		{name: "square brackets", function: "toSquareBrackets", transform: openAPIPathToSquareBrackets},
		{name: "angle brackets", function: "toAngleBrackets", transform: openAPIPathToAngleBrackets},
		{name: "dollar", function: "toDollar", transform: openAPIPathToDollar},
	}

	path := runtime.Get("__ccodeNative").ToObject(runtime).
		Get("openapi").ToObject(runtime).
		Get("path").ToObject(runtime)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			function, ok := goja.AssertFunction(path.Get(test.function))
			require.True(t, ok)

			input := "/users/{userId}/orders/{orderId}"
			preserved, err := function(goja.Undefined(), runtime.ToValue(input))
			require.NoError(t, err)
			assert.Equal(t, test.transform(input, false), preserved.String())

			omitted, err := function(goja.Undefined(), runtime.ToValue(input), runtime.ToValue(true))
			require.NoError(t, err)
			assert.Equal(t, test.transform(input, true), omitted.String())
		})
	}
}

func TestRunnerNativeUtilitiesDelegateToExistingTransformations(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))

	tests := []struct {
		name        string
		namespace   string
		function    string
		input       string
		initialisms []string
		expected    string
	}{
		{name: "camel case", namespace: "string", function: "camelCase", input: "api response", initialisms: []string{"API"}, expected: stringToCamelCase("api response", []string{"API"})},
		{name: "pascal case", namespace: "string", function: "pascalCase", input: "api response", initialisms: []string{"API"}, expected: stringToPascalCase("api response", []string{"API"})},
		{name: "snake case", namespace: "string", function: "snakeCase", input: "hello world", expected: stringToSnakeCase("hello world")},
		{name: "kebab case", namespace: "string", function: "kebabCase", input: "hello world", expected: stringToKebabCase("hello world")},
		{name: "constant case", namespace: "string", function: "constantCase", input: "hello world", expected: stringToConstantCase("hello world")},
		{name: "dot case", namespace: "string", function: "dotCase", input: "hello world", expected: stringToDotCase("hello world")},
		{name: "path case", namespace: "string", function: "pathCase", input: "hello world", expected: stringToPathCase("hello world")},
		{name: "title case", namespace: "string", function: "titleCase", input: "api response", initialisms: []string{"API"}, expected: stringToTitleCase("api response", []string{"API"})},
		{name: "sentence case", namespace: "string", function: "sentenceCase", input: "api response", initialisms: []string{"API"}, expected: stringToSentenceCase("api response", []string{"API"})},
		{name: "upper first", namespace: "string", function: "upperFirst", input: "hello world", expected: stringToUpperFirst("hello world")},
		{name: "lower first", namespace: "string", function: "lowerFirst", input: "Hello World", expected: stringToLowerFirst("Hello World")},
		{name: "normalize space", namespace: "string", function: "normalizeSpace", input: "  hello   world  ", expected: stringToNormalizeSpace("  hello   world  ")},
		{name: "Go exported identifier", namespace: "go", function: "toExportedIdentifier", input: "user id", initialisms: []string{"ID"}, expected: stringToGoExported("user id", []string{"ID"})},
		{name: "Go unexported identifier", namespace: "go", function: "toUnexportedIdentifier", input: "user id", initialisms: []string{"ID"}, expected: stringToGoUnexported("user id", []string{"ID"})},
		{name: "Go package", namespace: "go", function: "toPackageName", input: "HTTP Utils", expected: stringToGoPackage("HTTP Utils")},
		{name: "TypeScript type identifier", namespace: "typescript", function: "toTypeIdentifier", input: "api response id", initialisms: []string{"API", "ID"}, expected: stringToTypeScriptTypeIdentifier("api response id", []string{"API", "ID"})},
		{name: "TypeScript value identifier", namespace: "typescript", function: "toValueIdentifier", input: "api response id", initialisms: []string{"API", "ID"}, expected: stringToTypeScriptValueIdentifier("api response id", []string{"API", "ID"})},
	}

	native := runtime.Get("__ccodeNative").ToObject(runtime)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			namespace := native.Get(test.namespace).ToObject(runtime)
			function, ok := goja.AssertFunction(namespace.Get(test.function))
			require.True(t, ok)

			arguments := []goja.Value{runtime.ToValue(test.input)}
			if test.initialisms != nil {
				arguments = append(arguments, runtime.ToValue(test.initialisms))
			}
			actual, err := function(goja.Undefined(), arguments...)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual.String())
		})
	}
}

func TestRunnerNativeInitialismTransformationsPreserveOmittedDefaults(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))

	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{name: "camel case", expression: `__ccodeNative.string.camelCase("HTTP server ID")`, expected: stringToCamelCase("HTTP server ID", nil)},
		{name: "pascal case", expression: `__ccodeNative.string.pascalCase("HTTP server ID")`, expected: stringToPascalCase("HTTP server ID", nil)},
		{name: "title case", expression: `__ccodeNative.string.titleCase("HTTP server ID")`, expected: stringToTitleCase("HTTP server ID", nil)},
		{name: "sentence case", expression: `__ccodeNative.string.sentenceCase("HTTP server ID")`, expected: stringToSentenceCase("HTTP server ID", nil)},
		{name: "Go exported identifier", expression: `__ccodeNative.go.toExportedIdentifier("HTTP server ID")`, expected: stringToGoExported("HTTP server ID", nil)},
		{name: "Go unexported identifier", expression: `__ccodeNative.go.toUnexportedIdentifier("HTTP server ID")`, expected: stringToGoUnexported("HTTP server ID", nil)},
		{name: "TypeScript type identifier", expression: `__ccodeNative.typescript.toTypeIdentifier("HTTP server ID")`, expected: stringToTypeScriptTypeIdentifier("HTTP server ID", nil)},
		{name: "TypeScript value identifier", expression: `__ccodeNative.typescript.toValueIdentifier("HTTP server ID")`, expected: stringToTypeScriptValueIdentifier("HTTP server ID", nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := runtime.RunString(test.expression)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual.String())
		})
	}
}

func TestRunnerNativeUtilitiesThrowCatchableJavaScriptErrors(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))

	tests := []struct {
		name       string
		expression string
		expected   string
	}{
		{name: "missing string input", expression: `__ccodeNative.string.camelCase()`, expected: "native string.camelCase requires a string value"},
		{name: "invalid Go input", expression: `__ccodeNative.go.toExportedIdentifier(42)`, expected: "native go.toExportedIdentifier requires a string value"},
		{name: "initialisms are not an array", expression: `__ccodeNative.string.camelCase("value", "API")`, expected: "native string.camelCase initialisms must be an array of strings"},
		{name: "initialism is not a string", expression: `__ccodeNative.string.pascalCase("value", ["API", 42])`, expected: "native string.pascalCase initialisms[1] must be a string"},
		{name: "initialism is blank", expression: `__ccodeNative.go.toUnexportedIdentifier("value", [" "])`, expected: "native go.toUnexportedIdentifier initialisms[0] must not be blank"},
		{name: "invalid TypeScript type input", expression: `__ccodeNative.typescript.toTypeIdentifier(42)`, expected: "native typescript.toTypeIdentifier requires a string value"},
		{name: "TypeScript initialisms are not an array", expression: `__ccodeNative.typescript.toValueIdentifier("value", "ID")`, expected: "native typescript.toValueIdentifier initialisms must be an array of strings"},
		{name: "TypeScript initialism is not a string", expression: `__ccodeNative.typescript.toValueIdentifier("value", [42])`, expected: "native typescript.toValueIdentifier initialisms[0] must be a string"},
		{name: "TypeScript initialism is blank", expression: `__ccodeNative.typescript.toTypeIdentifier("value", [" "])`, expected: "native typescript.toTypeIdentifier initialisms[0] must not be blank"},
		{name: "OpenAPI path input missing", expression: `__ccodeNative.openapi.path.toColon()`, expected: "native openapi.path.toColon requires a string value"},
		{name: "OpenAPI path input invalid", expression: `__ccodeNative.openapi.path.toSquareBrackets(42)`, expected: "native openapi.path.toSquareBrackets requires a string value"},
		{name: "OpenAPI path option string", expression: `__ccodeNative.openapi.path.toAngleBrackets("/users/{id}", "true")`, expected: "native openapi.path.toAngleBrackets omitLeadingSlash must be a boolean"},
		{name: "OpenAPI path option undefined", expression: `__ccodeNative.openapi.path.toDollar("/users/{id}", undefined)`, expected: "native openapi.path.toDollar omitLeadingSlash must be a boolean"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			script := fmt.Sprintf(`
                try {
                  %s;
                  "not thrown";
                } catch (error) {
                  error instanceof TypeError ? error.message : "wrong error type";
                }
            `, test.expression)
			actual, err := runtime.RunString(script)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual.String())
		})
	}
}
