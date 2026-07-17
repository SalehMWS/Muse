package domain_test

import (
	"testing"
	"time"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

func TestConnectedAccount_EffectiveStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		account domain.ConnectedAccount
		want    domain.AccountStatus
	}{
		{
			name:    "active and valid",
			account: domain.ConnectedAccount{Status: domain.AccountStatusActive, TokenExpiresAt: now.Add(time.Hour)},
			want:    domain.AccountStatusActive,
		},
		{
			name:    "active but expired token",
			account: domain.ConnectedAccount{Status: domain.AccountStatusActive, TokenExpiresAt: now.Add(-time.Hour)},
			want:    domain.AccountStatusExpired,
		},
		{
			name:    "revoked stays revoked even if token valid",
			account: domain.ConnectedAccount{Status: domain.AccountStatusRevoked, TokenExpiresAt: now.Add(time.Hour)},
			want:    domain.AccountStatusRevoked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.account.EffectiveStatus(now); got != tt.want {
				t.Fatalf("EffectiveStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsPublishableAccountType(t *testing.T) {
	publishable := []string{"BUSINESS", "CREATOR", "MEDIA_CREATOR"}
	for _, at := range publishable {
		if !domain.IsPublishableAccountType(at) {
			t.Fatalf("IsPublishableAccountType(%q) = false, want true", at)
		}
	}

	for _, at := range []string{"PERSONAL", "", "business"} {
		if domain.IsPublishableAccountType(at) {
			t.Fatalf("IsPublishableAccountType(%q) = true, want false", at)
		}
	}
}
