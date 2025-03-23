package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/stretchr/testify/require"
)

func TestCreateTxIntegration(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name             string
		req              models.NewTransactionRequest
		dateErr          error
		expDateErrSubStr string
		txErr            error
		expTxErrSubStr   string
	}{
		{
			name: "successful create tx",
			req: models.NewTransactionRequest{
				Date:             "2025-02-24",
				Merchant:         "Costco",
				Amount:           145.56,
				DetailedCategory: 40,
			},
			dateErr:          nil,
			expDateErrSubStr: "",
			txErr:            nil,
			expTxErrSubStr:   "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			db, err := sql.Open("sqlite3", ":memory:")
			require.NoError(t, err)
			defer db.Close()
			_, err = db.Exec("PRAGMA foreign_keys = ON")
			require.NoError(t, err)

			helpers.CreateTestingSchema(t, db)

			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			txQ := database.NewRealTransactionQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)

			seedCreateTxTestData(t, db, userQ, userID)

			svc := NewTransactionService(txQ)

			tx, err := svc.CreateTransaction(ctx, userID.String(), tc.req)
			require.NoError(t, err)
			expectedTx := &models.Transaction{
				ID:               uuid.NewString(),
				UserID:           userID.String(),
				Date:             tc.req.Date,
				Merchant:         tc.req.Merchant,
				Amount:           tc.req.Amount,
				DetailedCategory: tc.req.DetailedCategory,
			}
			if diff := cmp.Diff(expectedTx, tx, cmpopts.IgnoreFields(models.Transaction{}, "ID")); diff != "" {
				t.Errorf("transaction mismatch (-want +got)\n%s", diff)
			}
			require.NotEmpty(t, tx.ID)
		})

	}
}

func seedCreateTxTestData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID) {
	helpers.SeedTestUser(t, userQ, userID)
	helpers.SeedTestCategories(t, db)
}
