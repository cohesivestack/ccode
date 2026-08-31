package ccode

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestStringToTypeScriptIdentifiers(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		initialisms []string
		typeName    string
		valueName   string
	}{
		{name: "words", input: "user id", typeName: "UserId", valueName: "userId"},
		{name: "default acronym normalization", input: "HTTP response", typeName: "HttpResponse", valueName: "httpResponse"},
		{name: "hyphen separator", input: "customer-account", typeName: "CustomerAccount", valueName: "customerAccount"},
		{name: "unsupported punctuation separator", input: "customer@account", typeName: "CustomerAccount", valueName: "customerAccount"},
		{name: "multiple unsupported separators", input: "customer::account!status", typeName: "CustomerAccountStatus", valueName: "customerAccountStatus"},
		{name: "explicit initialisms", input: "api response id", initialisms: []string{"API", "ID"}, typeName: "APIResponseID", valueName: "apiResponseID"},
		{name: "mixed case initialism", input: "graph ql client", initialisms: []string{"GraphQL"}, typeName: "GraphQLClient", valueName: "graphQLClient"},
		{name: "leading digits", input: "123 users", typeName: "T123Users", valueName: "_123Users"},
		{name: "empty", input: "", typeName: "", valueName: ""},
		{name: "separator only", input: "---", typeName: "", valueName: ""},
		{name: "whitespace only", input: "   ", typeName: "", valueName: ""},
		{name: "Unicode letters", input: "élève compte", typeName: "ÉlèveCompte", valueName: "élèveCompte"},
		{name: "reserved class source", input: "class", typeName: "Class", valueName: "class_"},
		{name: "reserved default source", input: "default", typeName: "Default", valueName: "default_"},
		{name: "punctuated initialism is sanitized", input: "api response", initialisms: []string{"A.P.I"}, typeName: "APIResponse", valueName: "apiResponse"},
		{name: "punctuation-only initialism is ignored", input: "value", initialisms: []string{"!!!"}, typeName: "Value", valueName: "value"},
		{name: "custom lowercase reserved type", input: "class", initialisms: []string{"class"}, typeName: "class_", valueName: "class_"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append([]string(nil), test.initialisms...)

			typeName := stringToTypeScriptTypeIdentifier(test.input, test.initialisms)
			valueName := stringToTypeScriptValueIdentifier(test.input, test.initialisms)

			assert.Equal(t, test.typeName, typeName)
			assert.Equal(t, test.valueName, valueName)
			assert.Equal(t, original, test.initialisms)
			assertValidConservativeTypeScriptIdentifier(t, typeName)
			assertValidConservativeTypeScriptIdentifier(t, valueName)
		})
	}
}

func TestStringToTypeScriptValueIdentifierReservedBindings(t *testing.T) {
	reserved := []string{
		"class", "default", "function", "import", "await", "yield",
		"arguments", "eval", "enum", "implements", "interface", "let",
		"package", "private", "protected", "public", "static", "null",
		"true", "false",
	}
	for _, input := range reserved {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, input+"_", stringToTypeScriptValueIdentifier(input, nil))
		})
	}
}

func TestStringToTypeScriptValueIdentifierAllowsContextualBindings(t *testing.T) {
	for _, input := range []string{"async", "from", "get", "meta", "of", "set", "target"} {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, input, stringToTypeScriptValueIdentifier(input, nil))
		})
	}
}

func assertValidConservativeTypeScriptIdentifier(t *testing.T, identifier string) {
	t.Helper()
	if identifier == "" {
		return
	}

	first, size := utf8.DecodeRuneInString(identifier)
	assert.True(t, first == '_' || unicode.IsLetter(first), "invalid first rune in %q", identifier)
	for _, r := range identifier[size:] {
		assert.True(t, r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r), "invalid rune %q in %q", r, identifier)
	}
	assert.False(t, isTypeScriptReservedBinding(identifier), "reserved identifier %q", identifier)
}
