package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"gocheckhub/internal/checks"
	"gocheckhub/internal/checks/memory"
)

func TestMemoryRepositoryContract(t *testing.T) {
	repository := memory.NewRepository()
	if repository == nil {
		t.Fatal("NewRepository returned nil")
	}
	if err := repository.Create(context.Background(), checks.Check{ID: "one", Target: "API.EXAMPLE.COM"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(context.Background(), checks.Check{ID: "two", Target: "api.example.com"}); !errors.Is(err, checks.ErrCheckExists) {
		t.Fatalf("duplicate error = %v", err)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repository.Create(cancelled, checks.Check{ID: "cancelled", Target: "cancelled.example.com"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled create error = %v", err)
	}

	var group sync.WaitGroup
	errorsSeen := make(chan error, 16)
	for index := 0; index < 16; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			errorsSeen <- repository.Create(context.Background(), checks.Check{ID: fmt.Sprint(index), Target: fmt.Sprintf("target-%d", index)})
		}(index)
	}
	group.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if err != nil {
			t.Fatal(err)
		}
	}
}
