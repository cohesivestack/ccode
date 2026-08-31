package ccode

import "strings"

func openAPIPathToColon(path string, omitLeadingSlash bool) string {
	return transformOpenAPIPath(path, omitLeadingSlash, func(parameter string) string {
		return ":" + parameter
	})
}

func openAPIPathToSquareBrackets(path string, omitLeadingSlash bool) string {
	return transformOpenAPIPath(path, omitLeadingSlash, func(parameter string) string {
		return "[" + parameter + "]"
	})
}

func openAPIPathToAngleBrackets(path string, omitLeadingSlash bool) string {
	return transformOpenAPIPath(path, omitLeadingSlash, func(parameter string) string {
		return "<" + parameter + ">"
	})
}

func openAPIPathToDollar(path string, omitLeadingSlash bool) string {
	return transformOpenAPIPath(path, omitLeadingSlash, func(parameter string) string {
		return "$" + parameter
	})
}

func transformOpenAPIPath(path string, omitLeadingSlash bool, formatParameter func(string) string) string {
	if omitLeadingSlash {
		path = strings.TrimPrefix(path, "/")
	}

	var result strings.Builder
	for start := 0; start < len(path); {
		openingOffset := strings.IndexByte(path[start:], '{')
		if openingOffset == -1 {
			result.WriteString(path[start:])
			break
		}

		opening := start + openingOffset
		result.WriteString(path[start:opening])

		closingOffset := strings.IndexByte(path[opening+1:], '}')
		if closingOffset == -1 {
			result.WriteString(path[opening:])
			break
		}

		closing := opening + 1 + closingOffset
		parameter := path[opening+1 : closing]
		if parameter == "" || strings.ContainsRune(parameter, '{') {
			result.WriteString(path[opening : closing+1])
		} else {
			result.WriteString(formatParameter(parameter))
		}
		start = closing + 1
	}

	return result.String()
}
