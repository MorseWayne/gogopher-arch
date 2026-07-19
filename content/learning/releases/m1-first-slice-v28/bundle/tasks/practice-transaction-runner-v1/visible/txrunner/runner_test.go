package txrunner

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

type state struct{ commits, rollbacks int }
type connector struct{ state *state }
type conn struct{ state *state }
type transaction struct{ state *state }

func (c connector) Connect(context.Context) (driver.Conn, error) { return &conn{state: c.state}, nil }
func (c connector) Driver() driver.Driver                        { return c }
func (c connector) Open(string) (driver.Conn, error)             { return &conn{state: c.state}, nil }
func (c *conn) Prepare(string) (driver.Stmt, error)              { return nil, errors.New("unused") }
func (c *conn) Close() error                                     { return nil }
func (c *conn) Begin() (driver.Tx, error)                        { return &transaction{state: c.state}, nil }
func (c *conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &transaction{state: c.state}, nil
}
func (tx *transaction) Commit() error   { tx.state.commits++; return nil }
func (tx *transaction) Rollback() error { tx.state.rollbacks++; return nil }

func TestWithinTxCommitAndRollback(t *testing.T) {
	current := &state{}
	db := sql.OpenDB(connector{state: current})
	defer db.Close()
	if err := WithinTx(context.Background(), db, nil, func(*sql.Tx) error { return nil }); err != nil {
		t.Fatal(err)
	}
	want := errors.New("stop")
	if err := WithinTx(context.Background(), db, nil, func(*sql.Tx) error { return want }); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
	if current.commits != 1 || current.rollbacks != 1 {
		t.Fatalf("commits=%d rollbacks=%d", current.commits, current.rollbacks)
	}
}
