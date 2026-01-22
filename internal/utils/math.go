package utils

import "math/rand"

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RandomNumber(min, max int) int {
	if min > max {
		return min
	}
	return rand.Intn(max-min+1) + min
}
