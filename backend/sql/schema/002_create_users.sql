-- +goose Up
/* -- enums -- */
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS(
        SELECT
            1
        FROM
            pg_type
        WHERE
            typname = 'risk_preference_enum') THEN
    CREATE TYPE risk_preference_enum AS ENUM(
        'LOW',
        'MEDIUM',
        'HIGH'
);
END IF;
    IF NOT EXISTS(
        SELECT
            1
        FROM
            pg_type
        WHERE
            typname = 'plan_type_enum') THEN
    CREATE TYPE plan_type_enum AS ENUM(
        'FREE',
        'PRO'
);
END IF;
END
$$;

-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS users (
    -- keys
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    -- business data
    email CITEXT NOT NULL UNIQUE,
    hashed_password text NOT NULL,
    risk_preference risk_preference_enum NOT NULL DEFAULT 'LOW',
    plan_type plan_type_enum NOT NULL DEFAULT 'FREE',
    stripe_customer_id text,
    stripe_subscription_id text,
    scholarship_flag boolean NOT NULL DEFAULT FALSE,
    -- lifecycle
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
);

/* -- indexing -- */
CREATE INDEX idx_users_email ON users (email);

/* -- updated_at auto-maintenance trigger -- */
CREATE TRIGGER users_set_timestamp BEFORE
UPDATE ON users FOR EACH ROW
EXECUTE FUNCTION trg_set_timestamp ();

-- +goose Down
DROP TRIGGER IF EXISTS users_set_timestamp ON users;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS plan_type_enum;

DROP TYPE IF EXISTS risk_preference_enum;
