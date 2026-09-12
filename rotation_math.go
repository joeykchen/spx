package spx

import "math"

// normalizeAngleRange chooses equivalent angles with the shortest rotation path.
func normalizeAngleRange(from, to float64) (float64, float64) {
	fromNorm := positiveDegrees(from)
	toNorm := positiveDegrees(to)

	if toNorm-fromNorm > halfCircleDegrees {
		fromNorm += fullCircleDegrees
	} else if fromNorm-toNorm > halfCircleDegrees {
		toNorm += fullCircleDegrees
	}

	return fromNorm, toNorm
}

// toRadian preserves the historical multiply-then-divide evaluation.
// engine.DegToRad uses a precomputed factor, which changes rounding and overflow.
func toRadian(dir float64) float64 {
	return math.Pi * dir / 180
}

// normalizeDirection normalizes a direction angle to the range (-180, 180].
func normalizeDirection(dir float64) float64 {
	dir = math.Mod(dir, fullCircleDegrees)
	if dir <= -halfCircleDegrees {
		dir += fullCircleDegrees
	} else if dir > halfCircleDegrees {
		dir -= fullCircleDegrees
	}
	return dir
}

func positiveDegrees(direction float64) float64 {
	direction = math.Mod(direction, fullCircleDegrees)
	if direction < 0 {
		direction += fullCircleDegrees
	}
	return direction
}
