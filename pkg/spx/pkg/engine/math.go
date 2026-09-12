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

	mathf "github.com/goplus/spbase/mathf"
)

func Abs(x float64) float64 {
	return math.Abs(x)
}
func Sign(x float64) int64 {
	if x < 0 {
		return -1
	} else if x > 0 {
		return 1
	}
	return 0
}

func DegToRad(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}

func RadToDeg(radians float64) float64 {
	return radians * (180.0 / math.Pi)
}

func AngleToPoint(from, to mathf.Vec2) float64 {
	return Angle(to.Sub(from))
}

// Angle returns the angle of v in radians, measured from the positive X axis.
func Angle(v mathf.Vec2) float64 {
	return mathf.Atan2(v.Y, v.X)
}

// HeadingToPoint returns the SPX heading in degrees needed to face from `from`
// toward `to`. The result uses SPX heading semantics, where 0 points up and 90
// points right, and is intentionally not normalized to (-180, 180].
func HeadingToPoint(from, to mathf.Vec2) float64 {
	return 90 - RadToDeg(AngleToPoint(from, to))
}

const (
	fullCircleDegrees = 360.0
	halfCircleDegrees = 180.0
)

// NormalizeAngleRange chooses equivalent angles in degrees with the shortest
// rotation path. Exactly half a turn keeps the order of the normalized angles.
func NormalizeAngleRange(from, to float64) (float64, float64) {
	fromNorm := positiveDegrees(from)
	toNorm := positiveDegrees(to)

	if toNorm-fromNorm > halfCircleDegrees {
		fromNorm += fullCircleDegrees
	} else if fromNorm-toNorm > halfCircleDegrees {
		toNorm += fullCircleDegrees
	}

	return fromNorm, toNorm
}

// NormalizeDegrees normalizes an angle in degrees to (-180, 180].
// Nonfinite angles produce NaN.
func NormalizeDegrees(angle float64) float64 {
	angle = math.Mod(angle, fullCircleDegrees)
	if angle <= -halfCircleDegrees {
		angle += fullCircleDegrees
	} else if angle > halfCircleDegrees {
		angle -= fullCircleDegrees
	}
	return angle
}

func positiveDegrees(angle float64) float64 {
	angle = math.Mod(angle, fullCircleDegrees)
	if angle < 0 {
		angle += fullCircleDegrees
	}
	return angle
}
