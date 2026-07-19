package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"
)

type smokeConnector struct{}

func (smokeConnector) Connect(context.Context) (driver.Conn, error) { return nil, errors.New("unused") }
func (smokeConnector) Driver() driver.Driver                        { return smokeConnector{} }
func (smokeConnector) Open(string) (driver.Conn, error)             { return nil, errors.New("unused") }
func TestPublicPoolSmoke(t *testing.T) {
	db := sql.OpenDB(smokeConnector{})
	defer db.Close()
	if err := ConfigurePool(db, PoolConfig{MaxOpen: 5, MaxIdle: 2, MaxLifetime: time.Minute, MaxIdleTime: time.Second}); err != nil {
		t.Fatal(err)
	}
	if db.Stats().MaxOpenConnections != 5 {
		t.Fatalf("pool stats = %#v", db.Stats())
	}
}
