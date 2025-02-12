package database

import (
	"context"
	"database/sql"

	"github.com/go-faster/errors"
)

func (q *Queries) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// Type assert that the underlying DBTX is a *sql.DB.
	db, ok := q.db.(*sql.DB)
	if !ok {
		return nil, errors.New("underlying DBTX is not a *sql.DB; cannot begin transaction")
	}
	return db.BeginTx(ctx, opts)
}
