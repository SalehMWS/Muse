-- name: CreateDocument :one
INSERT INTO documents (id, user_id, title, source, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDocumentByIDForUser :one
SELECT * FROM documents WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListDocumentsByUser :many
SELECT * FROM documents WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateDocumentStatus :one
UPDATE documents
SET status = $2, chunk_count = $3, last_error = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDocumentForUser :exec
DELETE FROM documents WHERE id = $1 AND user_id = $2;
