package scratch

import (
	"math"
	"strconv"
	"strings"
)

// FormatNumber formats float64 values without changing native integer or Stringer formatting.
func FormatNumber(n float64) string {
	if n == 0 {
		return "0"
	}
	if math.IsInf(n, 1) {
		return "Infinity"
	}
	if math.IsInf(n, -1) {
		return "-Infinity"
	}
	if math.IsNaN(n) {
		return "NaN"
	}
	format := byte('e')
	if math.Abs(n) >= 1e-6 && math.Abs(n) < 1e21 {
		format = 'f'
	}
	text := strconv.FormatFloat(n, format, -1, 64)
	if mantissa, exponent, ok := strings.Cut(text, "e"); ok {
		text = mantissa + "e" + exponent[:1] + strings.TrimLeft(exponent[1:], "0")
	}
	return text
}
