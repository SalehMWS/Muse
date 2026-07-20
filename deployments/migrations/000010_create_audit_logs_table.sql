-- +goose Up
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID,
    action TEXT NOT NULL,
    result TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL DEFAULT '',
    correlation_id TEXT NOT NULL DEFAULT '',
    trace_id TEXT NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_user_created_at ON audit_logs (user_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs (action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);

-- +goose StatementBegin
CREATE FUNCTION reject_audit_log_mutation() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only';
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER audit_logs_reject_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION reject_audit_log_mutation();

CREATE TRIGGER audit_logs_reject_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW EXECUTE FUNCTION reject_audit_log_mutation();

-- +goose Down
DROP TRIGGER IF EXISTS audit_logs_reject_delete ON audit_logs;
DROP TRIGGER IF EXISTS audit_logs_reject_update ON audit_logs;
DROP FUNCTION IF EXISTS reject_audit_log_mutation;
DROP TABLE IF EXISTS audit_logs;
