package ccode

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeTypeScriptTypeValidation(t *testing.T) {
	runtime := goja.New()
	require.NoError(t, installRunnerNativeUtilities(runtime))
	tests := []struct {
		expression string
		error      string
	}{
		{`__ccodeNative.typescript.toType()`, "requires a string value"},
		{`__ccodeNative.typescript.toType(42)`, "requires a string value"},
		{`__ccodeNative.typescript.toType(" ")`, "type name must not be blank"},
		{`__ccodeNative.typescript.toType("custom")`, `cannot infer TypeScript type from "custom"`},
		{`__ccodeNative.typescript.toType("custom", {fallback: null})`, "cannot infer TypeScript type"},
		{`__ccodeNative.typescript.toType("int", null)`, "options must be an object"},
		{`__ccodeNative.typescript.toType("int", [])`, "options must be an object"},
		{`__ccodeNative.typescript.toType("int", {format: null})`, "format must be a string"},
		{`__ccodeNative.typescript.toType("int", {source: "other"})`, "unsupported source"},
		{`__ccodeNative.typescript.toType("int", {unsigned: "true"})`, "unsigned must be a boolean"},
		{`__ccodeNative.typescript.toType("int", {fallback: 42})`, "fallback must be a non-blank string or null"},
		{`__ccodeNative.typescript.toType("int", {fallback: " "})`, "fallback must be a non-blank string or null"},
		{`__ccodeNative.typescript.toType("int", {overrides: []})`, "overrides must be a mapping"},
		{`__ccodeNative.typescript.toType("int", {overrides: {long: undefined}})`, "non-blank string key and value"},
		{`__ccodeNative.typescript.toType("int", {overrides: {long: "number", " LONG ": "bigint"}})`, "conflicting overrides"},
		{`__ccodeNative.typescript.toType("int", {precision: 10})`, "unexpected option"},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			actual, err := runtime.RunString(fmt.Sprintf(`
try { %s; "not thrown"; }
catch (error) { error instanceof TypeError ? error.message : "wrong error type"; }
`, test.expression))
			require.NoError(t, err)
			assert.Contains(t, actual.String(), "native typescript.toType ")
			assert.Contains(t, actual.String(), test.error)
		})
	}
}

func TestRunnerTypeScriptToTypePublicAPI(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunnerTypeScriptToTypePublicAPI")
	processFile := filepath.Join(projectDir, "types.proc.ts")
	require.NoError(t, os.WriteFile(processFile, []byte(`
import type { Context } from "@ccode/context";
import * as Go from "@ccode/go";
import * as TypeScript from "@ccode/typescript";
import { toType, type TypeOptions } from "@ccode/typescript";

const options: TypeScript.TypeOptions = Object.freeze({
  source: "postgresql",
  overrides: Object.freeze({numeric: "Decimal"}),
});
const permissive: TypeOptions = {fallback: "unknown"};
const results = {
  boolean: TypeScript.toType("boolean"),
  integer: TypeScript.toType("integer"),
  int64: TypeScript.toType("integer", {source: "openapi", format: "int64"}),
  date: TypeScript.toType("string", {format: "date-time"}),
  binary: TypeScript.toType("string", {format: "binary"}),
  json: TypeScript.toType("jsonb"),
  nullType: TypeScript.toType("null"),
  postgres: TypeScript.toType("int8", {source: "postgresql"}),
  mysql: TypeScript.toType("bigint", {source: "mysql", unsigned: true}),
  sqlite: TypeScript.toType("integer", {source: "sqlite"}),
  decimal: TypeScript.toType("NUMERIC(10,2)", options),
  fallback: toType("custom", permissive),
  override: TypeScript.toType("long", {overrides: Object.freeze({long: "number"})}),
  defaultAfterOverride: TypeScript.toType("long"),
  omittedOptions: TypeScript.toType("integer", undefined),
  omittedFields: TypeScript.toType("integer", {source: undefined, format: undefined, unsigned: undefined, overrides: undefined, fallback: undefined}),
  goUnchanged: Go.toType("integer"),
  goFormatUnchanged: Go.toType("integer", {format: "int64"}),
  identifierUnchanged: TypeScript.toTypeIdentifier("integer"),
};

export default function main(ctx: Context) {
  ctx.println(JSON.stringify(results));
  try {
    TypeScript.toType("custom", {fallback: null});
    throw new Error("expected strict inference to fail");
  } catch (error) {
    if (!(error instanceof TypeError) || !error.message.includes("cannot infer TypeScript type")) throw error;
  }
}
`), 0644))
	var output bytes.Buffer
	ctx.stdout = &output
	require.NoError(t, ctx.Run("types"))
	assert.JSONEq(t, `{
  "boolean":"boolean", "integer":"number", "int64":"bigint", "date":"Date",
  "binary":"Uint8Array", "json":"unknown", "nullType":"null", "postgres":"bigint",
  "mysql":"bigint", "sqlite":"bigint", "decimal":"Decimal", "fallback":"unknown",
  "override":"number", "defaultAfterOverride":"bigint", "omittedOptions":"number",
  "omittedFields":"number", "goUnchanged":"int", "goFormatUnchanged":"int64", "identifierUnchanged":"Integer"
}`, output.String())
}
