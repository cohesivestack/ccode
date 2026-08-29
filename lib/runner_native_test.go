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

	for _, name := range []string{"Go", "Strings", "go", "string"} {
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
