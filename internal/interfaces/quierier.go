package interfaces

import (
	"context"
	"database/sql"

	"github.com/seanhuebl/unity-wealth/internal/database"
)

type Quierier interface {
	CreateUser(ctx context.Context, params database.CreateUserParams) error
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	RevokeToken(ctx context.Context, arg database.RevokeTokenParams) error
	GetDeviceInfoByUser(ctx context.Context, arg database.GetDeviceInfoByUserParams) (string, error)
	CreateRefreshToken(ctx context.Context, arg database.CreateRefreshTokenParams) error
	CreateDeviceInfo(ctx context.Context, arg database.CreateDeviceInfoParams) (string, error)
	WithTx(tx *sql.Tx) *database.Queries
	CreateTransaction(ctx context.Context, arg database.CreateTransactionParams) error
	GetDetailedCategoryId(ctx context.Context, name string) (int64, error)
	UpdateTransactionByID(ctx context.Context, arg database.UpdateTransactionByIDParams) (database.UpdateTransactionByIDRow, error)
	GetPrimaryCategories(ctx context.Context) ([]database.PrimaryCategory, error)
	GetDetailedCategories(ctx context.Context) ([]database.DetailedCategory, error)
	DeleteTransactionById(ctx context.Context, arg database.DeleteTransactionByIdParams) error
}
