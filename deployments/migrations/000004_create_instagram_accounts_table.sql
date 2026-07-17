-- +goose Up
CREATE TABLE instagram_accounts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    instagram_user_id TEXT NOT NULL,
    username TEXT NOT NULL,
    account_type TEXT,
    access_token TEXT NOT NULL,
    token_expires_at TIMESTAMPTZ NOT NULL,
    scopes TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    connected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_refreshed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_instagram_accounts_user_ig_user UNIQUE (user_id, instagram_user_id)
);

CREATE INDEX idx_instagram_accounts_user_id ON instagram_accounts (user_id);
CREATE INDEX idx_instagram_accounts_token_expires_at ON instagram_accounts (token_expires_at);

-- +goose Down
DROP TABLE IF EXISTS instagram_accounts;
