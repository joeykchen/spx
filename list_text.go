package spx

import (
	"strings"
	"unicode/utf8"
)

// String reports list contents using Scratch's type-sensitive separator rule.
func (p *List) String() string {
	sep := ""
	items := make([]string, len(p.data))
	for i, item := range p.data {
		val := toString(item)
		if !isScratchListLetter(item) {
			sep = " "
		}
		items[i] = val
	}
	return strings.Join(items, sep)
}

// Scratch's separator test uses the original string type and UTF-16 length.
func isScratchListLetter(value any) bool {
	text, ok := fromObj(value).(string)
	if !ok || text == "" {
		return false
	}
	r, size := utf8.DecodeRuneInString(text)
	return size == len(text) && r <= 0xffff
}
