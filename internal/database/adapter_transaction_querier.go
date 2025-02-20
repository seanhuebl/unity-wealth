package database

import "context"

type RealTransactionQuerier struct {
	q SqlTransactionalQuerier
}

func NewRealTransactionQuerier(q SqlTransactionalQuerier) TransactionQuerier {
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

func (rt *RealTransactionQuerier) GetUserTransactionsFirstPage(ctx context.Context, arg GetUserTransactionsFirstPageParams) ([]GetUserTransactionsFirstPageRow, error) {
	return rt.q.GetUserTransactionsFirstPage(ctx, arg)
}

func (rt *RealTransactionQuerier) GetUserTransactionsPaginated(ctx context.Context, arg GetUserTransactionsPaginatedParams) ([]GetUserTransactionsPaginatedRow, error) {
	return rt.q.GetUserTransactionsPaginated(ctx, arg)
}

func (rt *RealTransactionQuerier) GetUserTransactionByID(ctx context.Context, arg GetUserTransactionByIDParams) (GetUserTransactionByIDRow, error) {
	return rt.q.GetUserTransactionByID(ctx, arg)
}

func (rt *RealTransactionQuerier) GetPrimaryCategories(ctx context.Context) ([]PrimaryCategory, error) {
	return rt.q.GetPrimaryCategories(ctx)
}

func (rt *RealTransactionQuerier) GetDetailedCategories(ctx context.Context) ([]DetailedCategory, error) {
	return rt.q.GetDetailedCategories(ctx)
}

func (rt *RealTransactionQuerier) GetDetailedCategoryID(ctx context.Context, name string) (int64, error) {
	return rt.q.GetDetailedCategoryID(ctx, name)
}
