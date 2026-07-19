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
	status      string
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
	case strings.Contains(normalized, "insert into run_commands"):
		if !strings.Contains(normalized, "on conflict (run_id, idempotency_key) do nothing") || !strings.Contains(normalized, "returning run_id") {
			return nil, errors.New("invalid idempotency insert")
		}
		key, err := stringArgument(args, 1)
		if err != nil {
			return nil, err
		}
		if c.state.seen[key] {
			return emptyRows("run_id"), nil
		}
		c.state.seen[key] = true
		return valueRows([]string{"run_id"}, []driver.Value{"run-1"}), nil
	case strings.Contains(normalized, "update check_runs"):
		if !strings.Contains(normalized, "version = version + 1") || !strings.Contains(normalized, "where id = $2 and version = $3") || !strings.Contains(normalized, "returning id, status, version") {
			return nil, errors.New("invalid conditional update")
		}
		c.state.updates++
		if c.state.updateError != nil {
			return nil, c.state.updateError
		}
		status, err := stringArgument(args, 0)
		if err != nil {
			return nil, err
		}
		runID, err := stringArgument(args, 1)
		if err != nil {
			return nil, err
		}
		expected, err := int64Argument(args, 2)
		if err != nil {
			return nil, err
		}
		if runID != "run-1" || expected != c.state.version {
			return emptyRows("id", "status", "version"), nil
		}
		c.state.version++
		c.state.status = status
		return valueRows([]string{"id", "status", "version"}, []driver.Value{"run-1", c.state.status, c.state.version}), nil
	case strings.Contains(normalized, "select id, status, version from check_runs"):
		if !strings.Contains(normalized, "where id = $1") {
			return nil, errors.New("invalid replay lookup")
		}
		c.state.selects++
		return valueRows([]string{"id", "status", "version"}, []driver.Value{"run-1", c.state.status, c.state.version}), nil
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
	got, err := store.Complete(context.Background(), CompleteCommand{RunID: "run-1", IdempotencyKey: "complete-1", Status: "succeeded", ExpectedVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got != (Run{ID: "run-1", Status: "succeeded", Version: 2}) {
		t.Fatalf("run = %#v", got)
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
	command := CompleteCommand{RunID: "run-1", IdempotencyKey: "complete-1", Status: "succeeded", ExpectedVersion: 1}
	first, err := store.Complete(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Complete(context.Background(), command)
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
	commands := []CompleteCommand{{RunID: "run-1", IdempotencyKey: "complete-a", Status: "succeeded", ExpectedVersion: 1}, {RunID: "run-1", IdempotencyKey: "complete-b", Status: "failed", ExpectedVersion: 1}}
	errorsSeen := make(chan error, len(commands))
	var wait sync.WaitGroup
	for _, command := range commands {
		wait.Add(1)
		go func(command CompleteCommand) {
			defer wait.Done()
			_, err := store.Complete(context.Background(), command)
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
	_, err := store.Complete(context.Background(), CompleteCommand{RunID: "run-1", IdempotencyKey: "complete-1", Status: "succeeded", ExpectedVersion: 1})
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
	state := &transactionState{seen: make(map[string]bool), version: 1, status: "running"}
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
