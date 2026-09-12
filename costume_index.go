package spx

import (
	"github.com/goplus/spx/v3/internal/scratch"
	"math"
)

// costumePosition rounds and wraps a zero-based Scratch costume selection
// before converting it to a native integer.
func costumePosition(index float64, count int) int {
	if math.IsInf(index, 0) || math.IsNaN(index) {
		index = 0
	}
	rounded := scratch.Round(index)
	if count == 0 {
		if rounded >= float64(math.MaxInt) || rounded < float64(math.MinInt) {
			return Invalid
		}
		return int(rounded)
	}
	index = math.Mod(rounded, float64(count))
	if index < 0 {
		index += float64(count)
	}
	return int(index)
}
