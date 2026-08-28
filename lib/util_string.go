package ccode

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type initialismCollection map[string]string

type resolvedWord struct {
	text       string
	initialism bool
}

// stringToCamelCase converts a string to camelCase.
func stringToCamelCase(s string, initialisms []string) string {
	if s == "" {
		return s
	}

	effectiveInitialisms := newInitialismCollection(initialisms)
	return arrayToCamelCase(splitStringIntoWords(s), effectiveInitialisms)
}

func arrayToCamelCase(words []string, initialisms initialismCollection) string {
	resolved := resolveInitialisms(words, initialisms)
	if len(resolved) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(lowerCamelWord(resolved[0]))
	for _, word := range resolved[1:] {
		result.WriteString(pascalWord(word))
	}

	return result.String()
}

// stringToPascalCase converts a string to PascalCase.
func stringToPascalCase(s string, initialisms []string) string {
	if s == "" {
		return s
	}

	effectiveInitialisms := newInitialismCollection(initialisms)
	return arrayToPascalCase(splitStringIntoWords(s), effectiveInitialisms)
}

func arrayToPascalCase(words []string, initialisms initialismCollection) string {
	var result strings.Builder
	for _, word := range resolveInitialisms(words, initialisms) {
		result.WriteString(pascalWord(word))
	}

	return result.String()
}

// stringToSnakeCase converts a string to snake_case.
func stringToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	return arrayToSnakeCase(splitStringIntoWords(s))
}

func arrayToSnakeCase(words []string) string {
	return arrayToLowerSeparatedCase(words, "_")
}

// stringToKebabCase converts a string to kebab-case.
func stringToKebabCase(s string) string {
	if s == "" {
		return s
	}

	return arrayToKebabCase(splitStringIntoWords(s))
}

func arrayToKebabCase(words []string) string {
	return arrayToLowerSeparatedCase(words, "-")
}

// stringToConstantCase converts a string to CONSTANT_CASE.
func stringToConstantCase(s string) string {
	if s == "" {
		return s
	}

	return arrayToConstantCase(splitStringIntoWords(s))
}

func arrayToConstantCase(words []string) string {
	words = normalizedWords(words)
	for i := range words {
		words[i] = strings.ToUpper(words[i])
	}

	return strings.Join(words, "_")
}

// stringToDotCase converts a string to dot.case.
func stringToDotCase(s string) string {
	if s == "" {
		return s
	}

	return arrayToDotCase(splitStringIntoWords(s))
}

func arrayToDotCase(words []string) string {
	return arrayToLowerSeparatedCase(words, ".")
}

// stringToPathCase converts a string to path/case.
func stringToPathCase(s string) string {
	if s == "" {
		return s
	}

	return arrayToPathCase(splitStringIntoWords(s))
}

func arrayToPathCase(words []string) string {
	return arrayToLowerSeparatedCase(words, "/")
}

// stringToTitleCase converts a string to Title Case.
func stringToTitleCase(s string, initialisms []string) string {
	if s == "" {
		return s
	}

	effectiveInitialisms := newInitialismCollection(initialisms)
	return arrayToTitleCase(splitStringIntoWords(s), effectiveInitialisms)
}

func arrayToTitleCase(words []string, initialisms initialismCollection) string {
	resolved := resolveInitialisms(words, initialisms)
	formatted := make([]string, 0, len(resolved))
	for _, word := range resolved {
		formatted = append(formatted, pascalWord(word))
	}

	return strings.Join(formatted, " ")
}

// stringToSentenceCase converts a string to Sentence case.
func stringToSentenceCase(s string, initialisms []string) string {
	if s == "" {
		return s
	}

	effectiveInitialisms := newInitialismCollection(initialisms)
	return arrayToSentenceCase(splitStringIntoWords(s), effectiveInitialisms)
}

func arrayToSentenceCase(words []string, initialisms initialismCollection) string {
	resolved := resolveInitialisms(words, initialisms)
	if len(resolved) == 0 {
		return ""
	}

	formatted := make([]string, 0, len(resolved))
	formatted = append(formatted, pascalWord(resolved[0]))
	for _, word := range resolved[1:] {
		formatted = append(formatted, word.text)
	}

	return strings.Join(formatted, " ")
}

