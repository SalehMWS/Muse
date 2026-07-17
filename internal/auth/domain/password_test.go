package domain_test

import (
	"errors"
	"testing"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		plain   string
		wantErr error
	}{
		{name: "meets policy", plain: "Str0ng!Passw0rd"},
		{name: "too short", plain: "Sh0rt!Aa", wantErr: domain.ErrWeakPassword},
		{name: "missing uppercase", plain: "weak12345!!!", wantErr: domain.ErrWeakPassword},
		{name: "missing lowercase", plain: "WEAK12345!!!", wantErr: domain.ErrWeakPassword},
		{name: "missing digit", plain: "WeakPassword!!!", wantErr: domain.ErrWeakPassword},
		{name: "missing special char", plain: "WeakPassword1234", wantErr: domain.ErrWeakPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidatePassword(tt.plain)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ValidatePassword(%q) error = %v, want %v", tt.plain, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("ValidatePassword(%q) unexpected error: %v", tt.plain, err)
			}
		})
	}
}
