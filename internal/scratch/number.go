package scratch

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

// ParseNumber follows Number(string), keeping invalid text distinct from
// zero for comparisons. Native Value.Int/Float retain their existing conversion.
func ParseNumber(s string) (float64, bool) {
	s = TrimSpace(s)
	if s == "" {
		return 0, true
	}
	switch s {
	case "Infinity", "+Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}
	if len(s) > 2 && s[0] == '0' {
		base := 0
		switch s[1] {
		case 'x', 'X':
			base = 16
		case 'o', 'O':
			base = 8
		case 'b', 'B':
			base = 2
		}
		if base != 0 {
			// A sign after the radix prefix is invalid in JavaScript.
			if s[2] == '+' || s[2] == '-' {
				return 0, false
			}
			n, ok := new(big.Int).SetString(s[2:], base)
			if !ok {
				return 0, false
			}
			value, _ := new(big.Float).SetInt(n).Float64()
			return value, true
		}
	}
	// ParseFloat accepts Go-only separators, hexadecimal floats and Inf spellings.
	if strings.ContainsAny(s, "_xXpPiInN") {
		return 0, false
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		if e, ok := err.(*strconv.NumError); !ok || e.Err != strconv.ErrRange {
			return 0, false
		}
	}
	return n, true
}

// TrimSpace removes JavaScript whitespace, including BOM but excluding NEL.
func TrimSpace(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r == '\ufeff' || r != '\u0085' && unicode.IsSpace(r) })
}
