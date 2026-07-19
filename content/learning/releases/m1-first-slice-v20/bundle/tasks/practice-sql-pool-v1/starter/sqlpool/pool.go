package sqlpool

import (
	"database/sql"
	"errors"
	"time"
)

type Config struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}

func Configure(db *sql.DB, config Config) error {
	return errors.New("TODO: validate and configure the connection pool")
}
