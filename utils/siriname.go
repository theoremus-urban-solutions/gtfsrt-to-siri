package utils

import (
	"strings"
	"unicode"
)

// NMToken makes s a valid xsd:NMTOKEN, as SIRI codes need, by turning each other character into "_".
func NMToken(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || strings.ContainsRune(".-_:", r) {
			return r
		}
		return '_'
	}, s)
}

// PlaceName drops the characters SIRI place names forbid (, [ ] { } ? $ % ^ = @ # ; :) and tidies the spaces left behind.
func PlaceName(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(",[]{}?$%^=@#;:", r) {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}
