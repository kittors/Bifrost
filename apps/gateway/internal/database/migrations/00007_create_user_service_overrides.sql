-- +goose Up
CREATE TABLE user_service_overrides (
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id text NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    effect text NOT NULL CHECK (effect IN ('allow', 'deny')),
    reason text NOT NULL DEFAULT '',
    created_by text NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, service_id)
);

CREATE INDEX idx_user_service_overrides_created_by ON user_service_overrides (created_by);

-- +goose Down
DROP TABLE IF EXISTS user_service_overrides;
