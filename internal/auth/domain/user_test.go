package domain_test

import (
	"testing"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestUserCanAuthenticate(t *testing.T) {
	tests := []struct {
		status domain.Status
		want   bool
	}{
		{status: domain.StatusActive, want: true},
		{status: domain.StatusPending, want: false},
		{status: domain.StatusSuspended, want: false},
		{status: domain.StatusDisabled, want: false},
		{status: domain.StatusDeleted, want: false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			user := domain.User{Status: tt.status}
			if got := user.CanAuthenticate(); got != tt.want {
				t.Fatalf("CanAuthenticate() with status %q = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
