package sqlpool

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"
)

type connector struct{}

func (connector) Connect(context.Context) (driver.Conn, error) { return nil, errors.New("unused") }
func (connector) Driver() driver.Driver                        { return connector{} }
func (connector) Open(string) (driver.Conn, error)             { return nil, errors.New("unused") }

func TestConfigurePool(t *testing.T) {
	db := sql.OpenDB(connector{})
	defer db.Close()
	config := Config{MaxOpen: 8, MaxIdle: 3, MaxLifetime: time.Minute, MaxIdleTime: 15 * time.Second}
	if err := Configure(db, config); err != nil {
		t.Fatal(err)
	}
	if got := db.Stats().MaxOpenConnections; got != config.MaxOpen {
		t.Fatalf("MaxOpenConnections = %d", got)
	}
	for _, invalid := range []Config{{}, {MaxOpen: 1, MaxIdle: 2, MaxLifetime: time.Second, MaxIdleTime: time.Second}} {
		if err := Configure(db, invalid); err == nil {
			t.Fatalf("accepted invalid config %#v", invalid)
		}
	}
}
