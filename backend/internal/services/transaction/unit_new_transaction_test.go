package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateTransaction(t *testing.T) {
	t.Parallel()
	userID := uuid.New()

	tests := []struct {
		name string
		req  models.NewTxRequest
		want error
	}{
		{
			name: "unsuccessful tx, invalid date",
			req: models.NewTxRequest{
				Date:             "03/27/94",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			want: transaction.ErrInvalidDateFormat,
		},
		{
			name: "unsuccessful tx, create tx failure",
			req: models.NewTxRequest{
				Date:             "2025-02-24",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			want: sentinels.ErrDBExecFailed,
		},
		{
			name: "success",
			req: models.NewTxRequest{
				Date:             "2025-02-24",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			want: nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			ctx = context.WithValue(ctx, constants.RequestIDKey, uuid.New())

			mockTxQ := dbmocks.NewTransactionQuerier(t)
			t.Cleanup(func() { mockTxQ.AssertExpectations(t) })

			switch {
			case tc.want == nil:

				mockTxQ.On("CreateTransaction", mock.Anything, mock.AnythingOfType("database.CreateTransactionParams")).
					Return(nil).Once()

			case errors.Is(tc.want, sentinels.ErrDBExecFailed):

				mockTxQ.On("CreateTransaction", mock.Anything, mock.AnythingOfType("database.CreateTransactionParams")).
					Return(errors.New("driver error")).Once()

			}

			svc := transaction.NewTransactionService(mockTxQ, &cursor.NoopSigner{}, zap.NewNop())
			tx, err := svc.CreateTransaction(ctx, userID, tc.req)

			if tc.want == nil {
				require.NoError(t, err)
				require.NotNil(t, tx)
				return
			}
			require.Error(t, err)
			require.ErrorIs(t, err, tc.want)
			require.Nil(t, tx)

		})
	}
}
