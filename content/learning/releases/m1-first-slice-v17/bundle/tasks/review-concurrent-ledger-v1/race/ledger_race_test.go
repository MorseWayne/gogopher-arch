package ledger

import (
	"sync"
	"testing"
)

func TestLedgerConcurrentAccessIsRaceFree(t *testing.T) {
	ledger := New()
	var group sync.WaitGroup
	group.Add(16)
	for range 16 {
		go func() {
			defer group.Done()
			for range 250 {
				ledger.Credit("cash", 1)
				_ = ledger.Balances()
			}
		}()
	}
	group.Wait()
	if got := ledger.Balances()["cash"]; got != 4000 {
		t.Fatalf("cash balance = %d, want 4000", got)
	}
}
