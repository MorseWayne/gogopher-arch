package ledger

type Ledger struct {
	balances map[string]int64
}

func New() *Ledger {
	return &Ledger{balances: make(map[string]int64)}
}

func (l *Ledger) Credit(account string, amount int64) {
	l.balances[account] += amount
}

func (l *Ledger) Balances() map[string]int64 {
	return l.balances
}
