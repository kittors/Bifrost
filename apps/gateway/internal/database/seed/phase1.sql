INSERT INTO users (id, username, display_name, email, password_hash, status)
VALUES
    ('user_admin', 'admin', 'Administrator', 'admin@example.com', 'seed:change-me', 'enabled'),
    ('user_alice', 'alice', 'Alice', 'alice@example.com', 'seed:change-me', 'enabled'),
    ('user_bob', 'bob', 'Bob', 'bob@example.com', 'seed:change-me', 'enabled')
ON CONFLICT (id) DO UPDATE
SET
    username = EXCLUDED.username,
    display_name = EXCLUDED.display_name,
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO roles (id, name, display_name, description, is_system)
VALUES
    ('role_admin', 'admin', 'Administrator', 'Full access to all managed services', true),
    ('role_developer', 'developer', 'Developer', 'Access to GitLab and Docs', true),
    ('role_ops', 'ops', 'Operations', 'Access to Jenkins and Docs', true)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    is_system = EXCLUDED.is_system,
    updated_at = now();

INSERT INTO services (id, key, name, description, group_name, protocol, upstream_url, public_path, status)
VALUES
    ('service_gitlab', 'gitlab', 'GitLab', 'Mock GitLab upstream for local development', 'engineering', 'http', 'http://mock-gitlab:8080', '/s/gitlab', 'enabled'),
    ('service_jenkins', 'jenkins', 'Jenkins', 'Mock Jenkins upstream for local development', 'operations', 'http', 'http://mock-jenkins:8080', '/s/jenkins', 'enabled'),
    ('service_docs', 'docs', 'Docs', 'Mock Docs upstream for local development', 'shared', 'http', 'http://mock-docs:8080', '/s/docs', 'enabled')
ON CONFLICT (id) DO UPDATE
SET
    key = EXCLUDED.key,
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    group_name = EXCLUDED.group_name,
    protocol = EXCLUDED.protocol,
    upstream_url = EXCLUDED.upstream_url,
    public_path = EXCLUDED.public_path,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO user_roles (user_id, role_id)
VALUES
    ('user_admin', 'role_admin'),
    ('user_alice', 'role_developer'),
    ('user_bob', 'role_ops')
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO role_services (role_id, service_id)
VALUES
    ('role_admin', 'service_gitlab'),
    ('role_admin', 'service_jenkins'),
    ('role_admin', 'service_docs'),
    ('role_developer', 'service_gitlab'),
    ('role_developer', 'service_docs'),
    ('role_ops', 'service_jenkins'),
    ('role_ops', 'service_docs')
ON CONFLICT (role_id, service_id) DO NOTHING;

INSERT INTO user_service_overrides (user_id, service_id, effect, reason, created_by)
VALUES
    ('user_bob', 'service_jenkins', 'deny', 'Ops role can access Jenkins by default, but bob is explicitly denied for testing.', 'user_admin')
ON CONFLICT (user_id, service_id) DO UPDATE
SET
    effect = EXCLUDED.effect,
    reason = EXCLUDED.reason,
    created_by = EXCLUDED.created_by,
    created_at = now();

INSERT INTO system_settings (key, value, updated_by)
VALUES
    ('audit.retention.days', '{"days":180}'::jsonb, 'user_admin'),
    ('desktop.defaultTheme', '{"theme":"system"}'::jsonb, 'user_admin')
ON CONFLICT (key) DO UPDATE
SET
    value = EXCLUDED.value,
    updated_by = EXCLUDED.updated_by,
    updated_at = now();
