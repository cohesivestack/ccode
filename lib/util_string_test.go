package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericStringCaseTransforms(t *testing.T) {
	tests := []struct {
		name      string
		transform func(string) string
		expected  string
	}{
		{name: "camel case", transform: func(s string) string { return stringToCamelCase(s, nil) }, expected: "httpServerConfig"},
		{name: "pascal case", transform: func(s string) string { return stringToPascalCase(s, nil) }, expected: "HttpServerConfig"},
		{name: "snake case", transform: stringToSnakeCase, expected: "http_server_config"},
		{name: "kebab case", transform: stringToKebabCase, expected: "http-server-config"},
		{name: "constant case", transform: stringToConstantCase, expected: "HTTP_SERVER_CONFIG"},
		{name: "dot case", transform: stringToDotCase, expected: "http.server.config"},
		{name: "path case", transform: stringToPathCase, expected: "http/server/config"},
		{name: "title case", transform: func(s string) string { return stringToTitleCase(s, nil) }, expected: "Http Server Config"},
		{name: "sentence case", transform: func(s string) string { return stringToSentenceCase(s, nil) }, expected: "Http server config"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.transform("HTTPServer_config"))
			assert.Empty(t, test.transform(""))
		})
	}
}

func TestGenericInitialismCaseTransforms(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		initialisms []string
		camel       string
		pascal      string
		title       string
		sentence    string
	}{
		{
			name: "neutral nil", input: "HTTP server ID",
			camel: "httpServerId", pascal: "HttpServerId",
			title: "Http Server Id", sentence: "Http server id",
		},
		{
			name: "neutral empty", input: "HTTP server ID", initialisms: []string{},
			camel: "httpServerId", pascal: "HttpServerId",
			title: "Http Server Id", sentence: "Http server id",
		},
		{
			name: "uppercase first and subsequent", input: "HTTP server ID", initialisms: []string{"HTTP", "ID"},
			camel: "httpServerID", pascal: "HTTPServerID",
			title: "HTTP Server ID", sentence: "HTTP server ID",
		},
		{
			name: "mixed case first", input: "OAuth client", initialisms: []string{"OAuth"},
			camel: "oAuthClient", pascal: "OAuthClient",
			title: "OAuth Client", sentence: "OAuth client",
		},
		{
			name: "mixed case sequence", input: "graph_ql OpenAPI oauth", initialisms: []string{"GraphQL", "OpenAPI", "OAuth"},
			camel: "graphQLOpenAPIOAuth", pascal: "GraphQLOpenAPIOAuth",
			title: "GraphQL OpenAPI OAuth", sentence: "GraphQL OpenAPI OAuth",
		},
		{
			name: "longest match", input: "open_api client", initialisms: []string{"API", "OpenAPI"},
			camel: "openAPIClient", pascal: "OpenAPIClient",
			title: "OpenAPI Client", sentence: "OpenAPI client",
		},
		{
			name: "case insensitive last spelling wins", input: "http id", initialisms: []string{"HTTP", "Id", "id", "iD"},
			camel: "httpiD", pascal: "HTTPiD",
			title: "HTTP iD", sentence: "HTTP iD",
		},
		{
			name: "blank entries and substring prevention", input: "valid id", initialisms: []string{"", "  ", "ID"},
			camel: "validID", pascal: "ValidID",
			title: "Valid ID", sentence: "Valid ID",
		},
		{
			name: "different separators", input: "customer/open-api.id", initialisms: []string{"OpenAPI", "ID"},
			camel: "customerOpenAPIID", pascal: "CustomerOpenAPIID",
			title: "Customer OpenAPI ID", sentence: "Customer OpenAPI ID",
		},
		{
			name: "already camel cased", input: "customerOpenAPIClient", initialisms: []string{"OpenAPI"},
			camel: "customerOpenAPIClient", pascal: "CustomerOpenAPIClient",
			title: "Customer OpenAPI Client", sentence: "Customer OpenAPI client",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := append(test.initialisms[:0:0], test.initialisms...)
			assert.Equal(t, test.camel, stringToCamelCase(test.input, test.initialisms))
			assert.Equal(t, test.pascal, stringToPascalCase(test.input, test.initialisms))
			assert.Equal(t, test.title, stringToTitleCase(test.input, test.initialisms))
			assert.Equal(t, test.sentence, stringToSentenceCase(test.input, test.initialisms))
			assert.Equal(t, original, test.initialisms)
		})
	}

	assert.Empty(t, stringToCamelCase("", []string{"HTTP"}))
	assert.Empty(t, stringToPascalCase("", []string{"HTTP"}))
	assert.Empty(t, stringToTitleCase("", []string{"HTTP"}))
	assert.Empty(t, stringToSentenceCase("", []string{"HTTP"}))
}

func TestStringFirstRuneTransforms(t *testing.T) {
	assert.Equal(t, "UserAccount", stringToUpperFirst("userAccount"))
	assert.Equal(t, "userAccount", stringToLowerFirst("UserAccount"))
	assert.Equal(t, "Éclair", stringToUpperFirst("éclair"))
	assert.Equal(t, "éclair", stringToLowerFirst("Éclair"))
	assert.Empty(t, stringToUpperFirst(""))
	assert.Empty(t, stringToLowerFirst(""))
}

func TestStringToNormalizeSpace(t *testing.T) {
	assert.Equal(t, "user account name", stringToNormalizeSpace("  user\t account\nname  "))
	assert.Equal(t, "userAccount", stringToNormalizeSpace(" userAccount "))
	assert.Empty(t, stringToNormalizeSpace(" \t\n "))
}

func TestSplitStringIntoWords(t *testing.T) {
	assert.Equal(t, []string{"HTTP", "Server", "config"}, splitStringIntoWords("HTTPServer_config"))
	assert.Equal(t, []string{"user", "account"}, splitStringIntoWords(" user...account "))
	assert.Empty(t, splitStringIntoWords(""))
}
