package transaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, req models.NewTxRequest) (*models.Tx, error)
	UpdateTransaction(ctx context.Context, txnID, userID uuid.UUID, req models.NewTxRequest) (*models.Tx, error)
	DeleteTransaction(ctx context.Context, txnID, userID uuid.UUID) error
	GetTransactionByID(ctx context.Context, userID, txnID uuid.UUID) (*models.Tx, error)
	ListUserTransactions(ctx context.Context, userID uuid.UUID, cursorToken string, pageSize int32) (transaction.ListTxResult, error)
}
