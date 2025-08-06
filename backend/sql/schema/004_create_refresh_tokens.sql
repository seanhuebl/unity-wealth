-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
    -- keys
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    device_info_id UUID NOT NULL,
    user_id UUID NOT NULL,
    -- business data
    token_hash TEXT NOT NULL,
    -- lifecycle
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- audit
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    -- constraints
    CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_refresh_tokens_device FOREIGN KEY (device_info_id) REFERENCES device_info_logs (id) ON DELETE CASCADE
);

/* -- updated_at auto-maintenance trigger -- */
CREATE TRIGGER refresh_tokens_touch BEFORE
UPDATE ON refresh_tokens FOR EACH ROW
EXECUTE FUNCTION trg_set_timestamp ();

/* -- indexes -- */
-- fast look-ups by hash
CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

-- queries like “all active tokens for user X”
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- quickly filter active tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active ON refresh_tokens (user_id)
WHERE
    revoked_at IS NULL;

-- +goose Down
DROP TRIGGER IF EXISTS refresh_tokens_touch ON refresh_tokens;

DROP INDEX IF EXISTS idx_refresh_tokens_active;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;

DROP TABLE IF EXISTS refresh_tokens;
