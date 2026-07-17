package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/scheduler/domain"
)

type ScheduleResponse struct {
	ID                 string  `json:"id"`
	ContentID          string  `json:"content_id"`
	InstagramAccountID string  `json:"instagram_account_id"`
	ScheduledFor       string  `json:"scheduled_for"`
	Timezone           string  `json:"timezone"`
	CronExpression     *string `json:"cron_expression,omitempty"`
	MediaType          *string `json:"media_type,omitempty"`
	Status             string  `json:"status"`
	RetryCount         int     `json:"retry_count"`
	MaxRetries         int     `json:"max_retries"`
	LastError          *string `json:"last_error,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

func newScheduleResponse(schedule domain.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID:                 schedule.ID.String(),
		ContentID:          schedule.ContentID.String(),
		InstagramAccountID: schedule.InstagramAccountID.String(),
		ScheduledFor:       schedule.ScheduledFor.Format(time.RFC3339),
		Timezone:           schedule.Timezone,
		CronExpression:     schedule.CronExpression,
		MediaType:          schedule.MediaType,
		Status:             string(schedule.Status),
		RetryCount:         schedule.RetryCount,
		MaxRetries:         schedule.MaxRetries,
		LastError:          schedule.LastError,
		CreatedAt:          schedule.CreatedAt.Format(time.RFC3339),
	}
}
