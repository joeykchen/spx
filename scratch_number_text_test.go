package spx

import (
	"math"
	"testing"
)

func TestScratchNumberText(t *testing.T) {
	for _, tt := range []struct {
		number float64
		text   string
	}{
		{1e6, "1000000"}, {1e20, "100000000000000000000"}, {1e21, "1e+21"},
		{1e-6, "0.000001"}, {1e-7, "1e-7"}, {-1e-7, "-1e-7"},
		{math.Copysign(0, -1), "0"}, {math.Inf(1), "Infinity"}, {math.Inf(-1), "-Infinity"}, {math.NaN(), "NaN"},
	} {
		value := Value{data: tt.number}
		if got := value.String(); got != tt.text {
			t.Errorf("Value.String(%v)=%q want %q", tt.number, got, tt.text)
		}
		list := NewList(tt.number, "marker")
		if got := list.String(); got != tt.text+" marker" {
			t.Errorf("List.String(%v)=%q", tt.number, got)
		}
		// Force string comparison: x prevents numeric coercion of the right side.
		if Compare(tt.number, tt.text+"x") != -1 {
			t.Errorf("fallback comparison for %v", tt.number)
		}
	}
}

func TestNumberTextPreservesNativeValues(t *testing.T) {
	for _, tt := range []struct {
		value any
		want  string
	}{
		{nil, ""}, {true, "true"}, {int64(9007199254740993), "9007199254740993"},
		{float32(1.2), "1.2"}, {"1e+06", "1e+06"},
	} {
		if got := toString(tt.value); got != tt.want {
			t.Errorf("toString(%v)=%q want %q", tt.value, got, tt.want)
		}
	}
}
