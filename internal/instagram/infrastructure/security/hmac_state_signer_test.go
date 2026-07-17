package security

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHMACStateSigner_SignVerifyRoundTrip(t *testing.T) {
	signer := NewHMACStateSigner("state-secret", 10*time.Minute)
	userID := uuid.New()

	state, err := signer.Sign(userID)
	if err != nil {
		t.Fatalf("Sign() unexpected error: %v", err)
	}

	got, err := signer.Verify(state)
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if got != userID {
		t.Fatalf("Verify() = %v, want %v", got, userID)
	}
}

func TestHMACStateSigner_Expired(t *testing.T) {
	signer := NewHMACStateSigner("state-secret", time.Minute)
	base := time.Unix(1_700_000_000, 0)
	signer.now = func() time.Time { return base }

	state, err := signer.Sign(uuid.New())
	if err != nil {
		t.Fatalf("Sign() unexpected error: %v", err)
	}

	signer.now = func() time.Time { return base.Add(2 * time.Minute) }
	if _, err := signer.Verify(state); !errors.Is(err, ErrStateExpired) {
		t.Fatalf("Verify() error = %v, want %v", err, ErrStateExpired)
	}
}

func TestHMACStateSigner_TamperingRejected(t *testing.T) {
	signer := NewHMACStateSigner("state-secret", time.Minute)
	other := NewHMACStateSigner("different-secret", time.Minute)

	state, _ := signer.Sign(uuid.New())
	if _, err := other.Verify(state); !errors.Is(err, ErrStateSignature) {
		t.Fatalf("Verify() with wrong secret error = %v, want %v", err, ErrStateSignature)
	}

	if _, err := signer.Verify("no-separator"); !errors.Is(err, ErrMalformedState) {
		t.Fatalf("Verify() malformed error = %v, want %v", err, ErrMalformedState)
	}
}
