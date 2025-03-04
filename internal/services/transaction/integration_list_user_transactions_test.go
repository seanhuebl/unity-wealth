package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIntegrationListUserTransactions(t *testing.T) {
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
			cursorDate:                   strPtr("2025-02-01"),
			expectedCursorDate:           "2025-02-10",
			cursorID:                     strPtr(uuid.NewString()),
			expectedHasMoreData:          true,
			pageSize:                     10,
			txSliceLength:                15,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
		{
			name:                         "paginated, no extra data, success",
			userID:                       uuid.New(),
			cursorDate:                   strPtr("2025-02-01"),
			expectedCursorDate:           "",
			cursorID:                     strPtr(uuid.NewString()),
			expectedHasMoreData:          false,
			pageSize:                     10,
			txSliceLength:                7,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fetchSize := tc.pageSize + 1
			expectedTxs := make([]Transaction, 0)
			ctx := context.Background()
			txID := uuid.New()
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_keys = ON")
			require.NoError(t, err)

			CreateTestingSchema(t, db)

			svc := NewTransactionService()

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
						expectedTxs = append(expectedTxs, svc.convertFirstPageRow(row))
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
						expectedTxs = append(expectedTxs, svc.convertPaginatedRow(row))
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
func integrationGenerateFirstPageRows(userID uuid.UUID, txSliceLength int) []database.GetUserTransactionsFirstPageRow {
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

func integrationGeneratePaginatedRows(userID uuid.UUID, txSliceLength int) []database.GetUserTransactionsPaginatedRow {
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

func integrationStrPtr(s string) *string {
	return &s
}
