package workerpool

import "testing"

func TestProcessTransformsEveryValueAndCloses(t *testing.T) {
	values := []int{2, 3, 5, 7}
	seen := make(map[int]int)
	for result := range Process(values, 2, func(value int) int { return value * 10 }) {
		seen[result.Index] = result.Value
	}
	if len(seen) != len(values) {
		t.Fatalf("received %d results, want %d", len(seen), len(values))
	}
	for index, value := range values {
		if seen[index] != value*10 {
			t.Fatalf("result[%d] = %d, want %d", index, seen[index], value*10)
		}
	}
}
