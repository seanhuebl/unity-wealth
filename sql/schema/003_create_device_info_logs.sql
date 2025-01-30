-- +goose Up
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
-- +goose Down
DROP TABLE IF EXISTS device_info_logs;