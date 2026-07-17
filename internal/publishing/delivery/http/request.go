package http

type PublishRequest struct {
	InstagramAccountID string `json:"instagram_account_id"`
	MediaType          string `json:"media_type"`
}
