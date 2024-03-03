package example

// CSum is a concurrent sum
func CSum(a, b int) int {
	var result int

	go func() {
		result += a
	}()

	go func() {
		result += b
	}()

	return result
}
