package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusExpired AccountStatus = "expired"
	AccountStatusRevoked AccountStatus = "revoked"
)

type ConnectedAccount struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	InstagramUserID string
	Username        string
	AccountType     *string
	AccessToken     string
	TokenExpiresAt  time.Time
	Scopes          *string
	Status          AccountStatus
	ConnectedAt     time.Time
	LastRefreshedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a ConnectedAccount) IsTokenExpired(now time.Time) bool {
	return !now.Before(a.TokenExpiresAt)
}

func (a ConnectedAccount) EffectiveStatus(now time.Time) AccountStatus {
	if a.Status == AccountStatusRevoked {
		return AccountStatusRevoked
	}
	if a.IsTokenExpired(now) {
		return AccountStatusExpired
	}
	return a.Status
}

func IsPublishableAccountType(accountType string) bool {
	switch accountType {
	case "BUSINESS", "CREATOR", "MEDIA_CREATOR":
		return true
	default:
		return false
	}
}
