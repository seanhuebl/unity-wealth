package database

import "context"

type RealTransactionQuerier struct {
	q *Queries
}

func NewRealTransactionQuerier(q *Queries) RealTransactionQuerier {
	return &RealTransactionQuerier{
		q: q,
	}
}

func (rt *RealTransactionQuerier) CreateTransaction(ctx context.Context, arg CreateTransactionParams) error {
	return rt.q.CreateTransaction(ctx, arg)
}

func (rt *RealTransactionQuerier) UpdateTransactionByID(ctx context.Context, arg UpdateTransactionByIDParams) (UpdateTransactionByIDRow, error) {
	return rt.q.UpdateTransactionByID(ctx, arg)
}

func (rt *RealTransactionQuerier) DeleteTransactionByID(ctx context.Context, arg DeleteTransactionByIDParams) error {
	return rt.q.DeleteTransactionByID(ctx, arg)
}

func (rt RealTransactionQuerier) GetUserTransactionsFirstPage(ctx context.Context, arg GetUserTransactionsFirstPageParams) ([]GetUserTransactionsFirstPageRow, error) {
	return rt.q.GetUserTransactionsFirstPage(ctx, arg)
}
