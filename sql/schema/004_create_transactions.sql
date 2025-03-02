-- +goose Up
CREATE TABLE IF NOT EXISTS primary_categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS detailed_categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    primary_category_id INTEGER NOT NULL,
    FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
);
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
-- +goose Down
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS detailed_categories;
DROP TABLE IF EXISTS primary_categories;