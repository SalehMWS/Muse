-- name: AppendAuditLog :exec
INSERT INTO audit_logs (
    id, user_id, action, result, resource_type, resource_id,
    ip_address, user_agent, request_id, correlation_id, trace_id, metadata, created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;
