-- +goose Up
CREATE TABLE contents (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    caption TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    language TEXT NOT NULL DEFAULT 'en',
    content_type TEXT NOT NULL DEFAULT 'image',
    visibility TEXT NOT NULL DEFAULT 'private',
    tags TEXT[] NOT NULL DEFAULT '{}',
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_contents_user_id ON contents (user_id);
CREATE INDEX idx_contents_status ON contents (status);
CREATE INDEX idx_contents_created_at ON contents (created_at);
CREATE INDEX idx_contents_user_created ON contents (user_id, created_at DESC, id DESC);
CREATE INDEX idx_contents_tags ON contents USING GIN (tags);

-- +goose Down
DROP TABLE IF EXISTS contents;
