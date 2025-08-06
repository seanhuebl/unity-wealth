package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListUserTransactions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name                string
		userID              uuid.UUID
		cursorToken         string
		expectedCursorDate  string
		expectedHasMoreData bool
		pageSize            int32
		txSliceLength       int
		getFirstPageErr     error
		getTxPaginatedErr   error
		wantErr             error
	}{
		{
			name:                "first page, more data, success",
			userID:              uuid.New(),
			cursorToken:         "",
			expectedCursorDate:  "2025-02-05",
			expectedHasMoreData: true,
			pageSize:            5,
			txSliceLength:       10,
			getFirstPageErr:     nil,
		},
		{
			name:                "first page, no extra data, success",
			userID:              uuid.New(),
			cursorToken:         "",
			expectedCursorDate:  "",
			expectedHasMoreData: false,
			pageSize:            5,
			txSliceLength:       2,
			getFirstPageErr:     nil,
		},
		{
			name:                "paginated, more data, success",
			userID:              uuid.New(),
			cursorToken:         "token",
			expectedCursorDate:  "2025-02-10",
			expectedHasMoreData: true,
			pageSize:            10,
			txSliceLength:       15,
			getTxPaginatedErr:   nil,
		},
		{
			name:                "paginated, no extra data, success",
			userID:              uuid.New(),
			cursorToken:         "token",
			expectedCursorDate:  "",
			expectedHasMoreData: false,
			pageSize:            10,
			txSliceLength:       7,
			getTxPaginatedErr:   nil,
		},
		{
			name:               "first page, db error",
			userID:             uuid.New(),
			cursorToken:        "",
			expectedCursorDate: "",
			pageSize:           1,
			getFirstPageErr:    errors.New("db error"),
			wantErr:            sentinels.ErrDBExecFailed,
		},
		{
			name:               "first page, no transactions found",
			userID:             uuid.New(),
			cursorToken:        "",
			expectedCursorDate: "",
			pageSize:           1,
			getFirstPageErr:    nil,
			wantErr:            transaction.ErrTxNotFound,
		},
		{
			name:               "paginated, db error",
			userID:             uuid.New(),
			cursorToken:        "token",
			expectedCursorDate: "",
			pageSize:           1,
			getTxPaginatedErr:  errors.New("db error"),
			wantErr:            sentinels.ErrDBExecFailed,
		},
		{
			name:               "paginated, no transactions found",
			userID:             uuid.New(),
			cursorToken:        "token",
			expectedCursorDate: "",
			pageSize:           1,
			getTxPaginatedErr:  nil,
			wantErr:            transaction.ErrTxNotFound,
		},
		{
			name:     "page size <= 0",
			userID:   uuid.New(),
			pageSize: 0,
			wantErr:  transaction.ErrInvalidPageSizeNonPositive,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			fetchSize := tc.pageSize + 1
			t.Cleanup(func() { mockTxQ.AssertExpectations(t) })

			fetchType := constants.FTFirst
			var nextRows []database.GetUserTransactionsPaginatedRow
			var firstPageRows []database.GetUserTransactionsFirstPageRow

			switch {
			case tc.cursorToken == "":
				firstPageRows = generateFirstPageRows(tc.userID, tc.txSliceLength)
				if len(firstPageRows) > int(fetchSize) {
					firstPageRows = firstPageRows[:fetchSize]
				}
				switch {
				case tc.wantErr == nil:
					mockTxQ.On("GetUserTransactionsFirstPage", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).
						Return(firstPageRows, nil).Once()
				case errors.Is(tc.wantErr, sentinels.ErrDBExecFailed):
					mockTxQ.On("GetUserTransactionsFirstPage", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).
						Return(nil, errors.New("db error")).Once()
				case errors.Is(tc.wantErr, transaction.ErrTxNotFound):
					mockTxQ.On("GetUserTransactionsFirstPage", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsFirstPageParams")).
						Return(nil, sql.ErrNoRows).Once()
				}
			default:
				fetchType = constants.FTPag
				nextRows = generatePaginatedRows(tc.userID, tc.txSliceLength)
				if len(nextRows) > int(fetchSize) {
					nextRows = nextRows[:fetchSize]
				}
				switch {
				case tc.wantErr == nil:
					mockTxQ.On("GetUserTransactionsPaginated", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).
						Return(nextRows, nil).Once()
				case errors.Is(tc.wantErr, sentinels.ErrDBExecFailed):
					mockTxQ.On("GetUserTransactionsPaginated", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).
						Return(nil, errors.New("db error")).Once()
				case errors.Is(tc.wantErr, transaction.ErrTxNotFound):
					mockTxQ.On("GetUserTransactionsPaginated", mock.Anything, mock.AnythingOfType("database.GetUserTransactionsPaginatedParams")).
						Return(nil, sql.ErrNoRows).Once()
				}
			}

			svc := transaction.NewTransactionService(mockTxQ, &cursor.NoopSigner{}, zap.NewNop())
			txResult, err := svc.ListUserTransactions(ctx, tc.userID, tc.cursorToken, tc.pageSize)

			if tc.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err, tc.wantErr)
				return

			}
			txs := make([]models.Tx, 0, tc.pageSize)
			if fetchType == constants.FTFirst {
				if len(firstPageRows) > int(tc.pageSize) {
					firstPageRows = firstPageRows[:tc.pageSize]
				}
				txs = helpers.AppendTxs(txs, helpers.SliceToTxRows(firstPageRows))
			} else {
				if len(nextRows) > int(tc.pageSize) {
					nextRows = nextRows[:tc.pageSize]
				}
				txs = helpers.AppendTxs(txs, helpers.SliceToTxRows(nextRows))
			}
			if tc.expectedHasMoreData {
				require.NotEmpty(t, txResult.NextCursor)
			}
			require.Equal(t, tc.expectedHasMoreData, txResult.HasMoreData)

			if diff := cmp.Diff(txs, txResult.Transactions, cmpopts.IgnoreFields(models.Tx{}, "ID")); diff != "" {
				t.Errorf("transaction slice mismatch (-want +got)\n%s", diff)
			}
			mockTxQ.AssertExpectations(t)
		})
	}
}

// Helpers
func generateFirstPageRows(userID uuid.UUID, txSliceLength int) []database.GetUserTransactionsFirstPageRow {
	var rows []database.GetUserTransactionsFirstPageRow
	for i := 0; i < txSliceLength; i++ {
		d := fmt.Sprintf("2025-02-%02d", i+1)
		date, err := time.Parse(constants.LayoutDate, d)
		if err != nil {
			log.Fatal("issue generating first page rows")
		}
		rows = append(rows, database.GetUserTransactionsFirstPageRow{
			ID:                 uuid.New(),
			UserID:             userID,
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
		d := fmt.Sprintf("2025-02-%02d", i+1)
		date, err := time.Parse(constants.LayoutDate, d)
		if err != nil {
			log.Fatal("issue generating first page rows")
		}
		rows = append(rows, database.GetUserTransactionsPaginatedRow{
			ID:                 uuid.New(),
			UserID:             userID,
			TransactionDate:    date,
			Merchant:           "costco",
			AmountCents:        15744,
			DetailedCategoryID: 40,
		})
	}
	return rows
}
