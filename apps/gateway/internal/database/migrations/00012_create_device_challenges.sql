-- +goose Up
CREATE TABLE device_challenges (
    id text PRIMARY KEY,
    device_id text NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    challenge text NOT NULL,
    expires_at timestamptz NOT NULL,
    verified_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_device_challenges_device_id ON device_challenges (device_id);
CREATE INDEX idx_device_challenges_expires_at ON device_challenges (expires_at);

-- +goose Down
DROP TABLE IF EXISTS device_challenges;
