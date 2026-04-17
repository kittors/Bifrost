-- +goose Up
CREATE TABLE sessions (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id text NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    refresh_token_hash text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('active', 'revoked', 'expired')),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz
);

CREATE INDEX idx_sessions_user_device ON sessions (user_id, device_id);
CREATE INDEX idx_sessions_status_expires_at ON sessions (status, expires_at);

-- +goose Down
DROP TABLE IF EXISTS sessions;
