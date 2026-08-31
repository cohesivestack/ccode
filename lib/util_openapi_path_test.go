package ccode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenAPIPathTransformations(t *testing.T) {
	tests := []struct {
		name             string
		transform        func(string, bool) string
		path             string
		omitLeadingSlash bool
		expected         string
	}{
		{name: "colon", transform: openAPIPathToColon, path: "/users/{userId}/orders/{orderId}", expected: "/users/:userId/orders/:orderId"},
		{name: "square brackets", transform: openAPIPathToSquareBrackets, path: "/users/{userId}/orders/{orderId}", expected: "/users/[userId]/orders/[orderId]"},
		{name: "angle brackets", transform: openAPIPathToAngleBrackets, path: "/users/{userId}/orders/{orderId}", expected: "/users/<userId>/orders/<orderId>"},
		{name: "dollar", transform: openAPIPathToDollar, path: "/users/{userId}/orders/{orderId}", expected: "/users/$userId/orders/$orderId"},
		{name: "repeated parameter", transform: openAPIPathToColon, path: "/users/{id}/related/{id}", expected: "/users/:id/related/:id"},
		{name: "static path", transform: openAPIPathToSquareBrackets, path: "/users/current", expected: "/users/current"},
		{name: "root path", transform: openAPIPathToAngleBrackets, path: "/", expected: "/"},
		{name: "root path without leading slash", transform: openAPIPathToAngleBrackets, path: "/", omitLeadingSlash: true, expected: ""},
		{name: "empty path", transform: openAPIPathToDollar, path: "", expected: ""},
		{name: "Unicode parameter", transform: openAPIPathToColon, path: "/users/{用户编号}", expected: "/users/:用户编号"},
		{name: "omit leading slash", transform: openAPIPathToColon, path: "/users/{userId}", omitLeadingSlash: true, expected: "users/:userId"},
		{name: "remove exactly one leading slash", transform: openAPIPathToColon, path: "//users/{userId}", omitLeadingSlash: true, expected: "/users/:userId"},
		{name: "no leading slash to omit", transform: openAPIPathToDollar, path: "users/{userId}", omitLeadingSlash: true, expected: "users/$userId"},
		{name: "empty parameter", transform: openAPIPathToColon, path: "/users/{}", expected: "/users/{}"},
		{name: "incomplete parameter", transform: openAPIPathToColon, path: "/users/{userId", expected: "/users/{userId"},
		{name: "closing brace only", transform: openAPIPathToColon, path: "/users/userId}", expected: "/users/userId}"},
		{name: "nested opening brace", transform: openAPIPathToColon, path: "/users/{{userId}", expected: "/users/{{userId}"},
		{name: "valid after malformed", transform: openAPIPathToSquareBrackets, path: "/users/{}/orders/{orderId}", expected: "/users/{}/orders/[orderId]"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.transform(test.path, test.omitLeadingSlash))
		})
	}
}
