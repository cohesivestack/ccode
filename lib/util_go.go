package ccode

import (
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

// stringToGoExported converts a string to an exported Go identifier.
func stringToGoExported(s string, initialisms []string) string {
	effectiveInitialisms := newInitialismCollection(commonTechnicalInitialisms, initialisms)
	words := splitStringIntoGoWords(s)
	identifier := arrayToGoExported(words, effectiveInitialisms)
	if identifier == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(identifier)
	if !unicode.IsUpper(first) {
		identifier = "X" + identifier
	}

	return identifier
}

func arrayToGoExported(words []string, initialisms initialismCollection) string {
	words = mergeGoPluralInitialismWords(words, initialisms)

	var result strings.Builder
	for _, word := range resolveInitialisms(words, initialisms) {
		if word.initialism {
			result.WriteString(word.text)
		} else {
			result.WriteString(goExportedWord(word.text, initialisms))
		}
	}

	return result.String()
}

// stringToGoUnexported converts a string to an unexported Go identifier.
func stringToGoUnexported(s string, initialisms []string) string {
	effectiveInitialisms := newInitialismCollection(commonTechnicalInitialisms, initialisms)
	words := splitStringIntoGoWords(s)
	identifier := arrayToGoUnexported(words, effectiveInitialisms)
	if identifier == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(identifier)
	if !unicode.IsLower(first) {
		identifier = "x" + identifier
	}
	if isGoKeyword(identifier) {
		identifier += "_"
	}

	return identifier
}

func arrayToGoUnexported(words []string, initialisms initialismCollection) string {
	words = mergeGoPluralInitialismWords(words, initialisms)
	resolved := resolveInitialisms(words, initialisms)
	if len(resolved) == 0 {
		return ""
	}

	var result strings.Builder
	if resolved[0].initialism {
		result.WriteString(lowerCamelWord(resolved[0]))
	} else {
		result.WriteString(goUnexportedWord(resolved[0].text, initialisms))
	}
	for _, word := range resolved[1:] {
		if word.initialism {
			result.WriteString(word.text)
		} else {
			result.WriteString(goExportedWord(word.text, initialisms))
		}
	}

	return result.String()
}

// stringToGoPackage converts a string to a conventional Go package name.
func stringToGoPackage(s string) string {
	words := splitStringIntoGoWords(s)
	packageName := strings.ReplaceAll(arrayToSnakeCase(words), "_", "")
	if packageName == "" {
		return ""
	}

	first, _ := utf8.DecodeRuneInString(packageName)
	if !unicode.IsLetter(first) {
		packageName = "pkg" + packageName
	}
	if isGoKeyword(packageName) {
		packageName += "pkg"
	}

	return packageName
}

func splitStringIntoGoWords(s string) []string {
	var normalized strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			normalized.WriteRune(r)
		} else {
			normalized.WriteRune(' ')
		}
	}

	return splitStringIntoWords(normalized.String())
}

func goExportedWord(word string, initialisms initialismCollection) string {
	if initialism, ok := matchInitialism(word, initialisms); ok {
		return initialism
	}
	if initialism, ok := goPluralInitialism(word, initialisms); ok {
		return initialism + "s"
	}

	return titleWord(word)
}

func goUnexportedWord(word string, initialisms initialismCollection) string {
	if initialism, ok := matchInitialism(word, initialisms); ok {
		return lowerCamelWord(resolvedWord{text: initialism, initialism: true})
	}
	if initialism, ok := goPluralInitialism(word, initialisms); ok {
		return lowerCamelWord(resolvedWord{text: initialism, initialism: true}) + "s"
	}

	return strings.ToLower(word)
}

func mergeGoPluralInitialismWords(words []string, initialisms initialismCollection) []string {
	merged := make([]string, 0, len(words))
	for start := 0; start < len(words); {
		candidate := words[start]
		matchedEnd := start + 1
		matchedCandidate := ""
		for end := start + 1; end < len(words); end++ {
			candidate += words[end]
			if _, ok := goPluralInitialism(candidate, initialisms); ok {
				matchedEnd = end + 1
				matchedCandidate = candidate
			}
		}

		if matchedCandidate != "" {
			merged = append(merged, matchedCandidate)
			start = matchedEnd
			continue
		}

		merged = append(merged, words[start])
		start++
	}

	return merged
}

func goPluralInitialism(word string, initialisms initialismCollection) (string, bool) {
	if _, ok := matchInitialism(word, initialisms); ok {
		return "", false
	}

	runes := []rune(word)
	if len(runes) < 2 || (runes[len(runes)-1] != 's' && runes[len(runes)-1] != 'S') {
		return "", false
	}

	return matchInitialism(string(runes[:len(runes)-1]), initialisms)
}

func matchInitialism(s string, initialisms initialismCollection) (string, bool) {
	initialism, ok := initialisms[initialismKey(s)]
	return initialism, ok
}

func isGoKeyword(s string) bool {
	return token.Lookup(s).IsKeyword()
}
