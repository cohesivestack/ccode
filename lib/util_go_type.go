package ccode

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type goTypeOptions struct {
	format    string
	source    string
	unsigned  bool
	overrides map[string]string
	fallback  *string
}

var goTypeOptionNames = []string{"format", "source", "unsigned", "overrides", "fallback"}

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

// Only numeric SQL size/precision modifiers are removed; arrays, unions,
// pointers and arbitrary expressions are never reduced to a primitive name.
var goTypeSQLParameters = regexp.MustCompile(`^([a-z][a-z0-9]*(?: [a-z][a-z0-9]*)*)\s*\(\s*[0-9]+(?:\s*,\s*[0-9]+)?\s*\)(.*)$`)

func normalizeGoTypeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

func parseGoTypeOptions(values map[string]any) (goTypeOptions, error) {
	options := goTypeOptions{}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := values[key]
		switch key {
		case "format", "source":
			text, ok := value.(string)
			if !ok {
				return options, fmt.Errorf("%s must be a string", key)
			}
			text = normalizeGoTypeName(text)
			if key == "format" {
				options.format = text
			} else {
				switch text {
				case "openapi", "postgresql", "mysql", "mariadb", "sqlite":
					options.source = text
				default:
					return options, fmt.Errorf("unsupported source %q", text)
				}
			}
		case "unsigned":
			var ok bool
			if options.unsigned, ok = value.(bool); !ok {
				return options, fmt.Errorf("unsigned must be a boolean")
			}
		case "fallback":
			if value == nil {
				continue
			}
			text, ok := value.(string)
			if !ok || isStringBlank(text) {
				return options, fmt.Errorf("fallback must be a non-blank string or null")
			}
			options.fallback = &text
		case "overrides":
			overrides, ok := value.(map[string]any)
			if !ok {
				return options, fmt.Errorf("overrides must be a mapping of non-blank strings to non-blank strings")
			}
			options.overrides = make(map[string]string, len(overrides))
			names := make([]string, 0, len(overrides))
			for name := range overrides {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				text, ok := overrides[name].(string)
				normalized := normalizeGoTypeName(name)
				if normalized == "" || !ok || isStringBlank(text) {
					return options, fmt.Errorf("override %q must have a non-blank string key and value", name)
				}
				if previous, exists := options.overrides[normalized]; exists && previous != text {
					return options, fmt.Errorf("conflicting overrides for %q", normalized)
				}
				options.overrides[normalized] = text
			}
		default:
			return options, fmt.Errorf("unexpected option %q", key)
		}
	}
	return options, nil
}

func stringToGoType(input string, options goTypeOptions) (string, error) {
	name := normalizeGoTypeName(input)
	if name == "" {
		return "", fmt.Errorf("type name must not be blank")
	}
	if override, ok := options.overrides[name]; ok {
		return override, nil
	}

	// Accept common SQL declarations in addition to inspection Type.name values.
	if parts := goTypeSQLParameters.FindStringSubmatch(name); parts != nil {
		name = normalizeGoTypeName(parts[1] + parts[2])
	}
	unsigned := options.unsigned
	if strings.HasSuffix(name, " unsigned") {
		name = strings.TrimSuffix(name, " unsigned")
		if options.source != "sqlite" {
			unsigned = true
		}
	}
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
