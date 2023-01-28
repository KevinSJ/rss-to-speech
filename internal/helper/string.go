package helper

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var table = []*unicode.RangeTable{
	unicode.Pf,
	unicode.Sc,
	unicode.Number,
}

func GetSanitizedLangCode(s string) string {
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
