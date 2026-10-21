package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathTransformations(t *testing.T) {
	tests := []struct {
		name             string
		transform        func(string, bool) string
		path             string
		omitLeadingSlash bool
		expected         string
	}{
		{name: "colon", transform: pathToColon, path: "/users/{userId}/orders/{orderId}", expected: "/users/:userId/orders/:orderId"},
		{name: "square brackets", transform: pathToSquareBrackets, path: "/users/{userId}/orders/{orderId}", expected: "/users/[userId]/orders/[orderId]"},
		{name: "angle brackets", transform: pathToAngleBrackets, path: "/users/{userId}/orders/{orderId}", expected: "/users/<userId>/orders/<orderId>"},
		{name: "dollar", transform: pathToDollar, path: "/users/{userId}/orders/{orderId}", expected: "/users/$userId/orders/$orderId"},
		{name: "repeated parameter", transform: pathToColon, path: "/users/{id}/related/{id}", expected: "/users/:id/related/:id"},
		{name: "static path", transform: pathToSquareBrackets, path: "/users/current", expected: "/users/current"},
		{name: "root path", transform: pathToAngleBrackets, path: "/", expected: "/"},
		{name: "root path without leading slash", transform: pathToAngleBrackets, path: "/", omitLeadingSlash: true, expected: ""},
		{name: "empty path", transform: pathToDollar, path: "", expected: ""},
		{name: "Unicode parameter", transform: pathToColon, path: "/users/{用户编号}", expected: "/users/:用户编号"},
		{name: "omit leading slash", transform: pathToColon, path: "/users/{userId}", omitLeadingSlash: true, expected: "users/:userId"},
		{name: "remove exactly one leading slash", transform: pathToColon, path: "//users/{userId}", omitLeadingSlash: true, expected: "/users/:userId"},
		{name: "no leading slash to omit", transform: pathToDollar, path: "users/{userId}", omitLeadingSlash: true, expected: "users/$userId"},
		{name: "empty parameter", transform: pathToColon, path: "/users/{}", expected: "/users/{}"},
		{name: "incomplete parameter", transform: pathToColon, path: "/users/{userId", expected: "/users/{userId"},
		{name: "closing brace only", transform: pathToColon, path: "/users/userId}", expected: "/users/userId}"},
		{name: "nested opening brace", transform: pathToColon, path: "/users/{{userId}", expected: "/users/{{userId}"},
		{name: "valid after malformed", transform: pathToSquareBrackets, path: "/users/{}/orders/{orderId}", expected: "/users/{}/orders/[orderId]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.transform(test.path, test.omitLeadingSlash))
		})
	}
}

// Every syntax must be accepted by every formatter, including the formatter
// for its own syntax. Use literal expectations to catch parser/formatter errors.
func TestPathTransformationsBetweenAllSyntaxes(t *testing.T) {
	syntaxes := []struct {
		name      string
		transform func(string, bool) string
		path      string
	}{
		{name: "curly", transform: pathToCurlyBrackets, path: "/users/{userId}/files/{file}.{extension}"},
		{name: "colon", transform: pathToColon, path: "/users/:userId/files/:file.:extension"},
		{name: "square", transform: pathToSquareBrackets, path: "/users/[userId]/files/[file].[extension]"},
		{name: "angle", transform: pathToAngleBrackets, path: "/users/<userId>/files/<file>.<extension>"},
		{name: "dollar", transform: pathToDollar, path: "/users/$userId/files/$file.$extension"},
	}
	for _, source := range syntaxes {
		for _, target := range syntaxes {
			t.Run(source.name+" to "+target.name, func(t *testing.T) {
				assert.Equal(t, target.path, target.transform(source.path, false))
				assert.Equal(t, target.path[1:], target.transform(source.path, true))
			})
		}
	}
}

func TestPathParameterRecognition(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{name: "mixed syntax", path: "/{team}/:user/[order]/<item>/$revision", expected: "/{team}/{user}/{order}/{item}/{revision}"},
		{name: "Unicode", path: "/:用户编号/[cafe\u0301]/<$用户>/$用户", expected: "/{用户编号}/{cafe\u0301}/<$用户>/{用户}"},
		{name: "underscores digits hyphens", path: "/:_user2/$user-id/[user_id]", expected: "/{_user2}/{user-id}/{user_id}"},
		{name: "embedded parameters", path: "/prefix-:id/$file.:extension", expected: "/prefix-{id}/{file}.{extension}"},
		{name: "adjacent parameters", path: "/:first$second/[third]<fourth>", expected: "/{first}{second}/{third}{fourth}"},
		{name: "static Unicode", path: "/用户/current/%7Bid%7D", expected: "/用户/current/%7Bid%7D"},
		{name: "empty parameters", path: "/{}/[]/<>/:/$", expected: "/{}/[]/<>/:/$"},
		{name: "malformed groups", path: "/{{id}/[[id]/<<id>/{:id}/[$id]", expected: "/{{id}/[[id]/<<id>/{:id}/[$id]"},
		{name: "mismatched delimiters", path: "/{id]/[id>/<id}", expected: "/{id]/[id>/<id}"},
		{name: "do not cross segments", path: "/{bad/name}/[bad/name]/<bad/name>", expected: "/{bad/name}/[bad/name]/<bad/name>"},
		{name: "recover after incomplete group", path: "/{bad/:id/[order]", expected: "/{bad/{id}/{order}"},
		{name: "invalid names", path: "/{a b}/[a.b]/<a:b>/{$id}", expected: "/{a b}/[a.b]/<a:b>/{$id}"},
		{name: "catch-all and optional groups", path: "/[...slug]/[[id]]/[[...rest]]", expected: "/[...slug]/[[id]]/[[...rest]]"},
		{name: "invalid prefixes", path: "/:/$/::/$$/123/:123/$123", expected: "/:/$/::/$$/123/:123/$123"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, pathToCurlyBrackets(test.path, false))
		})
	}
}
