package database

import (
	"context"
	"database/sql"
)

type UserQuerier interface {
	CreateUser(ctx context.Context, params CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
}

type DeviceQuerier interface {
	GetDeviceInfoByUser(ctx context.Context, arg GetDeviceInfoByUserParams) (string, error)
	CreateDeviceInfo(ctx context.Context, arg CreateDeviceInfoParams) (string, error)
}

type TokenQuerier interface {
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error
	RevokeToken(ctx context.Context, arg RevokeTokenParams) error
	GetRefreshByUserAndDevice(ctx context.Context, arg GetRefreshByUserAndDeviceParams) (RefreshToken, error)
}

type TransactionQuerier interface {
	CreateTransaction(ctx context.Context, arg CreateTransactionParams) error
	UpdateTransactionByID(ctx context.Context, arg UpdateTransactionByIDParams) (UpdateTransactionByIDRow, error)
	DeleteTransactionByID(ctx context.Context, arg DeleteTransactionByIDParams) (string, error)
	GetUserTransactionsFirstPage(ctx context.Context, arg GetUserTransactionsFirstPageParams) ([]GetUserTransactionsFirstPageRow, error)
	GetUserTransactionsPaginated(ctx context.Context, arg GetUserTransactionsPaginatedParams) ([]GetUserTransactionsPaginatedRow, error)
	GetUserTransactionByID(ctx context.Context, arg GetUserTransactionByIDParams) (GetUserTransactionByIDRow, error)
	GetPrimaryCategories(ctx context.Context) ([]PrimaryCategory, error)
	GetDetailedCategories(ctx context.Context) ([]DetailedCategory, error)
	GetDetailedCategoryID(ctx context.Context, name string) (int64, error)
}

type SqlTxQuerier interface {
	WithTx(tx *sql.Tx) SqlTransactionalQuerier
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SqlTransactionalQuerier interface {
	SqlTxQuerier
	DeviceQuerier
	TokenQuerier
	TransactionQuerier
	UserQuerier
}
