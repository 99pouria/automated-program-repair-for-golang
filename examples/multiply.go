package mymath

func Multiplier(a, b int) int {
	result := a

	for range b - 1 {
		result += a
	}

	return result
}
