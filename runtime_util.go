/*
 * Copyright (c) 2021 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package spx

import (
	"strings"

	"github.com/goplus/spx/v4/internal/engine"
	"github.com/goplus/spx/v4/internal/scratch"
	itime "github.com/goplus/spx/v4/internal/time"
)

// Rand__0 returns a random integer between from and to (inclusive).
func Rand__0(from, to int) int {
	if to < from {
		from, to = to, from
	}
	return from + randomIntn(to-from+1)
}

// Rand__1 returns a random float64 between from and to.
func Rand__1(from, to float64) float64 {
	if to < from {
		from, to = to, from
	}
	return randomFloat64()*(to-from) + from
}

// Iround returns an integer value, while math.Round returns a float value.
func Iround(v float64) int {
	if v >= 0 {
		return int(v + 0.5)
	}
	return int(v - 0.5)
}

// LimitPrecision rounds a coordinate when it is strictly within 1e-9 of an
// integer; NaN and infinities pass through. Apply it to a sprite's own x/y
// reporters, leaving property queries unchanged.
func LimitPrecision(coordinate float64) float64 {
	return scratch.LimitPrecision(coordinate)
}

// Div divides a by b, converting a NaN result to zero and preserving infinities.
func Div(a, b float64) float64 {
	return scratch.Div(a, b)
}

// FloorMod returns the remainder using floored division. A non-zero result
// has the divisor's sign. A zero divisor produces NaN.
func FloorMod(dividend, divisor float64) float64 {
	return scratch.FloorMod(dividend, divisor)
}

// Contains reports whether substr occurs in s, ignoring case.
func Contains(s, substr string) bool {
	return scratch.Contains(s, substr)
}

// Compare compares numbers when both operands are numeric, otherwise strings
// ignoring case. It returns -1, 0, or 1 without Equal's numeric tolerance.
func Compare(v1, v2 any) int {
	return scratch.Compare(unwrapValue(v1), unwrapValue(v2))
}

// Equal compares like Compare, allowing up to two ULPs of numeric rounding error.
// Numeric comparisons involving NaN return false.
func Equal(v1, v2 any) bool {
	return scratch.Equal(unwrapValue(v1), unwrapValue(v2))
}

// DeltaTime returns the time elapsed since the previous frame.
func DeltaTime() Seconds {
	return itime.DeltaTime()
}

// TimeSinceLevelLoad returns the time elapsed since the current level started.
func TimeSinceLevelLoad() Seconds {
	return itime.TimeSinceLevelLoad()
}

// CharAt returns the rune at the 0-based index in s, or "" if out of range.
func CharAt(s string, idx int) string {
	return scratch.CharAt(s, idx)
}

// Exit__0 exits the program with exit code 0.
func Exit__0() {
	engine.RequestExit(0)
}

// Exit__1 exits the program with the specified exit code.
func Exit__1(code int) {
	engine.RequestExit(int64(code))
}

// PenColorParamFromString converts a string to its corresponding PenColorParam.
// It accepts "color", "saturation", "brightness", and "transparency" (case-insensitive).
// Returns PenNone if the string is not recognized.
func PenColorParamFromString(s string) PenColorParam {
	switch strings.ToLower(s) {
	case "color":
		return PenHue
	case "saturation":
		return PenSaturation
	case "brightness":
		return PenBrightness
	case "transparency":
		return PenTransparency
	default:
		return PenNone
	}
}
