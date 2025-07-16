package transaction_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeleteTransaction(t *testing.T) {
	t.Parallel()
	txnID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name    string
		repoRes driver.Result
		repoErr error
		wantErr error
	}{
		{
			name:    "delete successful",
			repoRes: sqlmock.NewResult(0, 1),
			repoErr: nil,
			wantErr: nil,
		},
		{
			name:    "delete transaction failure",
			repoRes: nil,
			repoErr: errors.New("driver error"),
			wantErr: sentinels.ErrDBExecFailed,
		},
		{
			name:    "no err but tx not found",
			repoRes: sqlmock.NewResult(0, 0),
			repoErr: nil,
			wantErr: transaction.ErrTxNotFound,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			t.Cleanup(func() { mockTxQ.AssertExpectations(t) })

			mockTxQ.On("DeleteTransactionByID", mock.Anything, mock.AnythingOfType("database.DeleteTransactionByIDParams")).
				Return(tc.repoRes, tc.repoErr).Once()

			svc := transaction.NewTransactionService(mockTxQ, &cursor.NoopSigner{}, zap.NewNop())

			err := svc.DeleteTransaction(ctx, txnID, userID)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
