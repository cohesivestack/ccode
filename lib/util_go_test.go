package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToGoExported(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		initialisms []string
		expected    string
	}{
		{name: "default ID", input: "user id", expected: "UserID"},
		{name: "default HTTP", input: "HTTP server", expected: "HTTPServer"},
		{name: "custom DB", input: "database db", initialisms: []string{"DB"}, expected: "DatabaseDB"},
		{name: "custom mixed case", input: "oauth client", initialisms: []string{"OAuth"}, expected: "OAuthClient"},
		{name: "custom adjacent sequence", input: "graph ql client", initialisms: []string{"GraphQL"}, expected: "GraphQLClient"},
		{name: "override default casing", input: "user ids", initialisms: []string{"Id"}, expected: "UserIds"},
		{name: "default plural", input: "user ids", expected: "UserIDs"},
		{name: "custom plural", input: "database dbs", initialisms: []string{"DB"}, expected: "DatabaseDBs"},
		{name: "direct match before plural", input: "database dbs", initialisms: []string{"DBS"}, expected: "DatabaseDBS"},
		{name: "both direct and singular", input: "database dbs", initialisms: []string{"DB", "DBS"}, expected: "DatabaseDBS"},
		{name: "remove one suffix only", input: "database dbss", initialisms: []string{"DB"}, expected: "DatabaseDbss"},
		{name: "plural of direct term", input: "database dbss", initialisms: []string{"DBS"}, expected: "DatabaseDBSs"},
		{name: "camel split plural", input: "dbsGet", initialisms: []string{"DB"}, expected: "DBsGet"},
		{name: "camel split not recursive plural", input: "dbssGet", initialisms: []string{"DB"}, expected: "DbssGet"},
		{name: "separate plural suffix", input: "user id s", expected: "UserIDs"},
		{name: "digit prefix", input: "123 users", expected: "X123Users"},
		{name: "empty", input: "", expected: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append([]string(nil), test.initialisms...)
			assert.Equal(t, test.expected, stringToGoExported(test.input, test.initialisms))
			assert.Equal(t, original, test.initialisms)
		})
	}
}

func TestStringToGoUnexported(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		initialisms []string
		expected    string
	}{
		{name: "default first initialism", input: "URL parser", expected: "urlParser"},
		{name: "default subsequent initialism", input: "get user ID", expected: "getUserID"},
		{name: "mixed case first", input: "OAuth client", initialisms: []string{"OAuth"}, expected: "oAuthClient"},
		{name: "mixed case subsequent", input: "customer OAuth ID", initialisms: []string{"OAuth"}, expected: "customerOAuthID"},
		{name: "first plural", input: "ids", expected: "ids"},
		{name: "subsequent plural", input: "user ids", expected: "userIDs"},
		{name: "custom subsequent plural", input: "database dbs", initialisms: []string{"DB"}, expected: "databaseDBs"},
		{name: "keyword", input: "type", expected: "type_"},
		{name: "digit prefix", input: "123 users", expected: "x123Users"},
		{name: "empty", input: "", expected: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, stringToGoUnexported(test.input, test.initialisms))
		})
	}
}

func TestStringToGoPackageUnaffectedByInitialisms(t *testing.T) {
	assert.Equal(t, "httpserver", stringToGoPackage("HTTP server"))
	assert.Equal(t, "typepkg", stringToGoPackage("type"))
	assert.Equal(t, "pkg123api", stringToGoPackage("123 API"))
	assert.Empty(t, stringToGoPackage(""))
}
