package transaction

import (
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
)

type FirstPageRowWrapper struct {
	database.GetUserTransactionsFirstPageRow
}

func (w FirstPageRowWrapper) GetTxID() uuid.UUID {
	return uuid.MustParse(w.GetUserTransactionsFirstPageRow.ID)
}

func (w FirstPageRowWrapper) GetUserID() uuid.UUID {
	return uuid.MustParse(w.GetUserTransactionsFirstPageRow.UserID)
}

func (w FirstPageRowWrapper) GetTxDate() string {
	return w.GetUserTransactionsFirstPageRow.TransactionDate
}

func (w FirstPageRowWrapper) GetMerchant() string {
	return w.GetUserTransactionsFirstPageRow.Merchant
}

func (w FirstPageRowWrapper) GetAmountCents() int64 {
	return w.GetUserTransactionsFirstPageRow.AmountCents
}

func (w FirstPageRowWrapper) GetDetailedCatID() int64 {
	return w.GetUserTransactionsFirstPageRow.DetailedCategoryID
}

type PaginatedRowWrapper struct {
	database.GetUserTransactionsPaginatedRow
}

func (w PaginatedRowWrapper) GetTxID() uuid.UUID {
	return uuid.MustParse(w.GetUserTransactionsPaginatedRow.ID)
}

func (w PaginatedRowWrapper) GetUserID() uuid.UUID {
	return uuid.MustParse(w.GetUserTransactionsPaginatedRow.UserID)
}

func (w PaginatedRowWrapper) GetTxDate() string {
	return w.GetUserTransactionsPaginatedRow.TransactionDate
}

func (w PaginatedRowWrapper) GetMerchant() string {
	return w.GetUserTransactionsPaginatedRow.Merchant
}

func (w PaginatedRowWrapper) GetAmountCents() int64 {
	return w.GetUserTransactionsPaginatedRow.AmountCents
}

func (w PaginatedRowWrapper) GetDetailedCatID() int64 {
	return w.GetUserTransactionsPaginatedRow.DetailedCategoryID
}

func WrapFirstPageRows(rows []database.GetUserTransactionsFirstPageRow) []FirstPageRowWrapper {
	wrapped := make([]FirstPageRowWrapper, len(rows))
	for i, r := range rows {
		wrapped[i] = FirstPageRowWrapper{r}
	}
	return wrapped
}

func WrapPaginatedRows(rows []database.GetUserTransactionsPaginatedRow) []PaginatedRowWrapper {
	wrapped := make([]PaginatedRowWrapper, len(rows))
	for i, r := range rows {
		wrapped[i] = PaginatedRowWrapper{r}
	}
	return wrapped
}
