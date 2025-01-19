-- +goose Up
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY,
    token_hash TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    revoked_at DATETIME,
    user_id UUID NOT NULL,
    device_info_id UUID NOT NULL,
    UNIQUE (user_id, device_info_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (device_info_id) REFERENCES device_info_logs (id) ON DELETE CASCADE
);
-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;