package ccode

import "fmt"

// The tables are read-only. Per-call overrides never modify shared defaults.
var defaultGoTypes = map[string]string{
	"bool": "bool", "boolean": "bool",
	"string": "string", "text": "string", "varchar": "string", "nvarchar": "string",
	"char": "string", "nchar": "string", "character": "string", "character varying": "string",
	"tinytext": "string", "mediumtext": "string", "longtext": "string", "clob": "string",
	"int": "int", "integer": "int", "int8": "int8",
	"int16": "int16", "short": "int16", "smallint": "int16", "int32": "int32",
	"int64": "int64", "long": "int64", "bigint": "int64",
	"uint": "uint", "uint8": "uint8", "uint16": "uint16", "uint32": "uint32", "uint64": "uint64",
	"uintptr": "uintptr", "byte": "byte", "rune": "rune",
	"float32": "float32", "float": "float32", "real": "float32",
	"float64": "float64", "double": "float64", "double precision": "float64", "number": "float64",
	"complex64": "complex64", "complex128": "complex128",
	"any": "any", "interface{}": "any", "interface {}": "any", "error": "error",
	"date": "time.Time", "datetime": "time.Time", "date-time": "time.Time",
	"timestamp": "time.Time", "timestamptz": "time.Time", "time": "time.Time", "time.time": "time.Time",
	"timestamp with time zone": "time.Time", "timestamp without time zone": "time.Time",
	"time with time zone": "time.Time", "time without time zone": "time.Time", "timetz": "time.Time",
	"binary": "[]byte", "varbinary": "[]byte", "blob": "[]byte", "bytea": "[]byte",
	"tinyblob": "[]byte", "mediumblob": "[]byte", "longblob": "[]byte",
	"json": "json.RawMessage", "jsonb": "json.RawMessage", "uuid": "string",
}

var postgresGoTypes = map[string]string{
	"smallint": "int16", "int2": "int16", "smallserial": "int16", "serial2": "int16",
	"integer": "int32", "int": "int32", "int4": "int32", "serial": "int32", "serial4": "int32",
	"bigint": "int64", "int8": "int64", "bigserial": "int64", "serial8": "int64",
	"real": "float32", "float4": "float32", "float": "float64", "float8": "float64",
	"double precision": "float64", "bpchar": "string",
}

var mysqlGoTypes = map[string]string{
	"tinyint": "int8", "smallint": "int16", "mediumint": "int32", "int": "int32", "integer": "int32",
	"bigint": "int64", "float": "float32", "double": "float64", "double precision": "float64",
	"real": "float64", "enum": "string",
}

var sqliteGoTypes = map[string]string{
	"integer": "int64", "int": "int64", "tinyint": "int64", "smallint": "int64", "mediumint": "int64",
	"bigint": "int64", "int2": "int64", "int8": "int64", "unsigned big int": "int64",
	"real": "float64", "float": "float64", "double": "float64", "double precision": "float64",
}

func stringToGoType(input string, options typeMappingOptions) (string, error) {
	name := normalizeTypeName(input)
	if name == "" {
		return "", fmt.Errorf("type name must not be blank")
	}
	if override, ok := options.overrides[name]; ok {
		return override, nil
	}

	name, unsigned := normalizeSQLTypeName(name, options)
	if override, ok := options.overrides[name]; ok {
		return override, nil
	}

	var profile map[string]string
	switch options.source {
	case "postgresql":
		profile = postgresGoTypes
	case "mysql", "mariadb":
		profile = mysqlGoTypes
	case "sqlite":
		profile = sqliteGoTypes
	}
	result, ok := profile[name]
	if !ok {
		result, ok = defaultGoTypes[name]
	}
	if !ok {
		if options.fallback != nil {
			return *options.fallback, nil
		}
		return "", fmt.Errorf("cannot infer Go type from %q", input)
	}

	// A format refines only a compatible base type. Unknown formats retain it.
	if name == "integer" || name == "number" {
		switch options.format {
		case "int32":
			result = "int32"
		case "int64":
			result = "int64"
		case "float":
			if name == "number" {
				result = "float32"
			}
		case "double":
			if name == "number" {
				result = "float64"
			}
		}
	}
	if name == "string" {
		switch options.format {
		case "date", "date-time", "time":
			result = "time.Time"
		case "byte", "binary":
			result = "[]byte"
		}
	}
	if unsigned {
		switch result {
		case "int", "int8", "int16", "int32", "int64":
			result = "u" + result
		}
	}
	return result, nil
}
