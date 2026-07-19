package workerpool

type Result struct {
	Index int
	Value int
}

// Process transforms every value using at most workers concurrent workers.
func Process(values []int, workers int, transform func(int) int) <-chan Result {
	results := make(chan Result)
	close(results)
	return results
}
