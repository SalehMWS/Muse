-- name: UpsertInstagramAccount :one
INSERT INTO instagram_accounts (
    id, user_id, instagram_user_id, username, account_type,
    access_token, token_expires_at, scopes, status, connected_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
ON CONFLICT (user_id, instagram_user_id) DO UPDATE
SET username = EXCLUDED.username,
    account_type = EXCLUDED.account_type,
    access_token = EXCLUDED.access_token,
    token_expires_at = EXCLUDED.token_expires_at,
    scopes = EXCLUDED.scopes,
    status = EXCLUDED.status,
    connected_at = now(),
    last_refreshed_at = NULL,
    updated_at = now()
RETURNING *;

-- name: GetInstagramAccountByIDForUser :one
SELECT * FROM instagram_accounts WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListInstagramAccountsByUser :many
SELECT * FROM instagram_accounts WHERE user_id = $1 ORDER BY connected_at DESC;

-- name: UpdateInstagramAccountToken :one
UPDATE instagram_accounts
SET access_token = $2,
    token_expires_at = $3,
    status = $4,
    last_refreshed_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteInstagramAccountForUser :exec
DELETE FROM instagram_accounts WHERE id = $1 AND user_id = $2;
