package constants

type contextKey string

const (
	ClaimsKey     = contextKey("claims")
	UserIDKey     = contextKey("userID")
	RequestKey    = contextKey("httpRequest")
	CreateTxTable = `
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
	CreateUsersTable = `
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

	CreatePrimCatTable = `
			CREATE TABLE IF NOT EXISTS primary_categories (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
			);
		`

	CreateDetCatTable = `
			CREATE TABLE IF NOT EXISTS detailed_categories (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			primary_category_id INTEGER NOT NULL,
			FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
			);
		`
	CreateDeviceInfoTable = `
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
	`
	CreateRefrTokenTable = `
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
	` // #nosec
)
