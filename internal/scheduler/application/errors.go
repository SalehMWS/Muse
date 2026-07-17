package application

import "errors"

var (
	ErrScheduleNotFound     = errors.New("schedule not found")
	ErrContentNotFound      = errors.New("content not found")
	ErrInvalidCron          = errors.New("invalid cron expression")
	ErrInvalidTimezone      = errors.New("invalid timezone")
	ErrScheduleInPast       = errors.New("scheduled time must be in the future")
	ErrScheduleTimeRequired = errors.New("either scheduled_for or cron_expression is required")
)
