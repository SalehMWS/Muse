package application_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
)

func TestConnectUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		oauth := &fakeOAuthClient{authURL: "https://www.instagram.com/oauth/authorize"}
		uc := application.NewConnectUseCase(oauth, fakeStateSigner{userID: userID})

		out, err := uc.Execute(context.Background(), userID)
		if err != nil {
			t.Fatalf("Execute() unexpected error: %v", err)
		}
		if out.State == "" {
			t.Fatal("Execute() returned empty state")
		}
		if !strings.Contains(out.AuthorizationURL, out.State) {
			t.Fatalf("AuthorizationURL %q does not embed state %q", out.AuthorizationURL, out.State)
		}
	})

	t.Run("sign error propagates", func(t *testing.T) {
		signErr := errors.New("sign failed")
		uc := application.NewConnectUseCase(&fakeOAuthClient{}, fakeStateSigner{signErr: signErr})

		if _, err := uc.Execute(context.Background(), uuid.New()); !errors.Is(err, signErr) {
			t.Fatalf("Execute() error = %v, want %v", err, signErr)
		}
	})
}
