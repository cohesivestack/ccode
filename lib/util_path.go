package ccode

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func pathToCurlyBrackets(path string, omitLeadingSlash bool) string {
	return transformPath(path, omitLeadingSlash, func(parameter string) string {
		return "{" + parameter + "}"
	})
}

func pathToColon(path string, omitLeadingSlash bool) string {
	return transformPath(path, omitLeadingSlash, func(parameter string) string {
		return ":" + parameter
	})
}

func pathToSquareBrackets(path string, omitLeadingSlash bool) string {
	return transformPath(path, omitLeadingSlash, func(parameter string) string {
		return "[" + parameter + "]"
	})
}

func pathToAngleBrackets(path string, omitLeadingSlash bool) string {
	return transformPath(path, omitLeadingSlash, func(parameter string) string {
		return "<" + parameter + ">"
	})
}

func pathToDollar(path string, omitLeadingSlash bool) string {
	return transformPath(path, omitLeadingSlash, func(parameter string) string {
		return "$" + parameter
	})
}

func transformPath(path string, omitLeadingSlash bool, formatParameter func(string) string) string {
	if omitLeadingSlash {
		path = strings.TrimPrefix(path, "/")
	}

	var result strings.Builder
	for start := 0; start < len(path); {
		switch path[start] {
		case '{', '[', '<':
			closing := byte('}')
			if path[start] == '[' {
				closing = ']'
			} else if path[start] == '<' {
				closing = '>'
			}

			// Never consume a later path segment while looking for a closing
			// delimiter. Preserve malformed groups without converting tokens inside.
			end := start + 1
			for end < len(path) && path[end] != closing && path[end] != '/' {
				end++
			}
			if end < len(path) && path[end] == closing {
				parameter := path[start+1 : end]
				if parameter != "" && pathParameterNameEnd(parameter) == len(parameter) {
					result.WriteString(formatParameter(parameter))
				} else {
					result.WriteString(path[start : end+1])
				}
				start = end + 1
			} else {
				result.WriteString(path[start:end])
				start = end
			}
		case ':', '$':
			end := start + 1 + pathParameterNameEnd(path[start+1:])
			if end > start+1 {
				result.WriteString(formatParameter(path[start+1 : end]))
			} else {
				result.WriteByte(path[start])
			}
			start = end
		default:
			result.WriteByte(path[start])
			start++
		}
	}

	return result.String()
}

// A shared name grammar makes parameters recognizable in every supported
// syntax. A dot terminates a name, allowing paths such as /:file.:extension.
func pathParameterNameEnd(value string) int {
	end := 0
	for end < len(value) {
		r, size := utf8.DecodeRuneInString(value[end:])
		if !unicode.IsLetter(r) && r != '_' &&
			(end == 0 || (!unicode.IsDigit(r) && !unicode.IsMark(r) && r != '-')) {
			break
		}
		end += size
	}
	return end
}
