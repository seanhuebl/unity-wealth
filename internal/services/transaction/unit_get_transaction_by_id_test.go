package transaction

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionByID(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name                string
		userID              uuid.UUID
		txnID               uuid.UUID
		txErr               error
		expectedTxErrSubstr string
	}{
		{
			name:                "successful retrieval",
			userID:              uuid.New(),
			txnID:               uuid.New(),
			txErr:               nil,
			expectedTxErrSubstr: "",
		},
		{
			name:                "database error",
			userID:              uuid.New(),
			txnID:               uuid.New(),
			txErr:               errors.New("db error"),
			expectedTxErrSubstr: "error getting transaction",
		},
		{
			name:                "userID / txnID pair not found",
			userID:              uuid.New(),
			txnID:               uuid.New(),
			txErr:               sql.ErrNoRows,
			expectedTxErrSubstr: "transaction not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			expectedRow := database.GetUserTransactionByIDRow{
				ID:                 tc.txnID.String(),
				UserID:             tc.userID.String(),
				TransactionDate:    "2025-02-25",
				Merchant:           "costco",
				AmountCents:        19725,
				DetailedCategoryID: 40,
			}
			expectedTxn := &models.Transaction{
				ID:               expectedRow.ID,
				UserID:           expectedRow.UserID,
				Date:             expectedRow.TransactionDate,
				Merchant:         expectedRow.Merchant,
				Amount:           helpers.CentsToDollars(expectedRow.AmountCents),
				DetailedCategory: expectedRow.DetailedCategoryID,
			}

			mockTxQ := dbmocks.NewTransactionQuerier(t)
			if tc.txErr != nil {
				expectedRow = database.GetUserTransactionByIDRow{}
			}
			mockTxQ.On("GetUserTransactionByID", ctx, mock.AnythingOfType("database.GetUserTransactionByIDParams")).Return(expectedRow, tc.txErr)

			svc := NewTransactionService(mockTxQ)

			txn, err := svc.GetTransactionByID(ctx, tc.userID.String(), tc.txnID.String())
			if tc.expectedTxErrSubstr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedTxErrSubstr)
				mockTxQ.AssertExpectations(t)
			} else {
				require.NoError(t, err)

				if diff := cmp.Diff(txn, expectedTxn); diff != "" {
					t.Errorf("transaction mismatch (-want +got)\n%s", diff)
				}
				mockTxQ.AssertExpectations(t)
			}

		})
	}
}
