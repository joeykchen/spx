package spx

import "testing"

func TestScratchUnicodeText(t *testing.T) {
	for _, tt := range []struct {
		left, right string
		want        int
	}{
		{"İ", "i\u0307", 0}, {"ΟΣ", "ος", 0}, {"Σ", "ς", 1},
		{"😀", "\ue000", -1}, {"a😀", "a\uffff", -1},
		{"a", "ab", -1}, {"中文", "中文", 0},
	} {
		if got := Compare(tt.left, tt.right); got != tt.want {
			t.Errorf("Compare(%q,%q)=%d want %d", tt.left, tt.right, got, tt.want)
		}
		list := NewList(tt.left)
		if got := list.Contains(tt.right); got != (tt.want == 0) {
			t.Errorf("list Contains(%q,%q)=%v", tt.left, tt.right, got)
		}
	}
	for _, tt := range []struct {
		text, part string
		want       bool
	}{
		{"İstanbul", "i\u0307s", true}, {"ΟΣ", "ος", true},
		{"ΟΣΑ", "οσα", true}, {"Σ", "ς", false}, {"hello", "", true},
	} {
		if got := Contains(tt.text, tt.part); got != tt.want {
			t.Errorf("Contains(%q,%q)=%v", tt.text, tt.part, got)
		}
	}
}
