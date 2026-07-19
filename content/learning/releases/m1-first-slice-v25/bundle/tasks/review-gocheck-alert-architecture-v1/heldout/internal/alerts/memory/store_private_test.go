package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"gocheckhub/internal/alerts"
	"gocheckhub/internal/alerts/memory"
)

func TestMemoryRepositoryContract(t *testing.T) {
	store := memory.NewStore()
	if store == nil {
		t.Fatal("NewStore returned nil")
	}
	if err := store.Save(context.Background(), alerts.Rule{ID: "one", Destination: "OPS@EXAMPLE.COM"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), alerts.Rule{ID: "two", Destination: "ops@example.com"}); !errors.Is(err, alerts.ErrRuleExists) {
		t.Fatalf("duplicate error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.Save(cancelled, alerts.Rule{ID: "cancelled", Destination: "cancelled@example.com"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled save error = %v", err)
	}

	var group sync.WaitGroup
	errorsSeen := make(chan error, 16)
	for index := 0; index < 16; index++ {
		group.Add(1)
		go func(index int) {
			defer group.Done()
			errorsSeen <- store.Save(context.Background(), alerts.Rule{ID: fmt.Sprint(index), Destination: fmt.Sprintf("team-%d@example.com", index)})
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
