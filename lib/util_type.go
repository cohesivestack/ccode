package ccode

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type typeMappingOptions struct {
	format    string
	source    string
	unsigned  bool
	overrides map[string]string
	fallback  *string
}

var typeMappingOptionNames = []string{"format", "source", "unsigned", "overrides", "fallback"}

// Only numeric SQL size/precision modifiers are removed; arrays, unions,
// pointers and arbitrary expressions are never reduced to a primitive name.
var typeSQLParameters = regexp.MustCompile(`^([a-z][a-z0-9]*(?: [a-z][a-z0-9]*)*)\s*\(\s*[0-9]+(?:\s*,\s*[0-9]+)?\s*\)(.*)$`)

func normalizeTypeName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), " "))
}

func parseTypeMappingOptions(values map[string]any) (typeMappingOptions, error) {
	options := typeMappingOptions{}
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
			text = normalizeTypeName(text)
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
				normalized := normalizeTypeName(name)
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

func normalizeSQLTypeName(name string, options typeMappingOptions) (string, bool) {
	if parts := typeSQLParameters.FindStringSubmatch(name); parts != nil {
		name = normalizeTypeName(parts[1] + parts[2])
	}
	unsigned := options.unsigned
	if strings.HasSuffix(name, " unsigned") {
		name = strings.TrimSuffix(name, " unsigned")
		if options.source != "sqlite" {
			unsigned = true
		}
	}
	return name, unsigned
}
