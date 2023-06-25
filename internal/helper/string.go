package helper

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var JAPANESE_UNICODE_RANGE = []*unicode.RangeTable{
	unicode.Hiragana, // Hiragana is the set of Unicode characters in script Hiragana.
}

var CHINESE_UNICODE_RANGE = []*unicode.RangeTable{
	unicode.Han, // Han is the set of Unicode characters in script Han.
}

// Guess the language code for a string by looking at the unicode
func guessLanguageByUnicode(title string) string {
	for _, c := range title {
		if unicode.In(c, CHINESE_UNICODE_RANGE...) {
			return "zh-CN"
		}
	}
	return "en-US"
}

func getSanitizedLanguageCode(s string) string {
	s2 := strings.Split(s, "-")

	return s2[0] + "-" + strings.ToUpper(s2[len(s2)-1])
}

// returns the splited string by the size, chunkSize will be rounded to smallest
// int divisble by the rune size
func chunksByte(s string, chunkSize int) []string {
	if len(s) == 0 {
		return nil
	}

	perRuneSize := len(s) / utf8.RuneCountInString(s)

	if chunkSize <= perRuneSize || len(s) <= chunkSize {
		return []string{s}
	}
	currentLen, currentStart := 0, 0

	chunks := make([]string, 0)

	for i, ch := range s {
		if runeLen := utf8.RuneLen(ch); runeLen != -1 {
			currentLen += runeLen
			if currentLen > chunkSize {
				chunks = append(chunks, s[currentStart:i])
				currentLen = runeLen
				currentStart = i
			}
		}
	}

	chunks = append(chunks, s[currentStart:])

	return chunks
}
