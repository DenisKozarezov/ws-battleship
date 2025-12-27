package math

import "golang.org/x/exp/constraints"

func Clamp[T constraints.Float | constraints.Integer](val T, min T, max T) T {
	if val < min {
		val = min
	}

	if val > max {
		val = max
	}

	return val
}
