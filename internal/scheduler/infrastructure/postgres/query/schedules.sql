-- name: CreateSchedule :one
INSERT INTO schedules (
    id, user_id, content_id, instagram_account_id, scheduled_for,
    timezone, cron_expression, media_type, max_retries
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetScheduleByIDForUser :one
SELECT * FROM schedules WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListSchedulesByContentForUser :many
SELECT * FROM schedules
WHERE content_id = $1 AND user_id = $2
ORDER BY scheduled_for DESC;

-- name: ClaimDueSchedules :many
UPDATE schedules
SET status = 'publishing', updated_at = now()
WHERE id IN (
    SELECT s.id FROM schedules s
    WHERE s.status = 'scheduled' AND s.scheduled_for <= $1
    ORDER BY s.scheduled_for
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: MarkSchedulePublished :exec
UPDATE schedules SET status = 'published', last_error = NULL, updated_at = now()
WHERE id = $1;

-- name: MarkScheduleFailed :exec
UPDATE schedules SET status = 'failed', last_error = $2, updated_at = now()
WHERE id = $1;

-- name: RescheduleSchedule :exec
UPDATE schedules
SET status = 'scheduled', scheduled_for = $2, retry_count = 0, next_retry_at = NULL, last_error = NULL, updated_at = now()
WHERE id = $1;

-- name: RetrySchedule :exec
UPDATE schedules
SET status = 'scheduled', retry_count = $2, scheduled_for = $3, next_retry_at = $3, last_error = $4, updated_at = now()
WHERE id = $1;

-- name: CancelSchedule :exec
UPDATE schedules SET status = 'cancelled', updated_at = now()
WHERE id = $1 AND user_id = $2 AND status = 'scheduled';
