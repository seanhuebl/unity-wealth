package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/models"
)

type TransactionService struct {
	queries interfaces.Querier
}

func NewTransactionService(queries interfaces.Querier) *TransactionService {
	return &TransactionService{queries: queries}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, req models.NewTransactionRequest) (*models.Transaction, error) {
	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	tx := models.NewTransaction(uuid.NewString(), userID, req.Date, req.Merchant, req.Amount, req.DetailedCategory)
	if err := s.queries.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 tx.ID,
		UserID:             tx.UserID,
		TransactionDate:    tx.Date,
		Merchant:           tx.Merchant,
		AmountCents:        int64(tx.Amount * 100),
		DetailedCategoryID: tx.DetailedCategory,
	}); err != nil {
		return nil, fmt.Errorf("unable to create transaction: %w", err)
	}
	return tx, nil
}

func (s *TransactionService) UpdateTransaction(ctx context.Context, txnID string, req models.NewTransactionRequest) (*models.Transaction, error) {

	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	txRow, err := s.queries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
		DetailedCategoryID: req.DetailedCategory,
		UpdatedAt:          sql.NullTime{Time: time.Now(), Valid: true},
		ID:                 txnID,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating transaction: %w", err)
	}
	txn := models.Transaction{
		ID: txRow.ID,
		UserID: "",
		Date: txRow.TransactionDate,
		Merchant: txRow.Merchant,
		Amount: float64(txRow.AmountCents) / 100,
		DetailedCategory: txRow.DetailedCategoryID,
	}

	return &txn, err
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, txnID, userID string) (map[string]any, error) {
	if err := s.queries.DeleteTransactionById(ctx, database.DeleteTransactionByIdParams{
		ID:     txnID,
		UserID: userID,
	}); err != nil {
		return nil, fmt.Errorf("error deleting transaction: %w", err)
	}
	response := make(map[string]any)
	response["transaction_deleted"] = "success"
	return response, nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID, txnID string) (*models.Transaction, error) {
	row, err := s.queries.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{UserID: userID, ID: txnID})
	if err != nil {
		return nil, fmt.Errorf("error getting transaction by user_id, ID pair: %w", err)
	}
	txn := models.Transaction{
		ID: row.ID,
		UserID: row.UserID,
		Date: row.TransactionDate,
		Merchant: row.Merchant,
		Amount: float64(row.AmountCents) / 100,
		DetailedCategory: row.DetailedCategoryID,
	}
	return &txn, nil
}

// Helpers
func (s *TransactionService) ListUserTransactions(
	ctx context.Context,
	userID uuid.UUID,
	cursorDate *string,
	cursorID *string,
	pageSize *int64,
) (transactions []models.Transaction, nextCursorDate, nextCursorID string, hasMoreData bool, err error) {

	transactions = make([]models.Transaction, 0, *pageSize)
	fetchSize := *pageSize + 1
	if cursorDate == nil || cursorID == nil {
		firstPageRows, err := s.queries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{UserID: userID.String(), Limit: fetchSize})
		if err != nil {
			return nil, "", "", false, fmt.Errorf("error loading first page of transactions: %w", err)
		}
		for _, txn := range firstPageRows {
			transactions = append(transactions, s.convertFirstPageRow(txn))
		}
	} else {
		nextRows, err := s.queries.GetUserTransactionsPaginated(ctx, database.GetUserTransactionsPaginatedParams{
			UserID:          userID.String(),
			TransactionDate: *cursorDate,
			ID:              *cursorID,
			Limit:           fetchSize,
		})
		if err != nil {
			return nil, "", "", false, fmt.Errorf("error loading next page: %w", err)
		}
		for _, txn := range nextRows {
			transactions = append(transactions, s.convertPaginatedRow(txn))
		}

	}

	if int64(len(transactions)) > *pageSize {
		hasMoreData = true
		lastTxn := transactions[*pageSize-1]
		nextCursorDate = lastTxn.Date
		nextCursorID = lastTxn.ID
		transactions = transactions[:*pageSize]

	} else {
		hasMoreData = false
	}

	return transactions, nextCursorDate, nextCursorID, hasMoreData, nil
}

func (s *TransactionService) convertFirstPageRow(row database.GetUserTransactionsFirstPageRow) models.Transaction {
	return models.Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           float64(row.AmountCents) / 100,
		DetailedCategory: row.DetailedCategoryID,
	}
}

func (s *TransactionService) convertPaginatedRow(row database.GetUserTransactionsPaginatedRow) models.Transaction {
	return models.Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           float64(row.AmountCents) / 100,
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