package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToTypeScriptTypeDefaults(t *testing.T) {
	groups := []struct {
		expected string
		names    []string
	}{
		{"boolean", []string{"bool", "boolean"}},
		{"string", []string{"string", "text", "varchar", "nvarchar", "char", "nchar", "character", "character varying", "tinytext", "mediumtext", "longtext", "clob", "uuid"}},
		{"number", []string{"int", "integer", "short", "smallint", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32", "byte", "number", "float", "double", "real", "float32", "float64", "double precision"}},
		{"bigint", []string{"long", "bigint", "int64", "uint64"}},
		{"Date", []string{"Date", "datetime", "date-time", "timestamp", "timestamptz", "time", "timetz", "timestamp with time zone", "timestamp without time zone", "time with time zone", "time without time zone"}},
		{"Uint8Array", []string{"binary", "varbinary", "blob", "bytea", "tinyblob", "mediumblob", "longblob", "Uint8Array"}},
		{"unknown", []string{"json", "jsonb", "unknown"}},
		{"any", []string{"any"}},
		{"never", []string{"never"}},
		{"void", []string{"void"}},
		{"null", []string{"null"}},
		{"undefined", []string{"undefined"}},
		{"symbol", []string{"symbol"}},
	}
	for _, group := range groups {
		for _, name := range group.names {
			t.Run(name, func(t *testing.T) {
				actual, err := stringToTypeScriptType(name, typeMappingOptions{})
				require.NoError(t, err)
				assert.Equal(t, group.expected, actual)
			})
		}
	}
}

func TestStringToTypeScriptTypeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  map[string]any
		expected string
	}{
		{name: "normalized name", input: " DOUBLE\tPRECISION\n", expected: "number"},
		{name: "SQL length", input: "CHARACTER VARYING (255)", expected: "string"},
		{name: "SQL temporal precision", input: "timestamp(6) with time zone", expected: "Date"},
		{name: "unsigned SQL expression", input: "BIGINT(20) UNSIGNED", options: map[string]any{"source": "mysql"}, expected: "bigint"},
		{name: "unsigned flag", input: "int64", options: map[string]any{"unsigned": true}, expected: "bigint"},
		{name: "unsigned 32-bit", input: "int32", options: map[string]any{"unsigned": true}, expected: "number"},
		{name: "openapi int64", input: "integer", options: map[string]any{"source": " OpenAPI ", "format": " INT64 "}, expected: "bigint"},
		{name: "openapi int32", input: "integer", options: map[string]any{"format": "int32"}, expected: "number"},
		{name: "number int64", input: "number", options: map[string]any{"format": "int64"}, expected: "bigint"},
		{name: "number int32", input: "number", options: map[string]any{"format": "int32"}, expected: "number"},
		{name: "float", input: "number", options: map[string]any{"format": "float"}, expected: "number"},
		{name: "double", input: "number", options: map[string]any{"format": "double"}, expected: "number"},
		{name: "date", input: "string", options: map[string]any{"format": "date"}, expected: "Date"},
		{name: "date-time", input: "string", options: map[string]any{"format": "date-time"}, expected: "Date"},
		{name: "time", input: "string", options: map[string]any{"format": "time"}, expected: "Date"},
		{name: "byte", input: "string", options: map[string]any{"format": "byte"}, expected: "Uint8Array"},
		{name: "binary", input: "string", options: map[string]any{"format": "binary"}, expected: "Uint8Array"},
		{name: "email", input: "string", options: map[string]any{"format": "email"}, expected: "string"},
		{name: "unknown format", input: "string", options: map[string]any{"format": "custom"}, expected: "string"},
		{name: "incompatible format", input: "boolean", options: map[string]any{"format": "int64"}, expected: "boolean"},
		{name: "format cannot identify unknown type", input: "custom", options: map[string]any{"format": "int64", "fallback": "unknown"}, expected: "unknown"},
		{name: "Postgres int2", input: "int2", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres int4", input: "int4", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres int8", input: "int8", options: map[string]any{"source": "postgresql"}, expected: "bigint"},
		{name: "Postgres smallserial", input: "smallserial", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres serial", input: "serial", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres bigserial", input: "bigserial", options: map[string]any{"source": "postgresql"}, expected: "bigint"},
		{name: "Postgres float4", input: "float4", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres float8", input: "float8", options: map[string]any{"source": "postgresql"}, expected: "number"},
		{name: "Postgres bpchar", input: "bpchar", options: map[string]any{"source": "postgresql"}, expected: "string"},
		{name: "MySQL tinyint", input: "tinyint(1)", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL smallint", input: "smallint", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL mediumint", input: "mediumint", options: map[string]any{"source": "mysql", "unsigned": true}, expected: "number"},
		{name: "MySQL int", input: "integer", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL bigint", input: "bigint", options: map[string]any{"source": "mysql", "unsigned": true}, expected: "bigint"},
		{name: "MySQL float", input: "float", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL double", input: "double", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL real", input: "real", options: map[string]any{"source": "mysql"}, expected: "number"},
		{name: "MySQL enum", input: "enum", options: map[string]any{"source": "mysql"}, expected: "string"},
		{name: "MariaDB bigint", input: "bigint", options: map[string]any{"source": "mariadb"}, expected: "bigint"},
		{name: "SQLite integer", input: "integer", options: map[string]any{"source": "sqlite"}, expected: "bigint"},
		{name: "SQLite int8", input: "int8", options: map[string]any{"source": "sqlite"}, expected: "bigint"},
		{name: "SQLite unsigned declaration", input: "unsigned big int", options: map[string]any{"source": "sqlite"}, expected: "bigint"},
		{name: "SQLite real", input: "real", options: map[string]any{"source": "sqlite"}, expected: "number"},
		{name: "SQLite float", input: "float", options: map[string]any{"source": "sqlite"}, expected: "number"},
		{name: "exact override precedence", input: "INTEGER", options: map[string]any{"source": "sqlite", "format": "int64", "unsigned": true, "overrides": map[string]any{" integer ": "Custom.ID"}}, expected: "Custom.ID"},
		{name: "profile override", input: "int8", options: map[string]any{"source": "postgresql", "overrides": map[string]any{"int8": "number"}}, expected: "number"},
		{name: "normalized decimal override", input: "numeric(10, 2)", options: map[string]any{"overrides": map[string]any{"numeric": "Decimal"}}, expected: "Decimal"},
		{name: "strict known type", input: "boolean", options: map[string]any{"fallback": nil}, expected: "boolean"},
		{name: "fallback unknown", input: "unrecognized", options: map[string]any{"fallback": "unknown"}, expected: "unknown"},
		{name: "fallback not applied to known type", input: "int64", options: map[string]any{"fallback": "number"}, expected: "bigint"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options, err := parseTypeMappingOptions(test.options)
			require.NoError(t, err)
			actual, err := stringToTypeScriptType(test.input, options)
			require.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestStringToTypeScriptTypeStrictFailures(t *testing.T) {
	for _, input := range []string{"", " \n", "custom", "decimal", "numeric(10,2)", "money", "interval", "geometry", "array", "object", "string[]", "Array<string>", "string | null", "{ id: number }", "varchar(255)[]", "time.Time", "json.RawMessage", "complex64", "uintptr", "uint64_t", "u64"} {
		t.Run(input, func(t *testing.T) {
			_, err := stringToTypeScriptType(input, typeMappingOptions{})
			require.Error(t, err)
		})
	}
}
