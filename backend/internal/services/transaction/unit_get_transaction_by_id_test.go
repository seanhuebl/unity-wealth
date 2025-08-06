package transaction_test

import (
	"context"
	"database/sql"
	"errors"
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

func TestGetTransactionByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tests := []struct {
		name   string
		userID uuid.UUID
		txID   uuid.UUID
		want   error
	}{
		{
			name:   "success",
			userID: uuid.New(),
			txID:   uuid.New(),
			want:   nil,
		},
		{
			name:   "database error",
			userID: uuid.New(),
			txID:   uuid.New(),
			want:   sentinels.ErrDBExecFailed,
		},
		{
			name:   "userID / txnID pair not found",
			userID: uuid.New(),
			txID:   uuid.New(),
			want:   transaction.ErrTxNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			date, err := time.Parse(constants.LayoutDate, "2025-02-25")
			require.NoError(t, err)
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			t.Cleanup(func() { mockTxQ.AssertExpectations(t) })
			expectedRow := database.GetUserTransactionByIDRow{
				ID:                 tc.txID,
				UserID:             tc.userID,
				TransactionDate:    date,
				Merchant:           "costco",
				AmountCents:        19725,
				DetailedCategoryID: 40,
			}
			expectedTx := &models.Tx{
				ID:               expectedRow.ID,
				UserID:           expectedRow.UserID,
				Date:             expectedRow.TransactionDate,
				Merchant:         expectedRow.Merchant,
				Amount:           helpers.CentsToDollars(expectedRow.AmountCents),
				DetailedCategory: expectedRow.DetailedCategoryID,
			}
			switch {
			case tc.want == nil:
				mockTxQ.On("GetUserTransactionByID", mock.Anything, mock.AnythingOfType("database.GetUserTransactionByIDParams")).
					Return(expectedRow, nil)
			case errors.Is(tc.want, sentinels.ErrDBExecFailed):
				mockTxQ.On("GetUserTransactionByID", mock.Anything, mock.AnythingOfType("database.GetUserTransactionByIDParams")).
					Return(database.GetUserTransactionByIDRow{}, errors.New("db error"))
			case errors.Is(tc.want, transaction.ErrTxNotFound):
				mockTxQ.On("GetUserTransactionByID", mock.Anything, mock.AnythingOfType("database.GetUserTransactionByIDParams")).
					Return(database.GetUserTransactionByIDRow{}, sql.ErrNoRows)
			}

			svc := transaction.NewTransactionService(mockTxQ, &cursor.NoopSigner{}, zap.NewNop())

			tx, err := svc.GetTransactionByID(ctx, tc.userID, tc.txID)
			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				require.Nil(t, tx)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(expectedTx, tx, cmpopts.IgnoreFields(*new(models.Tx), "UpdatedAt")); diff != "" {
					t.Errorf("transaction mismatch (-want +got)\n%s", diff)
				}
			}
		})
	}
}
