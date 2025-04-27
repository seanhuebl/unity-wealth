package transaction_test

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
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
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
			expectedTxs := make([]models.Transaction, 0)
			ctx := context.Background()
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_keys = ON")
			require.NoError(t, err)

			testhelpers.CreateTestingSchema(t, db)
			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			txQ := database.NewRealTransactionQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)
			seedListUserTxTestData(t, db, userQ, tc.userID)

			firstPageRows := integrationGenerateFirstPageRows(tc.userID, tc.txSliceLength)

			if len(firstPageRows) > int(fetchSize) {
				firstPageRows = firstPageRows[:fetchSize]
			}
			if tc.isFirstPage {
				wrappedFirstPageRows := transaction.WrapFirstPageRows(firstPageRows)
				testhelpers.SeedMultipleTestTransactions(t, txQ, wrappedFirstPageRows)
				svc := transaction.NewTransactionService(txQ)

				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, nil, nil, tc.pageSize)
				require.NoError(t, err)
				if len(firstPageRows) > int(tc.pageSize) {
					firstPageRows = firstPageRows[:tc.pageSize]
				}

				for _, row := range firstPageRows {
					expectedTxs = append(expectedTxs, svc.ConvertFirstPageRow(row))
				}
				if hasMoreData == true {
					require.NotEmpty(t, nextCursorID)
				}
				require.Equal(t, tc.expectedCursorDate, nextCursorDate)
				require.Equal(t, tc.expectedHasMoreData, hasMoreData)

				if diff := cmp.Diff(expectedTxs, transactions, cmpopts.IgnoreFields(models.Transaction{}, "ID")); diff != "" {
					t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
				}

			} else {
				paginatedRows := integrationGeneratePaginatedRows(tc.userID, tc.txSliceLength)
				wrappedPaginatedRows := transaction.WrapPaginatedRows(paginatedRows)
				testhelpers.SeedMultipleTestTransactions(t, txQ, wrappedPaginatedRows)

				if len(paginatedRows) > int(fetchSize) {
					paginatedRows = paginatedRows[:fetchSize]
				}
				svc := transaction.NewTransactionService(txQ)
				cursorDate := wrappedPaginatedRows[0].GetTxDate()
				cursorID := wrappedPaginatedRows[0].GetTxID().String()
				transactions, nextCursorDate, nextCursorID, hasMoreData, err := svc.ListUserTransactions(ctx, tc.userID, &cursorDate, &cursorID, tc.pageSize)
				require.NoError(t, err)
				fmt.Println(len(paginatedRows))
				if len(paginatedRows) > int(tc.pageSize) {
					// Page size +1 because we are
					// Simulating the behavior where the query fetches one extra row to determine if there’s more data
					// Then we discard the first row (used as the cursor) and use the remaining rows as the expected transactions
					paginatedRows = paginatedRows[:tc.pageSize+1]
				}
				for i := 1; i < len(paginatedRows); i++ {
					expectedTxs = append(expectedTxs, svc.ConvertPaginatedRow(paginatedRows[i]))
				}
				if hasMoreData == true {
					require.NotEmpty(t, nextCursorID)
				}
				require.Equal(t, tc.expectedCursorDate, nextCursorDate)
				require.Equal(t, tc.expectedHasMoreData, hasMoreData)
				if diff := cmp.Diff(expectedTxs, transactions, cmpopts.IgnoreFields(models.Transaction{}, "ID")); diff != "" {
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
	testhelpers.SeedTestUser(t, userQ, userID)
	testhelpers.SeedTestCategories(t, db)
}
