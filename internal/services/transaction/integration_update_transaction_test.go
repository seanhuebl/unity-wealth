package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/stretchr/testify/require"
)

func TestIntegrationUpdateTransaction(t *testing.T) {

	tests := []struct {
		name                  string
		req                   models.NewTransactionRequest
		dateErr               error
		expectedDateErrSubStr string
		txErr                 error
		expectedTxErrSubStr   string
	}{
		{
			name: "successful update",
			req: models.NewTransactionRequest{
				Date:             "2025-02-24",
				Merchant:         "costco",
				Amount:           157.98,
				DetailedCategory: 40,
			},
			dateErr:               nil,
			expectedDateErrSubStr: "",
			txErr:                 nil,
			expectedTxErrSubStr:   "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			txID := uuid.New()
			userID := uuid.New()
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_Keys = ON")
			require.NoError(t, err)

			helpers.CreateTestingSchema(t, db)

			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			txQ := database.NewRealTransactionQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)

			seedUpdateTxTestData(t, db, userQ, userID, txQ, txID)

			expectedTx := &models.Transaction{
				ID:               txID.String(),
				UserID:           userID.String(),
				Date:             tc.req.Date,
				Merchant:         tc.req.Merchant,
				Amount:           tc.req.Amount,
				DetailedCategory: tc.req.DetailedCategory,
			}

			svc := NewTransactionService(txQ)
			tx, err := svc.UpdateTransaction(ctx, txID.String(), userID.String(), tc.req)
			require.NoError(t, err)
			require.NotNil(t, tx)

			if diff := cmp.Diff(expectedTx, tx); diff != "" {
				t.Errorf("transaction mismatch (-want +got)\n%s", diff)
			}
		})
	}
}

func seedUpdateTxTestData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID, txQ database.TransactionQuerier, txID uuid.UUID) {
	helpers.SeedTestUser(t, userQ, userID)
	helpers.SeedTestCategories(t, db)
	helpers.SeedTestTransaction(t, txQ, userID, txID, &models.NewTransactionRequest{
		Date:             "2025-2-24",
		Merchant:         "sam's club",
		Amount:           200.25,
		DetailedCategory: 40,
	})
}
