package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/stretchr/testify/require"
)

func TestIntegrationListUserTransactions(t *testing.T) {
	tests := []struct {
		name                         string
		userID                       uuid.UUID
		isFirstPage                  bool
		expectedCursorDate           string
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
			isFirstPage:                true,
			expectedCursorDate:         "2025-02-05",
			expectedHasMoreData:        true,
			pageSize:                   5,
			txSliceLength:              10,
			getFirstPageErr:            nil,
			expectedFirstPageErrSubStr: "",
		},
		{
			name:                       "first page, no extra data, success",
			userID:                     uuid.New(),
			isFirstPage:                true,
			expectedCursorDate:         "",
			expectedHasMoreData:        false,
			pageSize:                   5,
			txSliceLength:              2,
			getFirstPageErr:            nil,
			expectedFirstPageErrSubStr: "",
		},
		{
			name:                         "paginated, more data, success",
			userID:                       uuid.New(),
			expectedCursorDate:           "2025-02-11",
			expectedHasMoreData:          true,
			pageSize:                     10,
			txSliceLength:                15,
			getTxPaginatedErr:            nil,
			expectedTxPaginatedErrSubStr: "",
		},
		{
			name:                         "paginated, no extra data, success",
			userID:                       uuid.New(),
			expectedCursorDate:           "",
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
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_keys = ON")
			require.NoError(t, err)

			CreateTestingSchema(t, db)
			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			txQ := database.NewRealTransactionQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)
			seedListUserTxTestData(t, db, userQ, tc.userID)

			firstPageRows := integrationGenerateFirstPageRows(tc.userID, tc.txSliceLength)

			if len(firstPageRows) > int(fetchSize) {
				firstPageRows = firstPageRows[:fetchSize]
			}
			if tc.isFirstPage {
				wrappedFirstPageRows := WrapFirstPageRows(firstPageRows)
				SeedMultipleTestTransactions(t, txQ, wrappedFirstPageRows)
				svc := NewTransactionService(txQ)

				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, nil, nil, tc.pageSize)
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

				if diff := cmp.Diff(expectedTxs, transactions, cmpopts.IgnoreFields(Transaction{}, "ID")); diff != "" {
					t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
				}

			} else {
				nextRows := integrationGeneratePaginatedRows(tc.userID, tc.txSliceLength)
				wrappedPaginatedRows := WrapPaginatedRows(nextRows)
				SeedMultipleTestTransactions(t, txQ, wrappedPaginatedRows)

				if len(nextRows) > int(fetchSize) {
					nextRows = nextRows[:fetchSize]
				}
				svc := NewTransactionService(txQ)
				cursorDate := wrappedPaginatedRows[0].GetTxDate()
				cursorID := wrappedPaginatedRows[0].GetTxID().String()
				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, &cursorDate, &cursorID, tc.pageSize)
				fmt.Println(transactions)
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
				if diff := cmp.Diff(expectedTxs, transactions, cmpopts.IgnoreFields(Transaction{}, "ID")); diff != "" {
					t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
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

func seedListUserTxTestData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID) {
	SeedTestUser(t, userQ, userID)
	SeedTestCategories(t, db)
}
