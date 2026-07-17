-- +goose Up
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual',
    status TEXT NOT NULL DEFAULT 'pending',
    chunk_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_documents_user_id ON documents (user_id);
CREATE INDEX idx_documents_status ON documents (status);

-- +goose Down
DROP TABLE IF EXISTS documents;
