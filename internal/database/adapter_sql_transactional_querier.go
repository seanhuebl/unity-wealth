package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

// RealTransactionalQuerier wraps a *Queries to implement SqlTransactionalQuerier.
type RealTransactionalQuerier struct {
	q *Queries
}

// NewRealTransactionalQuerier returns an adapter wrapping the given *Queries.
func NewRealTransactionalQuerier(q *Queries) SqlTransactionalQuerier {
	return &RealTransactionalQuerier{q: q}
}

// WithTx calls the underlying WithTx method and wraps the result.
func (r *RealTransactionalQuerier) WithTx(tx *sql.Tx) SqlTransactionalQuerier {
	return NewRealTransactionalQuerier(r.q.WithTx(tx))
}

// BeginTx delegates to the underlying *Queries.
func (r *RealTransactionalQuerier) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return r.q.BeginTx(ctx, opts)
}

// DeviceQuerier methods.
func (r *RealTransactionalQuerier) GetDeviceInfoByUser(ctx context.Context, arg GetDeviceInfoByUserParams) (uuid.UUID, error) {
	return r.q.GetDeviceInfoByUser(ctx, arg)
}

func (r *RealTransactionalQuerier) CreateDeviceInfo(ctx context.Context, arg CreateDeviceInfoParams) (uuid.UUID, error) {
	return r.q.CreateDeviceInfo(ctx, arg)
}

// TokenQuerier methods.
func (r *RealTransactionalQuerier) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	return r.q.CreateRefreshToken(ctx, arg)
}

func (r *RealTransactionalQuerier) RevokeToken(ctx context.Context, arg RevokeTokenParams) error {
	return r.q.RevokeToken(ctx, arg)
}

func (r *RealTransactionalQuerier) GetRefreshByUserAndDevice(ctx context.Context, arg GetRefreshByUserAndDeviceParams) (models.RefreshToken, error) {
	return r.q.GetRefreshByUserAndDevice(ctx, arg)
}

// TransactionQuerier methods.
func (r *RealTransactionalQuerier) CreateTransaction(ctx context.Context, arg CreateTransactionParams) error {
	return r.q.CreateTransaction(ctx, arg)
}

func (r *RealTransactionalQuerier) UpdateTransactionByID(ctx context.Context, arg UpdateTransactionByIDParams) (UpdateTransactionByIDRow, error) {
	return r.q.UpdateTransactionByID(ctx, arg)
}

func (r *RealTransactionalQuerier) DeleteTransactionByID(ctx context.Context, arg DeleteTransactionByIDParams) (uuid.UUID, error) {
	return r.q.DeleteTransactionByID(ctx, arg)
}

func (r *RealTransactionalQuerier) GetUserTransactionsFirstPage(ctx context.Context, arg GetUserTransactionsFirstPageParams) ([]GetUserTransactionsFirstPageRow, error) {
	return r.q.GetUserTransactionsFirstPage(ctx, arg)
}

func (r *RealTransactionalQuerier) GetUserTransactionsPaginated(ctx context.Context, arg GetUserTransactionsPaginatedParams) ([]GetUserTransactionsPaginatedRow, error) {
	return r.q.GetUserTransactionsPaginated(ctx, arg)
}

func (r *RealTransactionalQuerier) GetUserTransactionByID(ctx context.Context, arg GetUserTransactionByIDParams) (GetUserTransactionByIDRow, error) {
	return r.q.GetUserTransactionByID(ctx, arg)
}

func (r *RealTransactionalQuerier) GetPrimaryCategories(ctx context.Context) ([]models.PrimaryCategory, error) {
	return r.q.GetPrimaryCategories(ctx)
}

func (r *RealTransactionalQuerier) GetDetailedCategories(ctx context.Context) ([]models.DetailedCategory, error) {
	return r.q.GetDetailedCategories(ctx)
}

func (r *RealTransactionalQuerier) GetDetailedCategoryID(ctx context.Context, name string) (int32, error) {
	return r.q.GetDetailedCategoryID(ctx, name)
}

// User methods

func (r *RealTransactionalQuerier) CreateUser(ctx context.Context, arg CreateUserParams) error {
	return r.q.CreateUser(ctx, arg)
}

func (r *RealTransactionalQuerier) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	return r.q.GetUserByEmail(ctx, email)
}
