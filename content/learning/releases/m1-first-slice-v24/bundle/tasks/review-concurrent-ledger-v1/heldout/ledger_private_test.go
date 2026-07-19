package ledger

import "testing"

func TestLedgerSnapshotOwnsState(t *testing.T) {
	ledger := New()
	ledger.Credit("cash", 10)
	copy := ledger.Balances()
	copy["cash"] = -1
	if got := ledger.Balances()["cash"]; got != 10 {
		t.Fatalf("Balances() shares internal map: %d", got)
	}
}
