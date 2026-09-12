package scratch

import "math"

// Round keeps rounding in floating point for callers such as costume
// selection, which wrap large indices before converting them to native ints.
func Round(v float64) float64 {
	integer, fraction := math.Modf(v)
	switch {
	case fraction >= 0.5:
		return integer + 1
	case fraction < -0.5:
		return integer - 1
	default:
		return integer
	}
}
