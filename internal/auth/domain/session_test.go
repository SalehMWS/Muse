package domain_test

import (
	"testing"
	"time"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestGenerateRefreshToken(t *testing.T) {
	first, err := domain.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() unexpected error: %v", err)
	}
	if first == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}

	second, err := domain.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() unexpected error: %v", err)
	}
	if first == second {
		t.Fatal("GenerateRefreshToken() returned identical tokens on successive calls")
	}
}

func TestHashRefreshToken(t *testing.T) {
	raw := "some-raw-refresh-token"

	first := domain.HashRefreshToken(raw)
	second := domain.HashRefreshToken(raw)

	if first != second {
		t.Fatalf("HashRefreshToken(%q) is not deterministic: %q != %q", raw, first, second)
	}
	if first == raw {
		t.Fatalf("HashRefreshToken(%q) returned the raw value unchanged", raw)
	}
}

func TestSessionIsExpired(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{name: "in the future", expiresAt: now.Add(time.Hour), want: false},
		{name: "in the past", expiresAt: now.Add(-time.Hour), want: true},
		{name: "exactly now", expiresAt: now, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := domain.Session{ExpiresAt: tt.expiresAt}
			if got := session.IsExpired(now); got != tt.want {
				t.Fatalf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
