package transaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, userID string, req models.NewTxRequest) (*models.Tx, error)
	UpdateTransaction(ctx context.Context, txnID, userID string, req models.NewTxRequest) (*models.Tx, error)
	DeleteTransaction(ctx context.Context, txnID, userID string) error
	GetTransactionByID(ctx context.Context, userID, txnID string) (*models.Tx, error)
	ListUserTransactions(
		ctx context.Context,
		userID uuid.UUID,
		cursorDate *string,
		cursorID *string,
		pageSize int64,
	) (transactions []models.Tx, nextCursorDate, nextCursorID string, hasMoreData bool, err error)
}
