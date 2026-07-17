-- name: CreateMedia :one
INSERT INTO media (id, content_id, url, media_type, position)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMediaByIDForContent :one
SELECT * FROM media WHERE id = $1 AND content_id = $2 LIMIT 1;

-- name: ListMediaByContent :many
SELECT * FROM media WHERE content_id = $1 ORDER BY position ASC, created_at ASC;

-- name: DeleteMediaForContent :exec
DELETE FROM media WHERE id = $1 AND content_id = $2;
