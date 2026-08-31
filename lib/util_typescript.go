package ccode

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// stringToTypeScriptTypeIdentifier converts a string to a conventional
// PascalCase TypeScript declaration identifier.
func stringToTypeScriptTypeIdentifier(s string, initialisms []string) string {
	identifier := arrayToPascalCase(
		splitStringIntoGoWords(s),
		newTypeScriptInitialismCollection(initialisms),
	)
	if identifier == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(identifier)
	if unicode.IsDigit(first) {
		identifier = "T" + identifier
	}
	if isTypeScriptReservedBinding(identifier) {
		identifier += "_"
	}

	return identifier
}

// stringToTypeScriptValueIdentifier converts a string to a conventional
// camelCase TypeScript declaration identifier.
func stringToTypeScriptValueIdentifier(s string, initialisms []string) string {
	identifier := arrayToCamelCase(
		splitStringIntoGoWords(s),
		newTypeScriptInitialismCollection(initialisms),
	)
	if identifier == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(identifier)
	if unicode.IsDigit(first) {
		identifier = "_" + identifier
	}
	if isTypeScriptReservedBinding(identifier) {
		identifier += "_"
	}

	return identifier
}

// newTypeScriptInitialismCollection preserves requested capitalization while
// removing characters outside the conservative identifier character set.
// Validation of nonblank entries remains the responsibility of the public
// adapters, as it is for the existing initialism-aware transformations.
func newTypeScriptInitialismCollection(initialisms []string) initialismCollection {
	collection := newInitialismCollection(initialisms)
	for key, spelling := range collection {
		var sanitized strings.Builder
		for _, r := range spelling {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sanitized.WriteRune(r)
			}
		}

		if sanitized.Len() == 0 {
			delete(collection, key)
			continue
		}
		collection[key] = sanitized.String()
	}

	return collection
}

// typeScriptReservedBindings contains ECMAScript keywords, strict/future
// reserved words, literals that cannot be binding names, and the two names
// forbidden as bindings in strict mode. Contextual words such as async, from,
// get, of, and set remain valid and are intentionally absent.
var typeScriptReservedBindings = map[string]struct{}{
	"arguments":  {},
	"await":      {},
	"break":      {},
	"case":       {},
	"catch":      {},
	"class":      {},
	"const":      {},
	"continue":   {},
	"debugger":   {},
	"default":    {},
	"delete":     {},
	"do":         {},
	"else":       {},
	"enum":       {},
	"eval":       {},
	"export":     {},
	"extends":    {},
	"false":      {},
	"finally":    {},
	"for":        {},
	"function":   {},
	"if":         {},
	"implements": {},
	"import":     {},
	"in":         {},
	"instanceof": {},
	"interface":  {},
	"let":        {},
	"new":        {},
	"null":       {},
	"package":    {},
	"private":    {},
	"protected":  {},
	"public":     {},
	"return":     {},
	"static":     {},
	"super":      {},
	"switch":     {},
	"this":       {},
	"throw":      {},
	"true":       {},
	"try":        {},
	"typeof":     {},
	"var":        {},
	"void":       {},
	"while":      {},
	"with":       {},
	"yield":      {},
}

func isTypeScriptReservedBinding(identifier string) bool {
	_, reserved := typeScriptReservedBindings[identifier]
	return reserved
}
