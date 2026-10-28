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

package scratchcompat_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/goplus/spx/v4"
	"github.com/goplus/spx/v4/scratchcompat"
)

func TestCompareAndEqual(t *testing.T) {
	items := spx.NewList("01", "Apple")
	// Init can retain wrapped elements; At adds one more Value wrapper.
	var nested spx.List
	nested.Init(spx.NewValue(2))
	value := spx.NewValue(2)
	var nilPointer *spx.Value
	twoULPs := math.Nextafter(math.Nextafter(116, math.Inf(1)), math.Inf(1))
	threeULPs := math.Nextafter(twoULPs, math.Inf(1))

	for _, api := range []struct {
		name    string
		compare func(any, any) int
		equal   func(any, any) bool
	}{
		{"spx", spx.Compare, spx.Equal},
		{"scratchcompat", scratchcompat.Compare, scratchcompat.Equal},
	} {
		t.Run(api.name, func(t *testing.T) {
			for _, tt := range []struct {
				name    string
				a, b    any
				compare int
				equal   bool
			}{
				{"list number", items.At(0), 1, 0, true},
				{"list text", items.At(1), "apple", 0, true},
				{"wrapped boolean", spx.NewValue(true), 1, 0, true},
				{"empty item", items.At(10), "", 0, true},
				{"empty item is not zero", items.At(10), 0, -1, false},
				{"blank is not zero", spx.NewValue(" "), 0, -1, false},
				{"old converter signs", spx.NewValue("--7"), 7, 0, true},
				{"nonnumeric right operand", 10, "2x", -1, false},
				{"single wrapper is numeric", value, 10, -1, false},
				{"nested wrapper stays textual", nested.At(0), 10, 1, false},
				{"pointer stays textual", &value, 10, 1, false},
				{"nil pointer stays textual", nilPointer, "<nil>", 0, true},
				{"two ULPs", spx.NewValue(twoULPs), 116, 1, true},
				{"three ULPs", spx.NewValue(threeULPs), 116, 1, false},
				{"NaN retains comparison behavior", math.NaN(), math.NaN(), 0, false},
				{"same infinity", math.Inf(1), math.Inf(1), 0, true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					if got := api.compare(tt.a, tt.b); got != tt.compare {
						t.Errorf("Compare(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.compare)
					}
					if got := api.equal(tt.a, tt.b); got != tt.equal {
						t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.equal)
					}
				})
			}
		})
	}
}

func TestOperators(t *testing.T) {
	negativeZero := math.Copysign(0, -1)
	for _, tt := range []struct {
		name      string
		got, want float64
	}{
		{"division", scratchcompat.Div(7, 2), 3.5},
		{"indeterminate division", scratchcompat.Div(0, 0), 0},
		{"infinite division", scratchcompat.Div(1, 0), math.Inf(1)},
		{"signed zero division", scratchcompat.Div(0, -1), negativeZero},
		{"negative dividend", scratchcompat.FloorMod(-7, 5), 3},
		{"negative divisor", scratchcompat.FloorMod(7, -5), -3},
		{"zero divisor", scratchcompat.FloorMod(1, 0), math.NaN()},
		{"negative coordinate", scratchcompat.LimitPrecision(-5e-10), negativeZero},
		{"inside threshold", scratchcompat.LimitPrecision(math.Nextafter(1e-9, 0)), 0},
		{"at threshold", scratchcompat.LimitPrecision(1e-9), 1e-9},
		{"fractional coordinate", scratchcompat.LimitPrecision(1.5), 1.5},
		{"infinite coordinate", scratchcompat.LimitPrecision(math.Inf(1)), math.Inf(1)},
		{"NaN coordinate", scratchcompat.LimitPrecision(math.NaN()), math.NaN()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if math.IsNaN(tt.want) && math.IsNaN(tt.got) {
				return
			}
			if tt.got != tt.want || math.Signbit(tt.got) != math.Signbit(tt.want) {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
	if !scratchcompat.Contains("Scratch", "RaTc") || scratchcompat.Contains("Scratch", "sprite") {
		t.Fatal("Contains lost case-insensitive substring matching")
	}
}

func TestCharAt(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		index      int
		want       string
	}{
		{"first", "Scratch", 1, "S"},
		{"last", "Scratch", 7, "h"},
		{"zero", "Scratch", 0, ""},
		{"negative", "Scratch", -1, ""},
		{"minimum index", "Scratch", math.MinInt, ""},
		{"maximum index", "Scratch", math.MaxInt, ""},
		{"past end", "Scratch", 8, ""},
		{"empty", "", 1, ""},
		{"Chinese", "你好", 2, "好"},
		{"emoji", "A😀B", 2, "😀"},
		{"after emoji", "A😀B", 3, "B"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := scratchcompat.CharAt(tt.text, tt.index); got != tt.want {
				t.Fatalf("CharAt(%q, %d) = %q, want %q", tt.text, tt.index, got, tt.want)
			}
		})
	}
	if got := spx.CharAt("Scratch", 0); got != "S" {
		t.Fatalf("spx.CharAt lost zero-based indexing: got %q", got)
	}
}

func ExampleLimitPrecision() {
	x := 5e-10
	// Only the sprite's own position reporter applies precision correction.
	fmt.Println(scratchcompat.LimitPrecision(x))
	// Property queries and stored coordinates retain their original precision.
	fmt.Println(x)
	// Output:
	// 0
	// 5e-10
}
