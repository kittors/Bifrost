-- +goose Up
CREATE TABLE services (
    id text PRIMARY KEY,
    key text NOT NULL UNIQUE,
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    group_name text NOT NULL,
    protocol text NOT NULL CHECK (protocol IN ('http', 'https')),
    upstream_url text NOT NULL,
    public_path text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('enabled', 'disabled', 'archived')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_services_status ON services (status);

-- +goose Down
DROP TABLE IF EXISTS services;
