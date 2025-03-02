package constants

type contextKey string

const (
	ClaimsKey  = contextKey("claims")
	UserIDKey  = contextKey("userID")
	RequestKey = contextKey("httpRequest")
)

const CreateTxTable = `
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
	`
const CreateUsersTable = `
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
	`