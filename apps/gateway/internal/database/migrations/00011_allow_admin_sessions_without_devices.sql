-- +goose Up
ALTER TABLE sessions
    ALTER COLUMN device_id DROP NOT NULL;

-- +goose Down
ALTER TABLE sessions
    ALTER COLUMN device_id SET NOT NULL;
