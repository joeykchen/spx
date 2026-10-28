/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
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

// Package scratchcompat provides SPX's compatibility operators for Scratch
// programs, using spx.Value and spx.List. Converters select them per block.
//
// These operators retain SPX's equality tolerance and NaN-to-zero division;
// they are not an exact emulation of the Scratch VM.
package scratchcompat

import (
	"github.com/goplus/spx/v4"
	"github.com/goplus/spx/v4/internal/scratch"
)

// LimitPrecision rounds a coordinate when it is strictly within 1e-9 of an
// integer. Apply it to a sprite's own x/y reporters, leaving property queries
// and stored coordinates unchanged. NaN and infinities pass through.
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

// CharAt returns the rune at the 1-based index in s, or "" if out of range.
// Like spx.CharAt, it counts Unicode code points, not UTF-16 code units.
func CharAt(s string, idx int) string {
	if idx < 1 {
		return ""
	}
	return scratch.CharAt(s, idx-1)
}

// Compare compares numbers when both operands are numeric, otherwise strings
// ignoring case. It accepts spx.Value and compound signs from older converters.
// It returns -1, 0, or 1 without Equal's numeric tolerance.
func Compare(a, b any) int {
	return scratch.Compare(unwrapValue(a), unwrapValue(b))
}

// Equal compares like Compare, allowing up to two ULPs of numeric rounding error.
// Numeric comparisons involving NaN return false. It accepts spx.Value operands.
func Equal(a, b any) bool {
	return scratch.Equal(unwrapValue(a), unwrapValue(b))
}

func unwrapValue(v any) any {
	if value, ok := v.(spx.Value); ok {
		return value.Raw()
	}
	return v
}
