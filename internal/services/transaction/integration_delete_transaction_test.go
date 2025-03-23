package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/stretchr/testify/require"
)

func TestIntegrationDeleteTransaction(t *testing.T) {

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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.deleteErr == nil {

				ctx := context.Background()
				txnID := uuid.New()
				userID := uuid.New()
				db, err := sql.Open("sqlite3", ":memory:")
				require.NoError(t, err)
				defer db.Close()
				_, err = db.Exec("PRAGMA foreign_keys = ON")
				require.NoError(t, err)

				helpers.CreateTestingSchema(t, db)

				transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
				txQ := database.NewRealTransactionQuerier(transactionalQ)
				userQ := database.NewRealUserQuerier(transactionalQ)
				seedDeleteTxTestData(t, db, userQ, userID, txQ, txnID)

				svc := NewTransactionService(txQ)

				err = svc.DeleteTransaction(ctx, txnID.String(), userID.String())
				require.NoError(t, err)

				_, err = txQ.GetUserTransactionByID(ctx, database.GetUserTransactionByIDParams{
					UserID: userID.String(),
					ID:     txnID.String(),
				})
				require.ErrorIs(t, err, sql.ErrNoRows)
			}

		})
	}
}

func seedDeleteTxTestData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID, txQ database.TransactionQuerier, txID uuid.UUID) {
	helpers.SeedTestUser(t, userQ, userID)
	helpers.SeedTestCategories(t, db)
	helpers.SeedTestTransaction(t, txQ, userID, txID, &models.NewTransactionRequest{
		Date:             "2025-02-24",
		Merchant:         "Costco",
		Amount:           145.56,
		DetailedCategory: 40,
	})
}
