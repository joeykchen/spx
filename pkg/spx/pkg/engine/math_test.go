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

package engine

import (
	"math"
	"testing"

	mathf "github.com/goplus/spbase/mathf"
)

func TestHeadingToPoint(t *testing.T) {
	from := mathf.NewVec2(0, 0)

	tests := []struct {
		name string
		to   mathf.Vec2
		want float64
	}{
		{name: "right", to: mathf.NewVec2(10, 0), want: 90},
		{name: "up", to: mathf.NewVec2(0, 10), want: 0},
		{name: "left", to: mathf.NewVec2(-10, 0), want: -90},
		{name: "down", to: mathf.NewVec2(0, -10), want: 180},
		{name: "bottom left keeps raw heading", to: mathf.NewVec2(-10, -10), want: 225},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HeadingToPoint(from, tt.to); Abs(got-tt.want) > 1e-9 {
				t.Fatalf("HeadingToPoint(%v, %v) = %v, want %v", from, tt.to, got, tt.want)
			}
		})
	}
}

func TestNormalizeDegrees(t *testing.T) {
	for _, tc := range []struct{ angle, want float64 }{
		{0, 0}, {math.Copysign(0, -1), math.Copysign(0, -1)},
		{180, 180}, {-180, 180}, {900, 180}, {-900, 180},
		{1081, 1}, {-1081, -1}, {720.25, 0.25}, {-720.25, -0.25},
	} {
		if got := NormalizeDegrees(tc.angle); math.Float64bits(got) != math.Float64bits(tc.want) {
			t.Errorf("NormalizeDegrees(%g) = %g, want %g", tc.angle, got, tc.want)
		}
	}
	for _, angle := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if !math.IsNaN(NormalizeDegrees(angle)) {
			t.Errorf("NormalizeDegrees(%g) must be NaN", angle)
		}
	}
}

func TestNormalizeAngleRange(t *testing.T) {
	for _, tc := range []struct{ from, to, wantFrom, wantTo float64 }{
		{-1079, 719, 361, 359}, {719, -1079, 359, 361},
		{0, 180, 0, 180}, {180, 0, 180, 0},
		{-180, 180, 180, 180}, {720, -720, 0, 0},
		{359.5, 0.5, 359.5, 360.5},
	} {
		from, to := NormalizeAngleRange(tc.from, tc.to)
		if from != tc.wantFrom || to != tc.wantTo {
			t.Errorf("NormalizeAngleRange(%g, %g) = (%g, %g), want (%g, %g)",
				tc.from, tc.to, from, to, tc.wantFrom, tc.wantTo)
		}
	}
}

func TestDegToRad(t *testing.T) {
	for _, tc := range []struct{ degrees, radians float64 }{
		{0, 0}, {math.Copysign(0, -1), math.Copysign(0, -1)},
		{180, math.Pi}, {-180, -math.Pi},
		{math.Inf(1), math.Inf(1)}, {math.Inf(-1), math.Inf(-1)},
		{math.SmallestNonzeroFloat64, 0},
	} {
		if got := DegToRad(tc.degrees); math.Float64bits(got) != math.Float64bits(tc.radians) {
			t.Errorf("DegToRad(%g) = %g, want %g", tc.degrees, got, tc.radians)
		}
	}
	for _, degrees := range []float64{1e308, -1e308} {
		radians := DegToRad(degrees)
		if math.IsInf(radians, 0) || math.IsNaN(radians) || math.Abs(RadToDeg(radians)/degrees-1) > 1e-15 {
			t.Errorf("large finite angle %g did not survive conversion: %g", degrees, radians)
		}
	}
	if !math.IsNaN(DegToRad(math.NaN())) {
		t.Fatal("NaN changed")
	}
}
