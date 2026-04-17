-- +goose Up
CREATE TABLE audit_events (
    id text PRIMARY KEY,
    request_id text NOT NULL,
    type text NOT NULL,
    actor_user_id text REFERENCES users(id),
    actor_device_id text REFERENCES devices(id),
    target_type text NOT NULL,
    target_id text,
    service_id text REFERENCES services(id),
    result text NOT NULL CHECK (result IN ('success', 'failure')),
    error_code text,
    source_ip inet,
    user_agent text,
    summary text NOT NULL DEFAULT '',
    details jsonb NOT NULL DEFAULT '{}'::jsonb,
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_events_occurred_at ON audit_events (occurred_at);
CREATE INDEX idx_audit_events_type_occurred_at ON audit_events (type, occurred_at);
CREATE INDEX idx_audit_events_actor_user_id_occurred_at ON audit_events (actor_user_id, occurred_at);
CREATE INDEX idx_audit_events_service_id_occurred_at ON audit_events (service_id, occurred_at);

-- +goose Down
DROP TABLE IF EXISTS audit_events;
