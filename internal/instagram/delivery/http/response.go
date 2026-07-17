package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type AccountResponse struct {
	ID              string  `json:"id"`
	InstagramUserID string  `json:"instagram_user_id"`
	Username        string  `json:"username"`
	AccountType     *string `json:"account_type,omitempty"`
	Status          string  `json:"status"`
	ConnectedAt     string  `json:"connected_at"`
	TokenExpiresAt  string  `json:"token_expires_at"`
	LastRefreshedAt *string `json:"last_refreshed_at,omitempty"`
}

func newAccountResponse(account domain.ConnectedAccount, now time.Time) AccountResponse {
	resp := AccountResponse{
		ID:              account.ID.String(),
		InstagramUserID: account.InstagramUserID,
		Username:        account.Username,
		AccountType:     account.AccountType,
		Status:          string(account.EffectiveStatus(now)),
		ConnectedAt:     account.ConnectedAt.Format(time.RFC3339),
		TokenExpiresAt:  account.TokenExpiresAt.Format(time.RFC3339),
	}
	if account.LastRefreshedAt != nil {
		formatted := account.LastRefreshedAt.Format(time.RFC3339)
		resp.LastRefreshedAt = &formatted
	}
	return resp
}

type ConnectResponse struct {
	AuthorizationURL string `json:"authorization_url"`
	State            string `json:"state"`
}
