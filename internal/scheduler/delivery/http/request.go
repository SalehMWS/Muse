package http

type CreateScheduleRequest struct {
	InstagramAccountID string `json:"instagram_account_id"`
	ScheduledFor       string `json:"scheduled_for"`
	CronExpression     string `json:"cron_expression"`
	Timezone           string `json:"timezone"`
	MediaType          string `json:"media_type"`
	MaxRetries         int    `json:"max_retries"`
}
