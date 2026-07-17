-- +goose Up
CREATE TABLE schedules (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content_id UUID NOT NULL REFERENCES contents (id) ON DELETE CASCADE,
    instagram_account_id UUID NOT NULL REFERENCES instagram_accounts (id) ON DELETE CASCADE,
    scheduled_for TIMESTAMPTZ NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    cron_expression TEXT,
    media_type TEXT,
    status TEXT NOT NULL DEFAULT 'scheduled',
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retries INTEGER NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_schedules_due ON schedules (status, scheduled_for);
CREATE INDEX idx_schedules_user_id ON schedules (user_id);
CREATE INDEX idx_schedules_content_id ON schedules (content_id);

-- +goose Down
DROP TABLE IF EXISTS schedules;
