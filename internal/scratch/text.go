package scratch

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"slices"
	"unicode/utf16"
)

// Lower owns its caser on every call: Unicode casing transformers may hold state.
func Lower(s string) string { return cases.Lower(language.Und).String(s) }

// CompareText orders lowercase JavaScript strings by UTF-16 code units.
func CompareText(left, right string) int {
	return slices.Compare(
		utf16.Encode([]rune(Lower(left))),
		utf16.Encode([]rune(Lower(right))),
	)
}
