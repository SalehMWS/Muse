-- name: CreateContent :one
INSERT INTO contents (id, user_id, title, caption, status, language, content_type, visibility, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetContentByIDForUser :one
SELECT * FROM contents WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL LIMIT 1;

-- name: UpdateContent :one
UPDATE contents
SET title = $3,
    caption = $4,
    status = $5,
    language = $6,
    content_type = $7,
    visibility = $8,
    tags = $9,
    updated_at = now()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: ListContents :many
SELECT * FROM contents
WHERE user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('language')::text IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('content_type')::text IS NULL OR content_type = sqlc.narg('content_type'))
  AND (sqlc.narg('tag')::text IS NULL OR sqlc.narg('tag') = ANY (tags))
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim');

-- name: ListContentsAfter :many
SELECT * FROM contents
WHERE user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('language')::text IS NULL OR language = sqlc.narg('language'))
  AND (sqlc.narg('content_type')::text IS NULL OR content_type = sqlc.narg('content_type'))
  AND (sqlc.narg('tag')::text IS NULL OR sqlc.narg('tag') = ANY (tags))
  AND (
    created_at < sqlc.arg('cursor_created_at')
    OR (created_at = sqlc.arg('cursor_created_at') AND id < sqlc.arg('cursor_id'))
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim');
