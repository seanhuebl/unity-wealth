package transaction_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListUserTransactions(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                         string
		userID                       uuid.UUID
		cursorDate                   *string
		expectedCursorDate           string
		cursorID                     *string
		expectedHasMoreData          bool
		pageSize                     int64
		expectedPageSizeErrSubStr    string
		txSliceLength                int
		getFirstPageErr              error
		expectedFirstPageErrSubStr   string
		getTxPaginatedErr            error
		expectedTxPaginatedErrSubStr string
	}{
		{
			name:                       "first page, more data, success",
			userID:                     uuid.New(),
			cursorDate:                 nil,
			expectedCursorDate:         "2025-02-05",
			cursorID:                   nil,
			expectedHasMoreData:        true,
			pageSize:                   5,
			txSliceLength:              10,
			getFirstPageErr:            nil,
			expectedFirstPageErrSubStr: "",
		},
		{
			name:                       "first page, no extra data, success",
			userID:                     uuid.New(),
			cursorDate:                 nil,
			expectedCursorDate:         "",
			cursorID:                   nil,
			expectedHasMoreData:        false,
			pageSize:                   5,
			txSliceLength:              2,
			getFirstPageErr:            nil,
			expectedFirstPageErrSubStr: "",
		},
		{
			name:                         "paginated, more data, success",
			userID:                       uuid.New(),
			cursorDate:                   testhelpers.StrPtr("2025-02-01"),
			expectedCursorDate:           "2025-02-10",
			cursorID:                     testhelpers.StrPtr(uuid.NewString()),
			expectedHasMoreData:          true,
			pageSize:                     10,
			txSliceLength:                15,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
		{
			name:                         "paginated, no extra data, success",
			userID:                       uuid.New(),
			cursorDate:                   testhelpers.StrPtr("2025-02-01"),
			expectedCursorDate:           "",
			cursorID:                     testhelpers.StrPtr(uuid.NewString()),
			expectedHasMoreData:          false,
			pageSize:                     10,
			txSliceLength:                7,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
		{
			name:                       "first page, db error",
			userID:                     uuid.New(),
			cursorDate:                 nil,
			expectedCursorDate:         "",
			cursorID:                   nil,
			pageSize:                   1,
			getFirstPageErr:            errors.New("db error"),
			expectedFirstPageErrSubStr: "error loading first page of transactions",
		},
		{
			name:                       "first page, no transactions found",
			userID:                     uuid.New(),
			cursorDate:                 nil,
			expectedCursorDate:         "",
			cursorID:                   nil,
			pageSize:                   1,
			getFirstPageErr:            nil,
			expectedFirstPageErrSubStr: "",
		},
		{
			name:                         "paginated, db error",
			userID:                       uuid.New(),
			cursorDate:                   testhelpers.StrPtr("2025-02-01"),
			expectedCursorDate:           "",
			cursorID:                     testhelpers.StrPtr(uuid.NewString()),
			pageSize:                     1,
			getTxPaginatedErr:            errors.New("db error"),
			expectedTxPaginatedErrSubStr: "error loading next page",
		},
		{
			name:                         "paginated, no transactions found",
			userID:                       uuid.New(),
			cursorDate:                   testhelpers.StrPtr("2025-02-01"),
			expectedCursorDate:           "",
			cursorID:                     testhelpers.StrPtr(uuid.NewString()),
			pageSize:                     1,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
		{
			name:                      "page size <= 0",
			userID:                    uuid.New(),
			pageSize:                  0,
			expectedPageSizeErrSubStr: "pageSize must be a positive integer",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			fetchSize := tc.pageSize + 1
			expectedTxs := make([]models.Tx, 0)
			nopLogger := zap.NewNop()
			svc := transaction.NewTransactionService(mockTxQ, nopLogger)

			firstPageRows := generateFirstPageRows(tc.userID, tc.txSliceLength)

			if len(firstPageRows) > int(fetchSize) {
				firstPageRows = firstPageRows[:fetchSize]
			}
			if tc.cursorDate == nil || tc.cursorID == nil {
				mockTxQ.On("GetUserTransactionsFirstPage", ctx, mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).Return(firstPageRows, tc.getFirstPageErr).Maybe()
				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, tc.cursorDate, tc.cursorID, tc.pageSize)
				if tc.expectedPageSizeErrSubStr != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.expectedPageSizeErrSubStr)

				} else if tc.getFirstPageErr != nil {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.expectedFirstPageErrSubStr)
					mockTxQ.AssertExpectations(t)
				} else {
					require.NoError(t, err)
					if len(firstPageRows) > int(tc.pageSize) {
						firstPageRows = firstPageRows[:tc.pageSize]
					}
					for _, row := range firstPageRows {
						expectedTxs = append(expectedTxs, transaction.ConvertFirstPageRow(row))
					}
					if hasMoreData == true {
						require.NotEmpty(t, nextCursorID)
					}
					require.Equal(t, tc.expectedCursorDate, nextCursorDate)
					require.Equal(t, tc.expectedHasMoreData, hasMoreData)

					if diff := cmp.Diff(expectedTxs, transactions); diff != "" {
						t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
					}
					mockTxQ.AssertExpectations(t)
				}

			} else {
				nextRows := generatePaginatedRows(tc.userID, tc.txSliceLength)
				if len(nextRows) > int(fetchSize) {
					nextRows = nextRows[:fetchSize]
				}
				mockTxQ.On("GetUserTransactionsPaginated", ctx, mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).Return(nextRows, tc.getTxPaginatedErr).Maybe()
				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, tc.cursorDate, tc.cursorID, tc.pageSize)
				if tc.expectedPageSizeErrSubStr != "" {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.expectedPageSizeErrSubStr)
				} else if tc.getTxPaginatedErr != nil {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.expectedTxPaginatedErrSubStr)
					mockTxQ.AssertExpectations(t)
				} else {
					require.NoError(t, err)
					if len(nextRows) > int(tc.pageSize) {
						nextRows = nextRows[:tc.pageSize]
					}
					for _, row := range nextRows {
						expectedTxs = append(expectedTxs, transaction.ConvertPaginatedRow(row))
					}
					if hasMoreData == true {
						require.NotEmpty(t, nextCursorID)
					}
					require.Equal(t, tc.expectedCursorDate, nextCursorDate)
					require.Equal(t, tc.expectedHasMoreData, hasMoreData)

					if diff := cmp.Diff(expectedTxs, transactions); diff != "" {
						t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
					}
					mockTxQ.AssertExpectations(t)
				}
			}

		})
	}
}

// Helpers
func generateFirstPageRows(userID uuid.UUID, txSliceLength int) []database.GetUserTransactionsFirstPageRow {
	var rows []database.GetUserTransactionsFirstPageRow
	for i := 0; i < txSliceLength; i++ {
		date := fmt.Sprintf("2025-02-%02d", i+1)
		rows = append(rows, database.GetUserTransactionsFirstPageRow{
			ID:                 uuid.NewString(),
			UserID:             userID.String(),
			TransactionDate:    date,
			Merchant:           "costco",
			AmountCents:        15744,
			DetailedCategoryID: 40,
		})
	}
	return rows
}

func generatePaginatedRows(userID uuid.UUID, txSliceLength int) []database.GetUserTransactionsPaginatedRow {
	var rows []database.GetUserTransactionsPaginatedRow
	for i := 0; i < txSliceLength; i++ {
		date := fmt.Sprintf("2025-02-%02d", i+1)
		rows = append(rows, database.GetUserTransactionsPaginatedRow{
			ID:                 uuid.NewString(),
			UserID:             userID.String(),
			TransactionDate:    date,
			Merchant:           "costco",
			AmountCents:        15744,
			DetailedCategoryID: 40,
		})
	}
	return rows
}
