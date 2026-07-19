package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

type rejectedConnector struct{}

func (rejectedConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("unexpected database call")
}
func (rejectedConnector) Driver() driver.Driver { return rejectedConnector{} }
func (rejectedConnector) Open(string) (driver.Conn, error) {
	return nil, errors.New("unexpected database call")
}

func TestStoreValidatesBeforeTransaction(t *testing.T) {
	if _, err := NewStore(nil); err == nil {
		t.Fatal("NewStore accepted nil DB")
	}
	db := sql.OpenDB(rejectedConnector{})
	defer db.Close()
	store, err := NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Complete(context.Background(), CompleteCommand{}); err == nil {
		t.Fatal("Complete accepted empty command")
	}
}
