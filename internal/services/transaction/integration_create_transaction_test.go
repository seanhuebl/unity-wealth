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
	"github.com/stretchr/testify/require"
)

func TestCreateTxIntegration(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name             string
		req              NewTransactionRequest
		dateErr          error
		expDateErrSubStr string
		txErr            error
		expTxErrSubStr   string
	}{
		{
			name: "successful create tx",
			req: NewTransactionRequest{
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

			createSchema(t, db)

			transactionalQ := database.NewRealTransactionalQuerier(database.New(db))
			txQ := database.NewRealTransactionQuerier(transactionalQ)
			userQ := database.NewRealUserQuerier(transactionalQ)

			seedData(t, db, userQ, userID)

			svc := NewTransactionService(txQ)

			tx, err := svc.CreateTransaction(ctx, userID.String(), tc.req)
			require.NoError(t, err)
			expectedTx := &Transaction{
				ID:               uuid.NewString(),
				UserID:           userID.String(),
				Date:             tc.req.Date,
				Merchant:         tc.req.Merchant,
				Amount:           tc.req.Amount,
				DetailedCategory: tc.req.DetailedCategory,
			}
			if diff := cmp.Diff(expectedTx, tx, cmpopts.IgnoreFields(Transaction{}, "ID")); diff != "" {
				t.Errorf("transaction mismatch (-want +got)\n%s", diff)
			}
			require.NotEmpty(t, tx.ID)
		})

	}
}

func createSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		hashed_password TEXT NOT NULL,
		risk_preference TEXT NOT NULL DEFAULT 'LOW',
		plan_type TEXT NOT NULL DEFAULT 'FREE',
		stripe_customer_id TEXT,
		stripe_subscription_id TEXT,
		scholarship_flag INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS primary_categories (
    	id INTEGER PRIMARY KEY,
    	name TEXT NOT NULL
		);
	`)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS detailed_categories (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		primary_category_id INTEGER NOT NULL,
		FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		transaction_date TEXT NOT NULL,
		merchant TEXT NOT NULL,
		amount_cents INTEGER NOT NULL CHECK(amount_cents <> 0),
		detailed_category_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (detailed_category_id) REFERENCES detailed_categories (id)
		);
	`)
	require.NoError(t, err)
}

func seedData(t *testing.T, db *sql.DB, userQ database.UserQuerier, userID uuid.UUID) {
	hashedPassword := "hashedpwd"
	email := "user@example.com"

	err := userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID.String(),
		Email:          email,
		HashedPassword: hashedPassword,
	})
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO primary_categories (id, name)
		VALUES (?1, ?2)
	`, 7, "Food")
	require.NoError(t, err)
	_, err = db.Exec(`
		INSERT INTO detailed_categories (id, name, description, primary_category_id)
		VALUES (?1, ?2, ?3, ?4)
	`, 40, "Groceries", "Purchases for fresh produce and groceries, including farmers' markets", 7)
	require.NoError(t, err)
}
