-- +goose Up
CREATE TABLE role_services (
    role_id text NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    service_id text NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (role_id, service_id)
);

CREATE INDEX idx_role_services_service_id ON role_services (service_id);

-- +goose Down
DROP TABLE IF EXISTS role_services;
