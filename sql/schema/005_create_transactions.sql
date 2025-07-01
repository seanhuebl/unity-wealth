-- +goose Up
/* -- PRIMARY_CATEGORIES -- (lookup table) */
CREATE TABLE IF NOT EXISTS primary_categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

/* -- DETAILED_CATEGORIES -- (child table) */
CREATE TABLE IF NOT EXISTS detailed_categories (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    primary_category_id INTEGER NOT NULL,
    CONSTRAINT fk_detailed_primary FOREIGN KEY (primary_category_id) REFERENCES primary_categories (id) ON DELETE CASCADE
);

/* -- helpful index: fast joins “find all detailed cats in primary X” -- */
CREATE INDEX IF NOT EXISTS idx_detailed_categories_primary ON detailed_categories (primary_category_id);

/* -- TRANSACTIONS  (fact table) -- */
CREATE TABLE IF NOT EXISTS transactions (
    -- keys first
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    detailed_category_id INTEGER NOT NULL,
    user_id UUID NOT NULL,
    -- business data
    transaction_date DATE NOT NULL,
    merchant CITEXT NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents <> 0),
    -- lifecycle
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- constraints
    CONSTRAINT fk_transactions_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_transactions_detailed FOREIGN KEY (detailed_category_id) REFERENCES detailed_categories (id)
);

/* -- trigger to auto-bump updated_at -- */
CREATE TRIGGER transactions_touch BEFORE
UPDATE ON transactions FOR EACH ROW
EXECUTE FUNCTION trg_set_timestamp ();

/* -- helpful indexes -- */
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);

CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions (transaction_date DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_detailed_cat ON transactions (detailed_category_id);

-- +goose Down
DROP TRIGGER IF EXISTS transactions_touch ON transactions;

DROP INDEX IF EXISTS idx_transactions_detailed_cat;

DROP INDEX IF EXISTS idx_transactions_date;

DROP INDEX IF EXISTS idx_transactions_user_id;

DROP TABLE IF EXISTS transactions;

DROP INDEX IF EXISTS uq_detailed_name_per_primary;

DROP INDEX IF EXISTS idx_detailed_categories_primary;

DROP TABLE IF EXISTS detailed_categories;

DROP TABLE IF EXISTS primary_categories;
