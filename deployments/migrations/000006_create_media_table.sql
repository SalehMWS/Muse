-- +goose Up
CREATE TABLE media (
    id UUID PRIMARY KEY,
    content_id UUID NOT NULL REFERENCES contents (id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    media_type TEXT NOT NULL DEFAULT 'image',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_media_content_id ON media (content_id);
CREATE INDEX idx_media_content_position ON media (content_id, position, created_at);

-- +goose Down
DROP TABLE IF EXISTS media;
