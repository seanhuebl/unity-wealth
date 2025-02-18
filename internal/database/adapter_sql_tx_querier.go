package database

import (
	"context"
	"database/sql"
)

type RealSqlTxQuerier struct {
	q *Queries
}

func NewRealSqlTxQuerier(q *Queries) SqlTxQuerier {
	return &RealSqlTxQuerier{
		q: q,
	}
}

func (rst *RealSqlTxQuerier) WithTx(tx *sql.Tx) *Queries {
	return rst.q.WithTx(tx)
}

func (rst *RealSqlTxQuerier) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return rst.q.BeginTx(ctx, opts)
}
