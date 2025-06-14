package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateTransaction(t *testing.T) {
	userID := uuid.New()
	txID := uuid.New()

	tests := []struct {
		name             string
		req              models.NewTxRequest
		dateErr          error
		expDateErrSubStr string
		txErr            error
		expTxErrSubStr   string
	}{
		{
			name: "unsuccessful tx, invalid date",
			req: models.NewTxRequest{
				Date:             "03/27/94",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			dateErr:          errors.New("date error"),
			expDateErrSubStr: "invalid date format",
			txErr:            nil,
			expTxErrSubStr:   "",
		},
		{
			name: "unsuccessful tx, create tx failure",
			req: models.NewTxRequest{
				Date:             "2025-02-24",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			dateErr:          nil,
			expDateErrSubStr: "",
			txErr:            errors.New("tx error"),
			expTxErrSubStr:   "unable to create transaction",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			nopLogger := zap.NewNop()
			if tc.dateErr == nil {
				mockTxQ.On("CreateTransaction", ctx, mock.AnythingOfType("database.CreateTransactionParams")).Return(tc.txErr)
			}

			svc := transaction.NewTransactionService(mockTxQ, nopLogger)
			tx, err := svc.CreateTransaction(ctx, userID.String(), tc.req)

			if tc.expDateErrSubStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expDateErrSubStr)
				require.Nil(t, tx)
			} else if tc.expTxErrSubStr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expTxErrSubStr)
				require.Nil(t, tx)
				mockTxQ.AssertExpectations(t)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tx)

				expectedTx := models.NewTransaction(txID.String(), userID.String(), tc.req.Date, tc.req.Merchant, tc.req.Amount, tc.req.DetailedCategory)
				if diff := cmp.Diff(expectedTx, tx, cmpopts.IgnoreFields(models.Tx{}, "ID")); diff != "" {
					t.Errorf("transaction mismatch (-want +got)\n%s", diff)
				}
				mockTxQ.AssertExpectations(t)
			}

		})
	}
}