// stringToUpperFirst uppercases the first rune and preserves the remainder.
func stringToUpperFirst(s string) string {
	if s == "" {
		return s
	}

	_, size := utf8.DecodeRuneInString(s)
	return cases.Upper(language.Und).String(s[:size]) + s[size:]
}

// stringToLowerFirst lowercases the first rune and preserves the remainder.
func stringToLowerFirst(s string) string {
	if s == "" {
		return s
	}

	_, size := utf8.DecodeRuneInString(s)
	return cases.Lower(language.Und).String(s[:size]) + s[size:]
}

// stringToNormalizeSpace trims surrounding whitespace and replaces each run of
// internal whitespace with a single ASCII space.
func stringToNormalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func arrayToLowerSeparatedCase(words []string, separator string) string {
	return strings.Join(normalizedWords(words), separator)
}

func normalizedWords(words []string) []string {
	normalized := make([]string, 0, len(words))
	for _, word := range words {
		if word != "" {
			normalized = append(normalized, strings.ToLower(word))
		}
	}

	return normalized
}

func titleWord(word string) string {
	return cases.Title(language.Und, cases.NoLower).String(strings.ToLower(word))
}

func newInitialismCollection(initialismGroups ...[]string) initialismCollection {
	collection := make(initialismCollection)
	for _, initialisms := range initialismGroups {
		for _, initialism := range initialisms {
			if isStringBlank(initialism) {
				continue
			}

			key := initialismKey(initialism)
			if key != "" {
				collection[key] = initialism
			}
		}
	}

	return collection
}

func initialismKey(s string) string {
	return strings.ToLower(strings.Join(splitStringIntoWords(s), ""))
}

func resolveInitialisms(words []string, initialisms initialismCollection) []resolvedWord {
	compact := make([]string, 0, len(words))
	for _, word := range words {
		if word != "" {
			compact = append(compact, word)
		}
	}

	resolved := make([]resolvedWord, 0, len(compact))
	for start := 0; start < len(compact); {
		var candidate strings.Builder
		matchedEnd := start
		matchedSpelling := ""
		for end := start; end < len(compact); end++ {
			candidate.WriteString(initialismKey(compact[end]))
			if spelling, ok := initialisms[candidate.String()]; ok {
				matchedEnd = end + 1
				matchedSpelling = spelling
			}
		}

		if matchedSpelling != "" {
			resolved = append(resolved, resolvedWord{text: matchedSpelling, initialism: true})
			start = matchedEnd
			continue
		}

		resolved = append(resolved, resolvedWord{text: strings.ToLower(compact[start])})
		start++
	}

	return resolved
}

func pascalWord(word resolvedWord) string {
	if word.initialism {
		return word.text
	}

	return titleWord(word.text)
}

func lowerCamelWord(word resolvedWord) string {
	if !word.initialism {
		return strings.ToLower(word.text)
	}
	if word.text == strings.ToUpper(word.text) {
		return strings.ToLower(word.text)
	}

	return stringToLowerFirst(word.text)
}

// splitStringIntoWords splits a string into words based on separators and
// lower-to-upper or acronym-to-word case boundaries.
func splitStringIntoWords(s string) []string {
	if s == "" {
		return []string{}
	}

	runes := []rune(s)
	words := make([]string, 0)
	var currentWord strings.Builder

	flushCurrentWord := func() {
		if currentWord.Len() > 0 {
			words = append(words, currentWord.String())
			currentWord.Reset()
		}
	}

	for i, r := range runes {
		if isSeparator(r) {
			flushCurrentWord()
			continue
		}

		if unicode.IsUpper(r) && currentWord.Len() > 0 {
			previousIsLowerOrDigit := unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])
			startsWordAfterAcronym := unicode.IsUpper(runes[i-1]) &&
				i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if previousIsLowerOrDigit || startsWordAfterAcronym {
				flushCurrentWord()
			}
		}

		currentWord.WriteRune(r)
	}

	flushCurrentWord()
	return words
}

// isSeparator checks if a rune is a word separator.
func isSeparator(r rune) bool {
	return unicode.IsSpace(r) || r == '_' || r == '-' || r == '.' || r == '/'
}

func isStringBlank[T ~string](s T) bool {
	return len(strings.TrimSpace(string(s))) == 0
}
