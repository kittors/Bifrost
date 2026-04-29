-- +goose Up
CREATE TABLE system_settings (
    key text PRIMARY KEY,
    value jsonb NOT NULL,
    updated_by text REFERENCES users(id),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS system_settings;
