package database

import (
	"context"
	"database/sql"
)

type RealSqlTxQuerier struct {
	q SqlTransactionalQuerier
}

func NewRealSqlTxQuerier(q SqlTransactionalQuerier) SqlTxQuerier {
	return &RealSqlTxQuerier{
		q: q,
	}
}

func (rst *RealSqlTxQuerier) WithTx(tx *sql.Tx) SqlTransactionalQuerier {
	return rst.q.WithTx(tx)
}

func (rst *RealSqlTxQuerier) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return rst.q.BeginTx(ctx, opts)
}
