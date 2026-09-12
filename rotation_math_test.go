package spx

import (
	"math"
	"testing"
)

func TestToRadianPreservesFloatingPointEdges(t *testing.T) {
	for _, tc := range []struct{ degrees, radians float64 }{
		{0, 0}, {math.Copysign(0, -1), math.Copysign(0, -1)},
		{180, math.Pi}, {-180, -math.Pi},
		{1e308, math.Inf(1)}, {-1e308, math.Inf(-1)},
		{math.SmallestNonzeroFloat64, 0},
	} {
		if got := toRadian(tc.degrees); math.Float64bits(got) != math.Float64bits(tc.radians) {
			t.Errorf("toRadian(%g)=%g, want %g", tc.degrees, got, tc.radians)
		}
	}
	if !math.IsNaN(toRadian(math.NaN())) {
		t.Fatal("NaN changed")
	}
}
