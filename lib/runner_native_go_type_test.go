package ccode

import (
	"fmt"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeGoTypeValidation(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))
	tests := []struct {
		expression string
		error      string
	}{
		{`__ccodeNative.go.toType()`, "requires a string value"},
		{`__ccodeNative.go.toType(42)`, "requires a string value"},
		{`__ccodeNative.go.toType(" ")`, "type name must not be blank"},
		{`__ccodeNative.go.toType("custom")`, `cannot infer Go type from "custom"`},
		{`__ccodeNative.go.toType("custom", {fallback: null})`, "cannot infer Go type"},
		{`__ccodeNative.go.toType("int", null)`, "options must be an object"},
		{`__ccodeNative.go.toType("int", [])`, "options must be an object"},
		{`__ccodeNative.go.toType("int", true)`, "options must be an object"},
		{`__ccodeNative.go.toType("int", {format: 42})`, "format must be a string"},
		{`__ccodeNative.go.toType("int", {format: null})`, "format must be a string"},
		{`__ccodeNative.go.toType("int", {source: "other"})`, "unsupported source"},
		{`__ccodeNative.go.toType("int", {unsigned: "true"})`, "unsigned must be a boolean"},
		{`__ccodeNative.go.toType("int", {fallback: 42})`, "fallback must be a non-blank string or null"},
		{`__ccodeNative.go.toType("int", {fallback: " "})`, "fallback must be a non-blank string or null"},
		{`__ccodeNative.go.toType("int", {overrides: []})`, "overrides must be a mapping"},
		{`__ccodeNative.go.toType("int", {overrides: null})`, "overrides must be a mapping"},
		{`__ccodeNative.go.toType("int", {overrides: {long: undefined}})`, "non-blank string key and value"},
		{`__ccodeNative.go.toType("int", {overrides: {long: false}})`, "non-blank string key and value"},
		{`__ccodeNative.go.toType("int", {overrides: {long: ""}})`, "non-blank string key and value"},
		{`__ccodeNative.go.toType("int", {overrides: {long: "int", " LONG ": "int64"}})`, "conflicting overrides"},
		{`__ccodeNative.go.toType("int", {precision: 10})`, "unexpected option"},
		{`__ccodeNative.go.toType("int", {foramt: undefined})`, "unexpected option"},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			actual, err := runtime.RunString(fmt.Sprintf(`
try { %s; "not thrown"; }
catch (error) { error instanceof TypeError ? error.message : "wrong error type"; }
`, test.expression))
			require.NoError(t, err)
			assert.Contains(t, actual.String(), "native go.toType ")
			assert.Contains(t, actual.String(), test.error)
		})
	}
}

func TestNativeGoTypeOptionalValuesAndImmutableDefaults(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))
	actual, err := runtime.RunString(`
const options = Object.freeze({overrides: Object.freeze({" LONG ": "Custom.Int"})});
JSON.stringify([
  __ccodeNative.go.toType("long", options),
  __ccodeNative.go.toType("long"),
  __ccodeNative.go.toType("integer", undefined),
  __ccodeNative.go.toType("integer", {format: undefined, unsigned: undefined, source: undefined, overrides: undefined, fallback: undefined}),
  __ccodeNative.go.toType("custom", {fallback: "any"}),
  __ccodeNative.go.toType("string", {format: "date-time"}),
  __ccodeNative.go.toType("bigint", {source: "mysql", unsigned: true}),
]);
`)
	require.NoError(t, err)
	assert.JSONEq(t, `["Custom.Int","int64","int","int","any","time.Time","uint64"]`, actual.String())
}
