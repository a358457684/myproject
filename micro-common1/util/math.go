package util

import "math"

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
