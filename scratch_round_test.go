package spx

import (
	"github.com/goplus/spx/v3/internal/scratch"
	"math"
	"testing"
)

func TestScratchRoundTies(t *testing.T) {
	for _, tt := range []struct {
		input float64
		want  int
	}{
		{-2.5, -2}, {-1.5, -1}, {-0.5, 0}, {0.5, 1}, {1.5, 2},
		{math.Nextafter(0.5, 0), 0}, {math.Nextafter(-1.5, math.Inf(-1)), -2},
		{math.Nextafter(-1.5, 0), -1}, {math.Nextafter(1, 0), 1},
	} {
		if got := Iround(tt.input); got != tt.want {
			t.Errorf("Iround(%.17g)=%d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestScratchRoundPreservesFloatingPointEdges(t *testing.T) {
	for _, v := range []float64{-0.1, -0.5, math.Copysign(0, -1)} {
		if got := scratch.Round(v); got != 0 || !math.Signbit(got) {
			t.Errorf("round(%g) did not retain negative zero", v)
		}
	}
	for _, v := range []float64{math.MaxFloat64, math.Inf(1), math.Inf(-1)} {
		if scratch.Round(v) != v {
			t.Errorf("round changed %g", v)
		}
	}
	if !math.IsNaN(scratch.Round(math.NaN())) {
		t.Fatal("round changed NaN")
	}
}
