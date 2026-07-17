-- name: CreatePublication :one
INSERT INTO publications (
    id, user_id, content_id, instagram_account_id, platform,
    platform_post_id, status, permalink, response_json, published_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListPublicationsByContentForUser :many
SELECT * FROM publications
WHERE content_id = $1 AND user_id = $2
ORDER BY created_at DESC;
