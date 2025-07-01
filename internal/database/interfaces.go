package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

type UserQuerier interface {
	CreateUser(ctx context.Context, params CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
}

type DeviceQuerier interface {
	GetDeviceInfoByUser(ctx context.Context, arg GetDeviceInfoByUserParams) (uuid.UUID, error)
	CreateDeviceInfo(ctx context.Context, arg CreateDeviceInfoParams) (uuid.UUID, error)
}

type TokenQuerier interface {
	CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error
	RevokeToken(ctx context.Context, arg RevokeTokenParams) error
	GetRefreshByUserAndDevice(ctx context.Context, arg GetRefreshByUserAndDeviceParams) (models.RefreshToken, error)
}

type TransactionQuerier interface {
	CreateTransaction(ctx context.Context, arg CreateTransactionParams) error
	UpdateTransactionByID(ctx context.Context, arg UpdateTransactionByIDParams) (UpdateTransactionByIDRow, error)
	DeleteTransactionByID(ctx context.Context, arg DeleteTransactionByIDParams) (uuid.UUID, error)
	GetUserTransactionsFirstPage(ctx context.Context, arg GetUserTransactionsFirstPageParams) ([]GetUserTransactionsFirstPageRow, error)
	GetUserTransactionsPaginated(ctx context.Context, arg GetUserTransactionsPaginatedParams) ([]GetUserTransactionsPaginatedRow, error)
	GetUserTransactionByID(ctx context.Context, arg GetUserTransactionByIDParams) (GetUserTransactionByIDRow, error)
	GetPrimaryCategories(ctx context.Context) ([]models.PrimaryCategory, error)
	GetDetailedCategories(ctx context.Context) ([]models.DetailedCategory, error)
	GetDetailedCategoryID(ctx context.Context, name string) (int32, error)
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

type TxRow interface {
	GetFields() (id, userID uuid.UUID, date time.Time, merchant string, cents int64, catID int32)
}

func (r GetUserTransactionsFirstPageRow) GetFields() (uuid.UUID, uuid.UUID, time.Time, string, int64, int32) {
	return r.ID, r.UserID, r.TransactionDate, r.Merchant, r.AmountCents, r.DetailedCategoryID
}

func (r GetUserTransactionsPaginatedRow) GetFields() (uuid.UUID, uuid.UUID, time.Time, string, int64, int32) {
	return r.ID, r.UserID, r.TransactionDate, r.Merchant, r.AmountCents, r.DetailedCategoryID
}
