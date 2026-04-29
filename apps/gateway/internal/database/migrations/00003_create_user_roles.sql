-- +goose Up
CREATE TABLE user_roles (
    user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id text NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);

-- +goose Down
DROP TABLE IF EXISTS user_roles;
