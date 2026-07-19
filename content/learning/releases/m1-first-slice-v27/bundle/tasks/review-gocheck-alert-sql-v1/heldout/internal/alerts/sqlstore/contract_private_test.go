package sqlstore_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"gocheckhub/internal/alerts"
	"gocheckhub/internal/alerts/sqlstore"
)

var register sync.Once
var current *driverState

type driverState struct {
	mu      sync.Mutex
	query   string
	args    []driver.NamedValue
	values  [][]driver.Value
	columns []string
	nextErr error
	closed  int
	block   bool
}
type testDriver struct{}
type testConn struct{}
type testRows struct {
	state *driverState
	index int
}

func (testDriver) Open(string) (driver.Conn, error)  { return testConn{}, nil }
func (testConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("prepare not supported") }
func (testConn) Close() error                        { return nil }
func (testConn) Begin() (driver.Tx, error)           { return nil, errors.New("tx not supported") }
func (testConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	current.record(query, args)
	if current.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return driver.RowsAffected(1), nil
}
func (testConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	current.record(query, args)
	return &testRows{state: current}, nil
}
func (s *driverState) record(query string, args []driver.NamedValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = query
	s.args = append([]driver.NamedValue(nil), args...)
}
func (r *testRows) Columns() []string { return r.state.columns }
func (r *testRows) Close() error {
	r.state.mu.Lock()
	r.state.closed++
	r.state.mu.Unlock()
	return nil
}
func (r *testRows) Next(dest []driver.Value) error {
	if r.index < len(r.state.values) {
		copy(dest, r.state.values[r.index])
		r.index++
		return nil
	}
	if r.state.nextErr != nil {
		err := r.state.nextErr
		r.state.nextErr = nil
		return err
	}
	return io.EOF
}
func openDB(t *testing.T, state *driverState) *sql.DB {
	t.Helper()
	register.Do(func() { sql.Register("m204-alerts", testDriver{}) })
	current = state
	db, err := sql.Open("m204-alerts", "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestPoolContract(t *testing.T) {
	db := openDB(t, &driverState{})
	config := sqlstore.PoolConfig{MaxOpen: 7, MaxIdle: 3, MaxLifetime: time.Minute, MaxIdleTime: time.Second}
	if err := sqlstore.ConfigurePool(db, config); err != nil {
		t.Fatal(err)
	}
	if db.Stats().MaxOpenConnections != 7 {
		t.Fatal("MaxOpen not configured")
	}
	if err := sqlstore.ConfigurePool(db, sqlstore.PoolConfig{}); err == nil {
		t.Fatal("accepted invalid pool")
	}
	if _, err := sqlstore.NewRepository(nil); err == nil {
		t.Fatal("accepted nil DB")
	}
}
func TestContextQueriesAndArguments(t *testing.T) {
	state := &driverState{}
	db := openDB(t, state)
	repository, err := sqlstore.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(context.Background(), alerts.Rule{ID: "c-1", Destination: "api"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToUpper(state.query), "INSERT") || len(state.args) != 2 || state.args[0].Value != "c-1" || state.args[1].Value != "api" {
		t.Fatalf("query=%q args=%#v", state.query, state.args)
	}
}
func TestRowsLifecycleAndIterationErrors(t *testing.T) {
	state := &driverState{columns: []string{"id", "target"}, values: [][]driver.Value{{"c-1", "api"}, {"c-2", "worker"}}}
	db := openDB(t, state)
	repository, _ := sqlstore.NewRepository(db)
	values, err := repository.List(context.Background())
	if err != nil || len(values) != 2 || values[1].Destination != "worker" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
	if state.closed != 1 {
		t.Fatalf("Rows.Close calls=%d", state.closed)
	}
	sentinel := errors.New("iteration failed")
	state.values = nil
	state.nextErr = sentinel
	if _, err := repository.List(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("iteration error=%v", err)
	}
	if state.closed != 2 {
		t.Fatalf("Rows.Close calls=%d", state.closed)
	}
}
func TestSingleRowErrorsStable(t *testing.T) {
	state := &driverState{columns: []string{"id", "target"}}
	db := openDB(t, state)
	repository, _ := sqlstore.NewRepository(db)
	if _, err := repository.Find(context.Background(), "missing"); !errors.Is(err, alerts.ErrNotFound) {
		t.Fatalf("missing error=%v", err)
	}
	state.values = [][]driver.Value{{"c-bad", nil}}
	if _, err := repository.Find(context.Background(), "bad"); err == nil || errors.Is(err, alerts.ErrNotFound) {
		t.Fatalf("scan error=%v", err)
	}
	state.values = [][]driver.Value{{"c-3", "cron"}}
	value, err := repository.Find(context.Background(), "c-3")
	if err != nil || value.Destination != "cron" {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
func TestQueryCancellationPropagates(t *testing.T) {
	state := &driverState{block: true}
	db := openDB(t, state)
	repository, _ := sqlstore.NewRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repository.Save(ctx, alerts.Rule{ID: "c", Destination: "api"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel error=%v", err)
	}
}
