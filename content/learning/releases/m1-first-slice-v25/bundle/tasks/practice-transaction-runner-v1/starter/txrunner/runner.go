package txrunner

import (
	"context"
	"database/sql"
	"errors"
)

func WithinTx(ctx context.Context, db *sql.DB, options *sql.TxOptions, fn func(*sql.Tx) error) error {
	return errors.New("TODO: run the callback in a transaction")
}
