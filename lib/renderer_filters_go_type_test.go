package ccode

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRendererGoTypeFilter(t *testing.T) {
	tests := []struct {
		source   string
		data     map[string]any
		expected string
	}{
		{source: `{{ "boolean" | goType }}`, expected: "bool"},
		{source: `{{ "integer" | goType(format="int64", source="openapi") }}`, expected: "int64"},
		{source: `{{ "string" | goType(format="date-time") }}`, expected: "time.Time"},
		{source: `{{ "bigint" | goType(source="mysql", unsigned=true) }}`, expected: "uint64"},
		{source: `{{ "int8" | goType(source="postgresql") }}`, expected: "int64"},
		{source: `{{ "int8" | goType }}`, expected: "int8"},
		{source: `{{ "float" | goType(source="sqlite") }}`, expected: "float64"},
		{source: `{{ "varchar(255)" | goType }}`, expected: "string"},
		{source: `{{ "long" | goType(overrides={" LONG ": "Custom.Int"}) }}`, expected: "Custom.Int"},
		{source: `{{ "integer" | goType(format="int64", unsigned=true, overrides={"integer": "string"}) }}`, expected: "string"},
		{source: `{{ "numeric" | goType(source="postgresql", overrides=data.overrides) }}`, data: map[string]any{"overrides": map[string]string{"numeric": "decimal.Decimal"}}, expected: "decimal.Decimal"},
		{source: `{{ "long" | goType(overrides={}) }}`, expected: "int64"},
		{source: `{{ "custom" | goType(fallback="any") }}`, expected: "any"},
		{source: `{{ "boolean" | goType(fallback=none) }}`, expected: "bool"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			actual, err := renderTemplateSource(t, test.source, test.data)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestRendererGoTypeValidation(t *testing.T) {
	tests := []struct {
		source string
		error  string
	}{
		{`{{ 42 | goType }}`, "input must be a string"},
		{`{{ " " | goType }}`, "type name must not be blank"},
		{`{{ "custom" | goType }}`, "cannot infer Go type"},
		{`{{ "custom" | goType(fallback=none) }}`, "cannot infer Go type"},
		{`{{ "int" | goType("int64") }}`, "positional arguments are not supported"},
		{`{{ "int" | goType(precision=10) }}`, "unexpected argument"},
		{`{{ "int" | goType(format=42) }}`, "format must be a string"},
		{`{{ "int" | goType(format=none) }}`, "format must be a string"},
		{`{{ "int" | goType(source="other") }}`, "unsupported source"},
		{`{{ "int" | goType(unsigned="true") }}`, "unsigned must be a boolean"},
		{`{{ "int" | goType(fallback=false) }}`, "fallback must be a non-blank string or null"},
		{`{{ "int" | goType(fallback=" ") }}`, "fallback must be a non-blank string or null"},
		{`{{ "int" | goType(overrides=[]) }}`, "overrides must be a mapping"},
		{`{{ "int" | goType(overrides=none) }}`, "overrides must be a mapping"},
		{`{{ "int" | goType(overrides={42: "int"}) }}`, "overrides must be a mapping"},
		{`{{ "int" | goType(overrides={"long": 42}) }}`, "overrides must be a mapping"},
		{`{{ "int" | goType(overrides={"long": " "}) }}`, "non-blank string key and value"},
		{`{{ "int" | goType(overrides={"long": "int", " LONG ": "int64"}) }}`, "conflicting overrides"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			_, err := renderTemplateSource(t, test.source, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `filter "goType"`)
			assert.Contains(t, err.Error(), test.error)
		})
	}
}

func TestRunnerGoTypePublicAPIAndFilterParity(t *testing.T) {
	ctx, projectDir := setupRunnerTestProject(t, "TestRunnerGoTypePublicAPIAndFilterParity")
	processFile := filepath.Join(projectDir, "types.process.ts")
	templateFile := filepath.Join(projectDir, "templates", "types.tpl")
	require.NoError(t, os.MkdirAll(filepath.Dir(templateFile), 0755))
	require.NoError(t, os.WriteFile(templateFile, []byte(`{{ "boolean" | goType }}|{{ "integer" | goType(source="openapi", format="int64") }}|{{ "string" | goType(format="date-time") }}|{{ "bigint" | goType(source="mysql", unsigned=true) }}|{{ "numeric" | goType(source="postgresql", overrides=data.overrides) }}|{{ "custom" | goType(fallback="any") }}|{{ "long" | goType }}`), 0644))
	require.NoError(t, os.WriteFile(processFile, []byte(`
import type { Context } from "@ccode/context";
import * as Go from "@ccode/go";

const overrides = Object.freeze({ numeric: "decimal.Decimal" });
const options: Go.TypeOptions = Object.freeze({ source: "postgresql", overrides });
const expected = [
  Go.toType("boolean"),
  Go.toType("integer", { source: "openapi", format: "int64" }),
  Go.toType("string", { format: "date-time" }),
  Go.toType("bigint", { source: "mysql", unsigned: true }),
  Go.toType("numeric", options),
  Go.toType("custom", { fallback: "any" }),
  Go.toType("long"),
].join("|");

export default function main(ctx: Context) {
  const rendered = ctx.renderTemplate("templates/types.tpl", { overrides });
  if (rendered !== expected) throw new Error("TypeScript/filter mismatch: " + rendered);
  const actual = [
    Go.toType("long", { overrides: { long: "int" } }),
    Go.toType("long"),
    Go.toType("integer", { format: undefined }),
  ];
  ctx.println(rendered);
  ctx.println(JSON.stringify(actual));
  try {
    Go.toType("custom", { fallback: null });
    throw new Error("expected strict inference to fail");
  } catch (error) {
    if (!(error instanceof TypeError) || !error.message.includes("cannot infer Go type")) throw error;
  }
}
`), 0644))
	var output bytes.Buffer
	ctx.stdout = &output
	require.NoError(t, ctx.Run("types"))
	assert.Equal(t, "bool|int64|time.Time|uint64|decimal.Decimal|any|int64\n[\"int\",\"int64\",\"int\"]\n", output.String())
}
