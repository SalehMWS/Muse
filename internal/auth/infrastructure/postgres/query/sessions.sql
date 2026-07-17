-- name: CreateSession :one
INSERT INTO sessions (id, user_id, refresh_token_hash, device, ip_address, user_agent, expires_at, last_activity_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
RETURNING *;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM sessions WHERE refresh_token_hash = $1 LIMIT 1;

-- name: RotateSession :one
UPDATE sessions
SET refresh_token_hash = $2, expires_at = $3, last_activity_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteSessionByRefreshTokenHash :exec
DELETE FROM sessions WHERE refresh_token_hash = $1;
