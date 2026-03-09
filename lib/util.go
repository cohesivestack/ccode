package ccode

import "strings"

func isStringBlank[T ~string](s T) bool {
	return len(strings.TrimSpace(string(s))) == 0
}
