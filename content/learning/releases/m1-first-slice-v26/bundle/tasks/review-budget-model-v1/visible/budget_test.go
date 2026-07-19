package budgetmodel

import "testing"

func TestBudgetSpendAndRefund(t *testing.T) {
	budget, err := NewBudget("api", 10)
	if err != nil {
		t.Fatal(err)
	}
	if err := budget.Spend(7); err != nil {
		t.Fatal(err)
	}
	if err := budget.Refund(2); err != nil {
		t.Fatal(err)
	}
	if budget.Remaining() != 5 || budget.Label() != "api:5" {
		t.Fatalf("budget = %#v label=%q", budget, budget.Label())
	}
}
