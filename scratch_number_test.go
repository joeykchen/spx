package spx

import (
	"math"
	"testing"
)

func TestScratchComparisonNumericSyntax(t *testing.T) {
	for _, tt := range []struct {
		value  string
		number float64
		equal  bool
	}{
		{"0x10", 16, true}, {"0b10", 2, true}, {"0o10", 8, true}, {"  +2e0 ", 2, true},
		{"0x10000000000000000", math.Exp2(64), true}, {"1e999", math.Inf(1), true},
		{"--1", 1, false}, {"1_0", 10, false}, {"0x1p0", 1, false}, {"+0x1", 1, false},
		{"\\ufeff2", 2, false}, {"\ufeff2", 2, true}, {"\u00852", 2, false}, {"", 0, false},
	} {
		t.Run(tt.value, func(t *testing.T) {
			if got := Equal(tt.value, tt.number); got != tt.equal {
				t.Fatalf("Equal(%q,%g)=%v", tt.value, tt.number, got)
			}
			list := NewList(tt.value)
			if list.Contains(tt.number) != tt.equal {
				t.Fatal("list comparison differs")
			}
		})
	}
}
