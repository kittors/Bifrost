-- +goose Up
CREATE TABLE devices (
    id text PRIMARY KEY,
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    os text NOT NULL,
    client_version text NOT NULL,
    public_key text NOT NULL,
    public_key_fingerprint text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('trusted', 'disabled')),
    last_seen_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_user_id ON devices (user_id);

-- +goose Down
DROP TABLE IF EXISTS devices;
