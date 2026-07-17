package application

import (
	"context"
	"time"
)

type Token struct {
	AccessToken string
	ExpiresIn   time.Duration
}

type Profile struct {
	UserID      string
	Username    string
	AccountType string
}

type OAuthClient interface {
	AuthorizationURL(state string) string
	ExchangeCode(ctx context.Context, code string) (Token, error)
	FetchProfile(ctx context.Context, accessToken string) (Profile, error)
	RefreshToken(ctx context.Context, accessToken string) (Token, error)
}
