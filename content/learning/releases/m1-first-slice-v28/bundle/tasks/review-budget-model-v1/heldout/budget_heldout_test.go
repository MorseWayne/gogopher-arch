package budgetmodel

import "testing"

func TestBudgetEnforcesConstructionInvariants(t *testing.T) {
	for _, input := range []struct {
		name  string
		limit int
	}{{"", 10}, {"  ", 10}, {"api", 0}, {"api", -2}} {
		if _, err := NewBudget(input.name, input.limit); err == nil {
			t.Fatalf("NewBudget(%q, %d) error = nil", input.name, input.limit)
		}
	}
	budget, err := NewBudget(" worker ", 9)
	if err != nil || budget.Label() != "worker:9" {
		t.Fatalf("valid budget = %#v, %v", budget, err)
	}
}

func TestBudgetBehaviorMethodsPreserveStateOnFailure(t *testing.T) {
	budget, err := NewBudget("db", 8)
	if err != nil {
		t.Fatal(err)
	}
	for _, amount := range []int{0, -1, 9} {
		if err := budget.Spend(amount); err == nil || budget.Remaining() != 8 {
			t.Fatalf("Spend(%d): err=%v remaining=%d", amount, err, budget.Remaining())
		}
	}
	if err := budget.Spend(5); err != nil {
		t.Fatal(err)
	}
	for _, amount := range []int{0, -1, 6} {
		if err := budget.Refund(amount); err == nil || budget.Remaining() != 3 {
			t.Fatalf("Refund(%d): err=%v remaining=%d", amount, err, budget.Remaining())
		}
	}
	if err := budget.Refund(5); err != nil || budget.Remaining() != 8 {
		t.Fatalf("full refund: err=%v remaining=%d", err, budget.Remaining())
	}
}
