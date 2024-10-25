package alg

import "math"

func GetSum(a, b int) int {
	x, y := int(math.Abs(float64(a))), int(math.Abs(float64(b)))
	if x < y {
		return GetSum(b, a)
	}

	sign := 1
	if a < 0 {
		sign = -1
	}

	if a*b >= 0 {
		if y != 0 {
			carry := (x & y) << 1
			x ^= y
			y = carry
		} else {
			for y != 0 {
				borrow := ((^x) & y) << 1
				x ^= y
				y = borrow
			}
		}
	}

	return x * sign
}
