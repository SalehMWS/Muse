-- +goose Up
CREATE TABLE publications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    content_id UUID NOT NULL REFERENCES contents (id) ON DELETE CASCADE,
    instagram_account_id UUID NOT NULL REFERENCES instagram_accounts (id) ON DELETE CASCADE,
    platform TEXT NOT NULL DEFAULT 'instagram',
    platform_post_id TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    permalink TEXT,
    response_json JSONB,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_publications_user_id ON publications (user_id);
CREATE INDEX idx_publications_content_id ON publications (content_id);
CREATE INDEX idx_publications_account_id ON publications (instagram_account_id);

-- +goose Down
DROP TABLE IF EXISTS publications;
