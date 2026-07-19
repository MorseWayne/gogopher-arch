package checker

import (
	"context"
	"sync"
	"testing"
)

func TestCheckAllVisitsEveryTarget(t *testing.T) {
	var mutex sync.Mutex
	seen := make(map[string]int)
	err := CheckAll(context.Background(), []string{"api", "db", "cache", "queue"}, 2, func(_ context.Context, target string) error {
		mutex.Lock()
		seen[target]++
		mutex.Unlock()
		return nil
	})
	if err != nil || len(seen) != 4 {
		t.Fatalf("CheckAll() error = %v, seen = %v", err, seen)
	}
	for target, count := range seen {
		if count != 1 {
			t.Fatalf("target %s checked %d times", target, count)
		}
	}
}
