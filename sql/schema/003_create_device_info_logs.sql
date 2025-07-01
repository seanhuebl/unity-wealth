-- +goose Up
CREATE TABLE IF NOT EXISTS device_info_logs (
    -- keys
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    user_id UUID NOT NULL,
    -- business data
    device_type TEXT NOT NULL,
    browser CITEXT NOT NULL,
    browser_version TEXT NOT NULL,
    os TEXT NOT NULL,
    os_version TEXT NOT NULL,
    app_info JSONB,
    -- lifecycle
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- constraints
    CONSTRAINT fk_device_info_logs_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

/* -- last_used_at auto-maintenance trigger -- */
CREATE TRIGGER device_info_logs_touch BEFORE
UPDATE ON device_info_logs FOR EACH ROW
EXECUTE FUNCTION trg_touch_last_used_at ();

/* -- indexing -- */
CREATE INDEX IF NOT EXISTS idx_device_info_logs_user_id ON device_info_logs (user_id);

CREATE INDEX IF NOT EXISTS idx_device_info_logs_last_used_at ON device_info_logs (last_used_at DESC);

-- +goose Down
DROP TRIGGER IF EXISTS device_info_logs_touch ON device_info_logs;

DROP INDEX IF EXISTS idx_device_info_logs_last_used_at;

DROP INDEX IF EXISTS idx_device_info_logs_user_id;

DROP TABLE IF EXISTS device_info_logs;
