package spx

import (
	"github.com/goplus/spx/v3/internal/scratch"
	"math"
)

// ScratchListIndex converts a Scratch one-based list index to a List position.
// Only the strings "all", "last", "random", and "any" select special positions;
// numeric indices are floored and invalid indices return Invalid. List methods
// validate the upper bound and whether the selected position is supported.
// For example, list.Delete(ScratchListIndex(-2)) leaves the list unchanged,
// while list.Delete(ScratchListIndex("all")) clears it.
func ScratchListIndex(value any) Pos {
	value = fromObj(value)
	var n float64
	if s, ok := value.(string); ok {
		switch s {
		case "all":
			return All
		case "last":
			return Last
		case "random", "any":
			return Random
		}
		n, _ = scratch.ParseNumber(s)
	} else {
		n, _ = toFloat64Any(value)
	}
	// Check before conversion so negative numbers and float-to-int overflow can
	// never become one of the negative special-position constants.
	if math.IsNaN(n) || n < 1 || n >= float64(math.MaxInt) {
		return Invalid
	}
	return int(math.Floor(n)) - 1
}
