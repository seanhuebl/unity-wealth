package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateTransaction(t *testing.T) {
	txID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name                  string
		req                   models.NewTxRequest
		dateErr               error
		expectedDateErrSubStr string
		txErr                 error
		expectedTxErrSubStr   string
	}{
		{
			name: "improper date format",
			req: models.NewTxRequest{
				Date:             "2/24/25",
				Merchant:         "costco",
				Amount:           157.98,
				DetailedCategory: 40,
			},
			dateErr:               errors.New("date error"),
			expectedDateErrSubStr: "invalid date format",
			txErr:                 nil,
			expectedTxErrSubStr:   "",
		},
		{
			name: "update tx failure",
			req: models.NewTxRequest{
				Date:             "2025-02-24",
				Merchant:         "costco",
				Amount:           157.98,
				DetailedCategory: 40,
			},
			dateErr:               nil,
			expectedDateErrSubStr: "",
			txErr:                 errors.New("tx error"),
			expectedTxErrSubStr:   "error updating transaction",
		},
		{
			name: "transaction not found",
			req: models.NewTxRequest{
				Date:             "2025-02-24",
				Merchant:         "costco",
				Amount:           157.98,
				DetailedCategory: 40,
			},
			dateErr:               nil,
			expectedDateErrSubStr: "",
			txErr:                 sql.ErrNoRows,
			expectedTxErrSubStr:   "transaction not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			expectedRow := database.UpdateTransactionByIDRow{
				ID:                 txID.String(),
				TransactionDate:    tc.req.Date,
				Merchant:           tc.req.Merchant,
				AmountCents:        int64(tc.req.Amount * 100),
				DetailedCategoryID: 40,
			}
			if tc.dateErr == nil {
				returnRow := expectedRow
				if tc.txErr != nil {
					returnRow = database.UpdateTransactionByIDRow{}
				}
				mockTxQ.On("UpdateTransactionByID", ctx, mock.AnythingOfType("database.UpdateTransactionByIDParams")).Return(returnRow, tc.txErr)
			}
			svc := transaction.NewTransactionService(mockTxQ)
			tx, err := svc.UpdateTransaction(ctx, txID.String(), userID.String(), tc.req)
			if tc.expectedDateErrSubStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedDateErrSubStr)
				require.Nil(t, tx)
			} else if tc.expectedTxErrSubStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedTxErrSubStr)
				require.Nil(t, tx)
				mockTxQ.AssertExpectations(t)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tx)
				mockTxQ.AssertExpectations(t)
				expectedTx := &models.Tx{
					ID:               expectedRow.ID,
					UserID:           userID.String(),
					Date:             expectedRow.TransactionDate,
					Merchant:         expectedRow.Merchant,
					Amount:           float64(expectedRow.AmountCents) / 100.0,
					DetailedCategory: expectedRow.DetailedCategoryID,
				}
				if diff := cmp.Diff(expectedTx, tx); diff != "" {
					t.Errorf("transaction mismatch (-want +got)\n%s", diff)
				}
			}
		})
	}
}
