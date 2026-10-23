package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToGoTypeDefaults(t *testing.T) {
	groups := []struct {
		expected string
		names    []string
	}{
		{"bool", []string{"bool", "boolean"}},
		{"string", []string{"string", "text", "varchar", "nvarchar", "char", "nchar", "character", "character varying", "tinytext", "mediumtext", "longtext", "clob", "uuid"}},
		{"int", []string{"int", "integer"}},
		{"int8", []string{"int8"}},
		{"int16", []string{"int16", "short", "smallint"}},
		{"int32", []string{"int32"}},
		{"int64", []string{"int64", "long", "bigint"}},
		{"uint", []string{"uint"}},
		{"uint8", []string{"uint8"}},
		{"uint16", []string{"uint16"}},
		{"uint32", []string{"uint32"}},
		{"uint64", []string{"uint64"}},
		{"uintptr", []string{"uintptr"}},
		{"byte", []string{"byte"}},
		{"rune", []string{"rune"}},
		{"float32", []string{"float32", "float", "real"}},
		{"float64", []string{"float64", "double", "double precision", "number"}},
		{"complex64", []string{"complex64"}},
		{"complex128", []string{"complex128"}},
		{"any", []string{"any", "interface{}", "interface {}"}},
		{"error", []string{"error"}},
		{"time.Time", []string{"date", "datetime", "date-time", "timestamp", "timestamptz", "time", "time.Time", "timetz", "timestamp with time zone", "timestamp without time zone", "time with time zone", "time without time zone"}},
		{"[]byte", []string{"binary", "varbinary", "blob", "bytea", "tinyblob", "mediumblob", "longblob"}},
		{"json.RawMessage", []string{"json", "jsonb"}},
	}
	for _, group := range groups {
		for _, name := range group.names {
			t.Run(name, func(t *testing.T) {
				actual, err := stringToGoType(name, typeMappingOptions{})
				require.NoError(t, err)
				assert.Equal(t, group.expected, actual)
			})
		}
	}
}

func TestStringToGoTypeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  map[string]any
		expected string
	}{
		{name: "normalized name", input: "  DOUBLE \t PRECISION\n", expected: "float64"},
		{name: "SQL string length", input: "VARCHAR(255)", expected: "string"},
		{name: "SQL spaced length", input: "CHARACTER VARYING ( 255 )", expected: "string"},
		{name: "SQL timestamp precision", input: "timestamp(6) with time zone", expected: "time.Time"},
		{name: "integer format", input: "integer", options: map[string]any{"format": " INT64 ", "source": " OpenAPI "}, expected: "int64"},
		{name: "int32 format", input: "integer", options: map[string]any{"format": "int32"}, expected: "int32"},
		{name: "number int64", input: "number", options: map[string]any{"format": "int64"}, expected: "int64"},
		{name: "number int32", input: "number", options: map[string]any{"format": "int32"}, expected: "int32"},
		{name: "float format", input: "number", options: map[string]any{"format": "float"}, expected: "float32"},
		{name: "double format", input: "number", options: map[string]any{"format": "double"}, expected: "float64"},
		{name: "date format", input: "string", options: map[string]any{"format": "date"}, expected: "time.Time"},
		{name: "date-time format", input: "string", options: map[string]any{"format": "date-time"}, expected: "time.Time"},
		{name: "time format", input: "string", options: map[string]any{"format": "time"}, expected: "time.Time"},
		{name: "byte format", input: "string", options: map[string]any{"format": "byte"}, expected: "[]byte"},
		{name: "binary format", input: "string", options: map[string]any{"format": "binary"}, expected: "[]byte"},
		{name: "uuid format", input: "string", options: map[string]any{"format": "uuid"}, expected: "string"},
		{name: "unknown format", input: "string", options: map[string]any{"format": "special"}, expected: "string"},
		{name: "incompatible format", input: "boolean", options: map[string]any{"format": "int64"}, expected: "bool"},
		{name: "unknown type with known format", input: "custom", options: map[string]any{"format": "int64", "fallback": "any"}, expected: "any"},
		{name: "unsigned native int", input: "int", options: map[string]any{"unsigned": true}, expected: "uint"},
		{name: "unsigned int8", input: "int8", options: map[string]any{"unsigned": true}, expected: "uint8"},
		{name: "unsigned int16", input: "smallint", options: map[string]any{"unsigned": true}, expected: "uint16"},
		{name: "unsigned format", input: "integer", options: map[string]any{"format": "int32", "unsigned": true}, expected: "uint32"},
		{name: "unsigned idempotent", input: "uint64", options: map[string]any{"unsigned": true}, expected: "uint64"},
		{name: "unsigned does not change float", input: "number", options: map[string]any{"unsigned": true}, expected: "float64"},
		{name: "unsigned SQL declaration", input: "BIGINT(20) UNSIGNED", options: map[string]any{"source": "mysql"}, expected: "uint64"},
		{name: "explicit signed", input: "bigint", options: map[string]any{"source": "mysql", "unsigned": false}, expected: "int64"},
		{name: "Postgres int8", input: "int8", options: map[string]any{"source": "postgresql"}, expected: "int64"},
		{name: "Postgres int", input: "int", options: map[string]any{"source": "postgresql"}, expected: "int32"},
		{name: "Postgres int2", input: "int2", options: map[string]any{"source": "postgresql"}, expected: "int16"},
		{name: "Postgres int4", input: "int4", options: map[string]any{"source": "postgresql"}, expected: "int32"},
		{name: "Postgres smallserial", input: "smallserial", options: map[string]any{"source": "postgresql"}, expected: "int16"},
		{name: "Postgres serial", input: "serial", options: map[string]any{"source": "postgresql"}, expected: "int32"},
		{name: "Postgres bigserial", input: "bigserial", options: map[string]any{"source": "postgresql"}, expected: "int64"},
		{name: "Postgres float4", input: "float4", options: map[string]any{"source": "postgresql"}, expected: "float32"},
		{name: "Postgres float8", input: "float8", options: map[string]any{"source": "postgresql"}, expected: "float64"},
		{name: "Postgres float", input: "float", options: map[string]any{"source": "postgresql"}, expected: "float64"},
		{name: "MySQL tinyint", input: "tinyint", options: map[string]any{"source": "mysql"}, expected: "int8"},
		{name: "MySQL tinyint is not boolean", input: "tinyint(1)", options: map[string]any{"source": "mysql"}, expected: "int8"},
		{name: "MySQL smallint", input: "smallint", options: map[string]any{"source": "mysql"}, expected: "int16"},
		{name: "MySQL mediumint", input: "mediumint", options: map[string]any{"source": "mysql", "unsigned": true}, expected: "uint32"},
		{name: "MySQL int", input: "integer", options: map[string]any{"source": "mysql"}, expected: "int32"},
		{name: "MySQL bigint", input: "bigint", options: map[string]any{"source": "mysql", "unsigned": true}, expected: "uint64"},
		{name: "MySQL enum", input: "enum", options: map[string]any{"source": "mysql"}, expected: "string"},
		{name: "MySQL real", input: "real", options: map[string]any{"source": "mysql"}, expected: "float64"},
		{name: "MariaDB mediumint", input: "mediumint", options: map[string]any{"source": "mariadb"}, expected: "int32"},
		{name: "SQLite int", input: "integer", options: map[string]any{"source": "sqlite"}, expected: "int64"},
		{name: "SQLite smallint", input: "smallint", options: map[string]any{"source": "sqlite"}, expected: "int64"},
		{name: "SQLite unsigned declaration remains signed", input: "bigint unsigned", options: map[string]any{"source": "sqlite"}, expected: "int64"},
		{name: "SQLite real", input: "real", options: map[string]any{"source": "sqlite"}, expected: "float64"},
		{name: "SQLite float", input: "float", options: map[string]any{"source": "sqlite"}, expected: "float64"},
		{name: "case insensitive override", input: "LONG", options: map[string]any{"overrides": map[string]any{" Long ": "Custom.Int"}}, expected: "Custom.Int"},
		{name: "override before format and unsigned", input: "integer", options: map[string]any{"format": "int64", "unsigned": true, "overrides": map[string]any{"integer": "int"}}, expected: "int"},
		{name: "override before profile", input: "int8", options: map[string]any{"source": "postgresql", "overrides": map[string]any{"int8": "string"}}, expected: "string"},
		{name: "decimal override", input: "NUMERIC(10, 2)", options: map[string]any{"overrides": map[string]any{"numeric": "decimal.Decimal"}}, expected: "decimal.Decimal"},
		{name: "explicit fallback", input: "custom", options: map[string]any{"fallback": "any"}, expected: "any"},
		{name: "fallback is not applied to known type", input: "boolean", options: map[string]any{"fallback": "any"}, expected: "bool"},
		{name: "null fallback accepts known types", input: "boolean", options: map[string]any{"fallback": nil}, expected: "bool"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options, err := parseTypeMappingOptions(test.options)
			require.NoError(t, err)
			actual, err := stringToGoType(test.input, options)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStringToGoTypeStrictFailures(t *testing.T) {
	for _, input := range []string{"", " \n ", "custom", "decimal", "numeric(10,2)", "money", "interval", "geometry", "array", "object", "null", "integer[]", "[]int", "int|string", "map[string]int", "varchar(255)[]", "uint64_t", "u64", "i32"} {
		t.Run(input, func(t *testing.T) {
			_, err := stringToGoType(input, typeMappingOptions{})
			require.Error(t, err)
		})
	}
	options, err := parseTypeMappingOptions(map[string]any{"fallback": nil})
	require.NoError(t, err)
	_, err = stringToGoType("unrecognized", options)
	require.EqualError(t, err, `cannot infer Go type from "unrecognized"`)
}

func TestGoTypeOptionsValidation(t *testing.T) {
	tests := []struct {
		options map[string]any
		error   string
	}{
		{map[string]any{"format": 42}, "format must be a string"},
		{map[string]any{"format": nil}, "format must be a string"},
		{map[string]any{"source": nil}, "source must be a string"},
		{map[string]any{"source": "sqlserver"}, `unsupported source "sqlserver"`},
		{map[string]any{"source": ""}, `unsupported source ""`},
		{map[string]any{"unsigned": "true"}, "unsigned must be a boolean"},
		{map[string]any{"fallback": false}, "fallback must be a non-blank string or null"},
		{map[string]any{"fallback": " "}, "fallback must be a non-blank string or null"},
		{map[string]any{"overrides": nil}, "overrides must be a mapping"},
		{map[string]any{"overrides": []any{}}, "overrides must be a mapping"},
		{map[string]any{"overrides": map[string]any{"": "int"}}, "non-blank string key and value"},
		{map[string]any{"overrides": map[string]any{"long": " "}}, "non-blank string key and value"},
		{map[string]any{"overrides": map[string]any{"long": 42}}, "non-blank string key and value"},
		{map[string]any{"overrides": map[string]any{"long": "int", " LONG ": "int64"}}, "conflicting overrides"},
		{map[string]any{"precision": 10}, `unexpected option "precision"`},
	}
	for _, test := range tests {
		t.Run(test.error, func(t *testing.T) {
			_, err := parseTypeMappingOptions(test.options)
			require.ErrorContains(t, err, test.error)
		})
	}
}
