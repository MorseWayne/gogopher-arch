package budgetmodel

type Budget struct {
	name  string
	limit int
	spent int
}

func NewBudget(name string, limit int) (Budget, error) {
	// TODO: 建立合法预算。
	return Budget{}, nil
}

func (b *Budget) Spend(amount int) error {
	// TODO: 校验并更新支出。
	return nil
}

func (b *Budget) Refund(amount int) error {
	// TODO: 校验并更新支出。
	return nil
}

func (b Budget) Remaining() int {
	return 0
}

func (b Budget) Label() string {
	return ""
}
