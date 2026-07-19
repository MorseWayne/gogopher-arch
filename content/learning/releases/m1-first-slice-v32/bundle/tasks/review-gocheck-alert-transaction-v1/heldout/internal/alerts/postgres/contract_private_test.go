package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

type transactionState struct {
	mu          sync.Mutex
	seen        map[string]bool
	version     int64
	actor       string
	begins      []driver.TxOptions
	commits     int
	rollbacks   int
	updates     int
	selects     int
	updateError error
}
type transactionConnector struct{ state *transactionState }
type transactionConn struct{ state *transactionState }
type transactionTx struct{ state *transactionState }
type transactionRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (c transactionConnector) Connect(context.Context) (driver.Conn, error) {
	return &transactionConn{state: c.state}, nil
}
func (c transactionConnector) Driver() driver.Driver { return c }
func (c transactionConnector) Open(string) (driver.Conn, error) {
	return &transactionConn{state: c.state}, nil
}
func (c *transactionConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unused") }
func (c *transactionConn) Close() error                        { return nil }
func (c *transactionConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *transactionConn) BeginTx(ctx context.Context, options driver.TxOptions) (driver.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.state.mu.Lock()
	c.state.begins = append(c.state.begins, options)
	c.state.mu.Unlock()
	return &transactionTx{state: c.state}, nil
}
func (c *transactionConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	normalized := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	switch {
	case strings.Contains(normalized, "insert into alert_commands"):
		if !strings.Contains(normalized, "on conflict (rule_id, idempotency_key) do nothing") || !strings.Contains(normalized, "returning rule_id") {
			return nil, errors.New("invalid idempotency insert")
		}
		key, err := stringArgument(args, 1)
		if err != nil {
			return nil, err
		}
		if c.state.seen[key] {
			return emptyRows("rule_id"), nil
		}
		c.state.seen[key] = true
		return valueRows([]string{"rule_id"}, []driver.Value{"rule-1"}), nil
	case strings.Contains(normalized, "update alert_rules"):
		if !strings.Contains(normalized, "version = version + 1") || !strings.Contains(normalized, "where id = $2 and version = $3") || !strings.Contains(normalized, "returning id, acknowledged_by, version") {
			return nil, errors.New("invalid conditional update")
		}
		c.state.updates++
		if c.state.updateError != nil {
			return nil, c.state.updateError
		}
		actor, err := stringArgument(args, 0)
		if err != nil {
			return nil, err
		}
		ruleID, err := stringArgument(args, 1)
		if err != nil {
			return nil, err
		}
		expected, err := int64Argument(args, 2)
		if err != nil {
			return nil, err
		}
		if ruleID != "rule-1" || expected != c.state.version {
			return emptyRows("id", "acknowledged_by", "version"), nil
		}
		c.state.version++
		c.state.actor = actor
		return valueRows([]string{"id", "acknowledged_by", "version"}, []driver.Value{"rule-1", c.state.actor, c.state.version}), nil
	case strings.Contains(normalized, "select id, acknowledged_by, version from alert_rules"):
		if !strings.Contains(normalized, "where id = $1") {
			return nil, errors.New("invalid replay lookup")
		}
		c.state.selects++
		return valueRows([]string{"id", "acknowledged_by", "version"}, []driver.Value{"rule-1", c.state.actor, c.state.version}), nil
	default:
		return nil, errors.New("unexpected query")
	}
}
func (tx *transactionTx) Commit() error {
	tx.state.mu.Lock()
	defer tx.state.mu.Unlock()
	tx.state.commits++
	return nil
}
func (tx *transactionTx) Rollback() error {
	tx.state.mu.Lock()
	defer tx.state.mu.Unlock()
	tx.state.rollbacks++
	return nil
}
func (rows *transactionRows) Columns() []string { return rows.columns }
func (rows *transactionRows) Close() error      { return nil }
func (rows *transactionRows) Next(destination []driver.Value) error {
	if rows.index >= len(rows.values) {
		return io.EOF
	}
	copy(destination, rows.values[rows.index])
	rows.index++
	return nil
}

func TestSerializableTransactionBoundary(t *testing.T) {
	store, state, closeDB := transactionStore(t)
	defer closeDB()
	got, err := store.Acknowledge(context.Background(), AcknowledgeCommand{RuleID: "rule-1", IdempotencyKey: "ack-1", Actor: "operator-a", ExpectedVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got != (Rule{ID: "rule-1", AcknowledgedBy: "operator-a", Version: 2}) {
		t.Fatalf("rule = %#v", got)
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if len(state.begins) != 1 || sql.IsolationLevel(state.begins[0].Isolation) != sql.LevelSerializable || state.commits != 1 || state.rollbacks != 0 {
		t.Fatalf("transaction state = %#v", state)
	}
}
func TestIdempotencyReplayStable(t *testing.T) {
	store, state, closeDB := transactionStore(t)
	defer closeDB()
	command := AcknowledgeCommand{RuleID: "rule-1", IdempotencyKey: "ack-1", Actor: "operator-a", ExpectedVersion: 1}
	first, err := store.Acknowledge(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Acknowledge(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("replay changed result: %#v != %#v", first, second)
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.updates != 1 || state.selects != 1 || state.commits != 2 {
		t.Fatalf("replay state = %#v", state)
	}
}
func TestConcurrentUpdateConflictStable(t *testing.T) {
	store, state, closeDB := transactionStore(t)
	defer closeDB()
	commands := []AcknowledgeCommand{{RuleID: "rule-1", IdempotencyKey: "ack-a", Actor: "operator-a", ExpectedVersion: 1}, {RuleID: "rule-1", IdempotencyKey: "ack-b", Actor: "operator-b", ExpectedVersion: 1}}
	errorsSeen := make(chan error, len(commands))
	var wait sync.WaitGroup
	for _, command := range commands {
		wait.Add(1)
		go func(command AcknowledgeCommand) {
			defer wait.Done()
			_, err := store.Acknowledge(context.Background(), command)
			errorsSeen <- err
		}(command)
	}
	wait.Wait()
	close(errorsSeen)
	passed, conflicts := 0, 0
	for err := range errorsSeen {
		if err == nil {
			passed++
		} else if errors.Is(err, ErrConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if passed != 1 || conflicts != 1 {
		t.Fatalf("passed=%d conflicts=%d", passed, conflicts)
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.commits != 1 || state.rollbacks != 1 || state.version != 2 {
		t.Fatalf("concurrent state = %#v", state)
	}
}
func TestFailureRollsBack(t *testing.T) {
	store, state, closeDB := transactionStore(t)
	defer closeDB()
	state.updateError = errors.New("write failed")
	_, err := store.Acknowledge(context.Background(), AcknowledgeCommand{RuleID: "rule-1", IdempotencyKey: "ack-1", Actor: "operator-a", ExpectedVersion: 1})
	if err == nil || !strings.Contains(err.Error(), "write failed") {
		t.Fatalf("error = %v", err)
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.commits != 0 || state.rollbacks != 1 {
		t.Fatalf("failure state = %#v", state)
	}
}

func transactionStore(t *testing.T) (*Store, *transactionState, func()) {
	t.Helper()
	state := &transactionState{seen: make(map[string]bool), version: 1}
	db := sql.OpenDB(transactionConnector{state: state})
	db.SetMaxOpenConns(4)
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	return store, state, func() { _ = db.Close() }
}
func stringArgument(args []driver.NamedValue, index int) (string, error) {
	if index >= len(args) {
		return "", errors.New("missing argument")
	}
	value, ok := args[index].Value.(string)
	if !ok {
		return "", errors.New("argument is not string")
	}
	return value, nil
}
func int64Argument(args []driver.NamedValue, index int) (int64, error) {
	if index >= len(args) {
		return 0, errors.New("missing argument")
	}
	value, ok := args[index].Value.(int64)
	if !ok {
		return 0, errors.New("argument is not int64")
	}
	return value, nil
}
func emptyRows(columns ...string) driver.Rows { return &transactionRows{columns: columns} }
func valueRows(columns []string, values []driver.Value) driver.Rows {
	return &transactionRows{columns: columns, values: [][]driver.Value{values}}
}
