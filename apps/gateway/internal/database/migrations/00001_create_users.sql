-- +goose Up
CREATE TABLE users (
    id text PRIMARY KEY,
    username text NOT NULL UNIQUE,
    display_name text NOT NULL,
    email text,
    password_hash text NOT NULL,
    status text NOT NULL CHECK (status IN ('enabled', 'disabled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE INDEX idx_users_status ON users (status);

-- +goose Down
DROP TABLE IF EXISTS users;
