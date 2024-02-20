package mymath

import "sync"

/*
	some sample functions are here...
*/

func MultipleSum(sourceNumber int, sums ...int) int {
	result := sourceNumber

	var wg sync.WaitGroup

	for _, num := range sums {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			result += n
		}(num)
	}

	wg.Wait()

	return result
}
