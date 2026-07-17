package application

import (
	"context"
	"errors"
	"testing"

	"github.com/SalehMWS/Muse/internal/worker/domain"
)

func TestDispatcher(t *testing.T) {
	dispatcher := NewDispatcher()

	called := false
	dispatcher.Register(domain.TypeInstagramPublish, HandlerFunc(func(context.Context, domain.Job) error {
		called = true
		return nil
	}))

	if err := dispatcher.Dispatch(context.Background(), domain.Job{Type: domain.TypeInstagramPublish}); err != nil {
		t.Fatalf("Dispatch() unexpected error: %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}

	if err := dispatcher.Dispatch(context.Background(), domain.Job{Type: "nope"}); !errors.Is(err, ErrNoHandler) {
		t.Fatalf("Dispatch() unknown error = %v, want %v", err, ErrNoHandler)
	}
}
