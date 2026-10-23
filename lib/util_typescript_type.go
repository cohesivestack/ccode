package ccode

import "fmt"

// Each target owns its mapping policy. Options and SQL normalization are shared
// with Go.toType, but TypeScript never inherits Go-specific type expressions.
var defaultTypeScriptTypes = map[string]string{
	"bool": "boolean", "boolean": "boolean",
	"string": "string", "text": "string", "varchar": "string", "nvarchar": "string",
	"char": "string", "nchar": "string", "character": "string", "character varying": "string",
	"tinytext": "string", "mediumtext": "string", "longtext": "string", "clob": "string", "uuid": "string",
	"int": "number", "integer": "number", "short": "number", "smallint": "number",
	"int8": "number", "int16": "number", "int32": "number",
	"uint": "number", "uint8": "number", "uint16": "number", "uint32": "number", "byte": "number",
	"long": "bigint", "bigint": "bigint", "int64": "bigint", "uint64": "bigint",
	"number": "number", "float": "number", "double": "number", "real": "number",
	"float32": "number", "float64": "number", "double precision": "number",
	"date": "Date", "datetime": "Date", "date-time": "Date", "timestamp": "Date", "timestamptz": "Date",
	"time": "Date", "timetz": "Date", "timestamp with time zone": "Date", "timestamp without time zone": "Date",
	"time with time zone": "Date", "time without time zone": "Date",
	"binary": "Uint8Array", "varbinary": "Uint8Array", "blob": "Uint8Array", "bytea": "Uint8Array",
	"tinyblob": "Uint8Array", "mediumblob": "Uint8Array", "longblob": "Uint8Array", "uint8array": "Uint8Array",
	"json": "unknown", "jsonb": "unknown", "any": "any", "unknown": "unknown", "never": "never",
	"void": "void", "null": "null", "undefined": "undefined", "symbol": "symbol",
}

var postgresTypeScriptTypes = map[string]string{
	"smallint": "number", "int2": "number", "smallserial": "number", "serial2": "number",
	"integer": "number", "int": "number", "int4": "number", "serial": "number", "serial4": "number",
	"bigint": "bigint", "int8": "bigint", "bigserial": "bigint", "serial8": "bigint",
	"real": "number", "float": "number", "float4": "number", "float8": "number", "double precision": "number",
	"bpchar": "string",
}

var mysqlTypeScriptTypes = map[string]string{
	"tinyint": "number", "smallint": "number", "mediumint": "number", "int": "number", "integer": "number",
	"bigint": "bigint", "float": "number", "double": "number", "double precision": "number",
	"real": "number", "enum": "string",
}

var sqliteTypeScriptTypes = map[string]string{
	"integer": "bigint", "int": "bigint", "tinyint": "bigint", "smallint": "bigint", "mediumint": "bigint",
	"bigint": "bigint", "int2": "bigint", "int8": "bigint", "unsigned big int": "bigint",
	"real": "number", "float": "number", "double": "number", "double precision": "number",
}

func stringToTypeScriptType(input string, options typeMappingOptions) (string, error) {
	name := normalizeTypeName(input)
	if name == "" {
		return "", fmt.Errorf("type name must not be blank")
	}
	if override, ok := options.overrides[name]; ok {
		return override, nil
	}
	// Signedness is accepted and normalized, but does not select a different
	// TypeScript type: both signed and unsigned 64-bit integers use bigint.
	name, _ = normalizeSQLTypeName(name, options)
	if override, ok := options.overrides[name]; ok {
		return override, nil
	}

	var profile map[string]string
	switch options.source {
	case "postgresql":
		profile = postgresTypeScriptTypes
	case "mysql", "mariadb":
		profile = mysqlTypeScriptTypes
	case "sqlite":
		profile = sqliteTypeScriptTypes
	}
	result, ok := profile[name]
	if !ok {
		result, ok = defaultTypeScriptTypes[name]
	}
	if !ok {
		if options.fallback != nil {
			return *options.fallback, nil
		}
		return "", fmt.Errorf("cannot infer TypeScript type from %q", input)
	}
	if name == "integer" || name == "number" {
		switch options.format {
		case "int64":
			result = "bigint"
		case "int32":
			result = "number"
		case "float", "double":
			if name == "number" {
				result = "number"
			}
		}
	}
	if name == "string" {
		switch options.format {
		case "date", "date-time", "time":
			result = "Date"
		case "byte", "binary":
			result = "Uint8Array"
		}
	}
	return result, nil
}
