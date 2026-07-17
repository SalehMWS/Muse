package domain_test

import (
	"errors"
	"testing"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr error
	}{
		{name: "valid lowercase", raw: "user@example.com", want: "user@example.com"},
		{name: "normalizes case and whitespace", raw: "  User@Example.COM  ", want: "user@example.com"},
		{name: "empty", raw: "", wantErr: domain.ErrInvalidEmail},
		{name: "missing at sign", raw: "not-an-email", wantErr: domain.ErrInvalidEmail},
		{name: "missing domain", raw: "user@", wantErr: domain.ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewEmail(tt.raw)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("NewEmail(%q) error = %v, want %v", tt.raw, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewEmail(%q) unexpected error: %v", tt.raw, err)
			}
			if got.String() != tt.want {
				t.Fatalf("NewEmail(%q) = %q, want %q", tt.raw, got.String(), tt.want)
			}
		})
	}
}
