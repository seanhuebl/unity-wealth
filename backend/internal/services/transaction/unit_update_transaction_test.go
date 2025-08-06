package transaction_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/cursor"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
	"github.com/seanhuebl/unity-wealth/internal/services/transaction"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUpdateTransaction(t *testing.T) {
	t.Parallel()
	txID := uuid.New()
	userID := uuid.New()
	ctx := context.Background()
	req := models.NewTxRequest{
		Date:             "2025-02-24",
		Merchant:         "costco",
		Amount:           157.98,
		DetailedCategory: 40,
	}

	tests := []struct {
		name string
		req  models.NewTxRequest
		want error
	}{
		{
			name: "success",
			req:  req,
			want: nil,
		},
		{
			name: "improper date format",
			req: models.NewTxRequest{
				Date:             "2/24/25",
				Merchant:         "costco",
				Amount:           157.98,
				DetailedCategory: 40,
			},
			want: transaction.ErrInvalidDateFormat,
		},
		{
			name: "update tx failure",
			req:  req,
			want: sentinels.ErrDBExecFailed,
		},
		{
			name: "transaction not found",
			req:  req,
			want: transaction.ErrTxNotFound,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockTxQ := dbmocks.NewTransactionQuerier(t)
			t.Cleanup(func() { mockTxQ.AssertExpectations(t) })

			date, _ := time.Parse(constants.LayoutDate, tc.req.Date)

			expectedRow := database.UpdateTransactionByIDRow{
				ID:                 txID,
				TransactionDate:    date,
				Merchant:           tc.req.Merchant,
				AmountCents:        int64(tc.req.Amount * 100),
				DetailedCategoryID: 40,
				UpdatedAt:          time.Now(),
			}

			switch {
			case tc.want == nil:

				mockTxQ.On("UpdateTransactionByID", mock.Anything, mock.AnythingOfType("database.UpdateTransactionByIDParams")).
					Return(expectedRow, tc.want).Once()

			case errors.Is(tc.want, sentinels.ErrDBExecFailed):

				mockTxQ.On("UpdateTransactionByID", mock.Anything, mock.AnythingOfType("database.UpdateTransactionByIDParams")).
					Return(database.UpdateTransactionByIDRow{}, errors.New("driver error")).Once()

			case errors.Is(tc.want, transaction.ErrTxNotFound):

				mockTxQ.On("UpdateTransactionByID", mock.Anything, mock.AnythingOfType("database.UpdateTransactionByIDParams")).
					Return(database.UpdateTransactionByIDRow{}, sql.ErrNoRows).Once()
			}

			svc := transaction.NewTransactionService(mockTxQ, &cursor.RealSigner{}, zap.NewNop())
			tx, err := svc.UpdateTransaction(ctx, txID, userID, tc.req)

			if tc.want != nil {
				require.ErrorIs(t, err, tc.want)
				require.Nil(t, tx)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tx)
				mockTxQ.AssertExpectations(t)
				expectedTx := &models.Tx{
					ID:               expectedRow.ID,
					UserID:           userID,
					Date:             expectedRow.TransactionDate,
					Merchant:         expectedRow.Merchant,
					Amount:           float64(expectedRow.AmountCents) / 100.0,
					DetailedCategory: expectedRow.DetailedCategoryID,
					UpdatedAt:        expectedRow.UpdatedAt,
				}
				if diff := cmp.Diff(expectedTx, tx); diff != "" {
					t.Errorf("transaction mismatch (-want +got)\n%s", diff)
				}
			}
		})
	}
}
