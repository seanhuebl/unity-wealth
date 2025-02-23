package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/stretchr/testify/require"
)

func TestLoginIntegration(t *testing.T) {

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("PRAGMA foeriegn_keys = ON")
	require.NoError(t, err)

	createSchema(t, db)
	transactionalQ := database.NewRealTransactionalQuerier(database.New(db))

	txQ := database.NewRealTransactionQuerier(transactionalQ)
	userQ := database.NewRealUserQuerier(transactionalQ)
	tokeGen := NewRealTokenGenerator("tokensecret", TokenType("unity-wealth"))
	tokenExt := NewRealTokenExtractor()
	pwdHasher := NewRealPwdHasher()
	userID, userEmail, userHashedPW := seedTestUser(t, db, pwdHasher, userQ)

}

// Helpers
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
		CREATE TABLE IF NOT EXISTS refresh_tokens (
		id TEXT PRIMARY KEY,
		token_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME,
		revoked_at DATETIME,
		user_id TEXT NOT NULL,
		device_info_id TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
		FOREIGN KEY (device_info_id) REFERENCES device_info_logs (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS device_info_logs (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		device_type TEXT NOT NULL,
		browser TEXT NOT NULL,
		browser_version TEXT NOT NULL,
		os TEXT NOT NULL,
		os_version TEXT NOT NULL,
		app_info TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_used_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		);
	`)
	require.NoError(t, err)
}

func seedTestUser(t *testing.T, db *sql.DB, hasher PasswordHasher, userQ database.UserQuerier) (uuid.UUID, string, string) {
	password := "Validpass1!"
	email := "user@example.com"
	userID := uuid.New()
	hashedPwd, err := hasher.HashPassword(password)
	require.NoError(t, err)

	err = userQ.CreateUser(context.Background(), database.CreateUserParams{
		ID:             userID.String(),
		Email:          email,
		HashedPassword: hashedPwd,
	})
	require.NoError(t, err)
	return userID, email, hashedPwd
}
