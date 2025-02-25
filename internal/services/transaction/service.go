package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type TransactionService struct {
	txQueries database.TransactionQuerier
}

func NewTransactionService(txQueries database.TransactionQuerier) *TransactionService {
	return &TransactionService{txQueries: txQueries}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, userID string, req NewTransactionRequest) (*Transaction, error) {
	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	tx := NewTransaction(uuid.NewString(), userID, req.Date, req.Merchant, req.Amount, req.DetailedCategory)
	if err := s.txQueries.CreateTransaction(ctx, database.CreateTransactionParams{
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

func (s *TransactionService) UpdateTransaction(ctx context.Context, txnID, userID string, req NewTransactionRequest) (*Transaction, error) {

	_, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	txRow, err := s.txQueries.UpdateTransactionByID(ctx, database.UpdateTransactionByIDParams{
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        int64(req.Amount * 100),
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
	txn := Transaction{
		ID:               txRow.ID,
		UserID:           userID,
		Date:             txRow.TransactionDate,
		Merchant:         txRow.Merchant,
		Amount:           float64(txRow.AmountCents) / 100.0,
		DetailedCategory: txRow.DetailedCategoryID,
	}

	return &txn, err
}

func (s *TransactionService) DeleteTransaction(ctx context.Context, txnID, userID string) error {
	if err := s.txQueries.DeleteTransactionByID(ctx, database.DeleteTransactionByIDParams{
		ID:     txnID,
		UserID: userID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no transaction found: %w", err)
		}
		return fmt.Errorf("error deleting transaction: %w", err)
	}
	return nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, userID, txnID string) (*Transaction, error) {
	row, err := s.txQueries.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{UserID: userID, ID: txnID})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("error getting transaction by user_id, ID pair: %w", err)
	}
	txn := Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           float64(row.AmountCents) / 100.0,
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
) (transactions []Transaction, nextCursorDate, nextCursorID string, hasMoreData bool, err error) {

	transactions = make([]Transaction, 0, pageSize)
	fetchSize := pageSize + 1
	if cursorDate == nil || cursorID == nil {
		firstPageRows, err := s.txQueries.GetUserTransactionsFirstPage(ctx, database.GetUserTransactionsFirstPageParams{UserID: userID.String(), Limit: fetchSize})
		if err != nil {
			return nil, "", "", false, fmt.Errorf("error loading first page of transactions: %w", err)
		}
		for _, txn := range firstPageRows {
			transactions = append(transactions, s.convertFirstPageRow(txn))
		}
	} else {
		nextRows, err := s.txQueries.GetUserTransactionsPaginated(ctx, database.GetUserTransactionsPaginatedParams{
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
func (s *TransactionService) convertFirstPageRow(row database.GetUserTransactionsFirstPageRow) Transaction {
	return Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           float64(row.AmountCents) / 100.0,
		DetailedCategory: row.DetailedCategoryID,
	}
}

func (s *TransactionService) convertPaginatedRow(row database.GetUserTransactionsPaginatedRow) Transaction {
	return Transaction{
		ID:               row.ID,
		UserID:           row.UserID,
		Date:             row.TransactionDate,
		Merchant:         row.Merchant,
		Amount:           float64(row.AmountCents) / 100.0,
		DetailedCategory: row.DetailedCategoryID,
	}
}

func (s *TransactionService) ConvertToResponse(txn *Transaction) *TransactionResponse {
	return &TransactionResponse{
		Date:             txn.Date,
		Merchant:         txn.Merchant,
		Amount:           txn.Amount,
		DetailedCategory: txn.DetailedCategory,
	}
}
