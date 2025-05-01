package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

type TransactionService struct {
	txQueries database.TransactionQuerier
}

func NewTransactionService(txQueries database.TransactionQuerier) *TransactionService {
	return &TransactionService{txQueries: txQueries}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, req models.NewTransactionRequest) (*models.Transaction, error) {
	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	tx := models.NewTransaction(uuid.NewString(), userID, req.Date, req.Merchant, req.Amount, req.DetailedCategory)
	if err := s.txQueries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 tx.ID,
		UserID:             tx.UserID,
		TransactionDate:    tx.Date,
		Merchant:           tx.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: tx.DetailedCategory,
	}); err != nil {
		return nil, fmt.Errorf("unable to create transaction: %w", err)
	}
	return tx, nil
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, txnID, userID string, req models.NewTransactionRequest) (*models.Transaction, error) {

	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	txRow, err := s.txQueries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: req.DetailedCategory,
		UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		ID:                 txnID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error updating transaction: %w", err)
	}
	txn := models.Transaction{
		ID:               txRow.ID,
		UserID:           userID,
		Date:             txRow.TransactionDate,
		Merchant:         txRow.Merchant,
		Amount:           helpers.CentsToDollars(txRow.AmountCents),
		DetailedCategory: txRow.DetailedCategoryID,
	}

	return &txn, err
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, txnID, userID string) error {
	_, err := s.txQueries.DeleteTransactionByID(ctx, database.DeleteTransactionByIDParams{
		ID:     txnID,
		UserID: userID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("transaction not found: %w", err)
		}
		return fmt.Errorf("error deleting transaction: %w", err)
	}
	return nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID, txnID string) (*models.Transaction, error) {
	row, err := s.txQueries.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{UserID: userID, ID: txnID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error getting transaction by user_id, ID pair: %w", err)
	}
	txn := models.Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           helpers.CentsToDollars(row.AmountCents),
		DetailedCategory: row.DetailedCategoryID,
	}
	return &txn, nil
}

func (s *TransactionService) ListUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	cursorDate *string,
	cursorID *string,
	pageSize int64,
) (transactions []models.Transaction, nextCursorDate, nextCursorID string, hasMoreData bool, err error) {
	if pageSize <= 0 {
		err := errors.New("pageSize <= 0")
		return nil, "", "", false, fmt.Errorf("pageSize must be a positive integer: %w", err)
	}
	transactions = make([]models.Transaction, 0, pageSize)
	fetchSize := pageSize + 1
	if cursorDate == nil || cursorID == nil {
		firstPageRows, err := s.txQueries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{UserID: userID.String(), Limit: fetchSize})
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, "", "", false, fmt.Errorf("no transactions found: %w", err)
			}
			return nil, "", "", false, fmt.Errorf("error loading first page of transactions: %w", err)
		}
		for _, txn := range firstPageRows {
			transactions = append(transactions, s.ConvertFirstPageRow(txn))
		}
	} else {
		nextRows, err := s.txQueries.GetUserTransactionsPaginated(ctx, database.GetUserTransactionsPaginatedParams{
			UserID:          userID.String(),
			TransactionDate: *cursorDate,
			ID:              *cursorID,
			Limit:           fetchSize,
		})

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, "", "", false, fmt.Errorf("no transactions found: %w", err)
			}
			return nil, "", "", false, fmt.Errorf("error loading next page: %w", err)
		}
		for _, txn := range nextRows {
			transactions = append(transactions, s.ConvertPaginatedRow(txn))
		}

	}

	if int64(len(transactions)) > pageSize {
		hasMoreData = true
		lastTxn := transactions[pageSize-1]
		nextCursorDate = lastTxn.Date
		nextCursorID = lastTxn.ID
		transactions = transactions[:pageSize]

	} else {
		hasMoreData = false
	}

	return transactions, nextCursorDate, nextCursorID, hasMoreData, nil
}

// Helpers
func (s *TransactionService) ConvertFirstPageRow(row database.GetUserTransactionsFirstPageRow) models.Transaction {
	return models.Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           helpers.CentsToDollars(row.AmountCents),
		DetailedCategory: row.DetailedCategoryID,
	}
}

func (s *TransactionService) ConvertPaginatedRow(row database.GetUserTransactionsPaginatedRow) models.Transaction {
	return models.Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           helpers.CentsToDollars(row.AmountCents),
		DetailedCategory: row.DetailedCategoryID,
	}
}

func (s *TransactionService) ConvertToResponse(txn *models.Transaction) *models.TransactionResponse {
	return &models.TransactionResponse{
		Date:             txn.Date,
		Merchant:         txn.Merchant,
		Amount:           txn.Amount,
		DetailedCategory: txn.DetailedCategory,
	}
}
