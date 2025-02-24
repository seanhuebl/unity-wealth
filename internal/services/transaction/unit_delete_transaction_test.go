package transaction

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
			expectedDeleteErrSubStr: "error deleting transaction",
		},
		{
			name:                    "no err but tx not found",
			deleteErr:               sql.ErrNoRows,
			expectedDeleteErrSubStr: "no transaction found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTxQ := dbmocks.NewTransactionQuerier(t)

			mockTxQ.On("DeleteTransactionByID", ctx, mock.AnythingOfType("database.DeleteTransactionByIDParams")).Return(tc.deleteErr)

			svc := NewTransactionService(mockTxQ)

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
