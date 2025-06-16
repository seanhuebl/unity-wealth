package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeleteTransaction(t *testing.T) {
	txnID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name                    string
		deleteErr               error
		expectedDeleteErrSubStr string
	}{
		{
			name:                    "delete successful",
			deleteErr:               nil,
			expectedDeleteErrSubStr: "",
		},
		{
			name:                    "delete transaction failure",
			deleteErr:               errors.New("delete error"),
			expectedDeleteErrSubStr: sentinels.ErrDBExecFailed.Error(),
		},
		{
			name:                    "no err but tx not found",
			deleteErr:               sql.ErrNoRows,
			expectedDeleteErrSubStr: transaction.ErrTxNotFound.Error(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			nopLogger := zap.NewNop()
			mockTxQ.On("DeleteTransactionByID", ctx, mock.AnythingOfType("database.DeleteTransactionByIDParams")).Return(txnID.String(), tc.deleteErr)

			svc := transaction.NewTransactionService(mockTxQ, nopLogger)

			err := svc.DeleteTransaction(ctx, txnID.String(), userID.String())

			if tc.deleteErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedDeleteErrSubStr)
				mockTxQ.AssertExpectations(t)
			} else {
				require.NoError(t, err)
				mockTxQ.AssertExpectations(t)
			}

		})
	}
}
