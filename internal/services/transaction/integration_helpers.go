package transaction

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/helpers"
	"github.com/stretchr/testify/require"
)

func CreateTestingSchema(t *testing.T, db *sql.DB) {
	_, err := db.Exec(constants.CreateUsersTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreatePrimCatTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateDetCatTable)
	require.NoError(t, err)
	_, err = db.Exec(constants.CreateTxTable)
	require.NoError(t, err)
}

func SeedTestUser(t *testing.T, userQ database.UserQuerier, userID uuid.UUID) {
	hashedPassword := "hashedpwd"
	email := "user@example.com"

	err := userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID.String(),
		Email:          email,
		HashedPassword: hashedPassword,
	})
	require.NoError(t, err)
}

func SeedTestCategories(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`
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

func SeedTestTransaction(t *testing.T, txQ database.TransactionQuerier, userID, txID uuid.UUID, req *NewTransactionRequest) {
	ctx := context.Background()
	err := txQ.CreateTransaction(ctx, database.CreateTransactionParams{
		ID:                 txID.String(),
		UserID:             userID.String(),
		TransactionDate:    req.Date,
		Merchant:           req.Merchant,
		AmountCents:        helpers.ConvertToCents(req.Amount),
		DetailedCategoryID: req.DetailedCategory,
	})
	require.NoError(t, err)
}

func SeedMultipleTestTransactions[T TxPageRow](t *testing.T, txQ database.TransactionQuerier, rows []T) {
	ctx := context.Background()
	for _, row := range rows {
		err := txQ.CreateTransaction(ctx, database.CreateTransactionParams{
			ID:                 row.GetTxID().String(),
			UserID:             row.GetUserID().String(),
			TransactionDate:    row.GetTxDate(),
			Merchant:           row.GetMerchant(),
			AmountCents:        row.GetAmountCents(),
			DetailedCategoryID: row.GetDetailedCatID(),
		})
		require.NoError(t, err)
	}
}
