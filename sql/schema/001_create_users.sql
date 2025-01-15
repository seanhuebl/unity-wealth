-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
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
-- +goose Down
DROP TABLE IF EXISTS users CASCADE;