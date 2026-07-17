package application

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCursorRoundTrip(t *testing.T) {
	id := uuid.New()
	created := time.Unix(0, 1_700_000_000_123_456_789)

	gotCreated, gotID, err := decodeCursor(encodeCursor(created, id))
	if err != nil {
		t.Fatalf("decodeCursor() unexpected error: %v", err)
	}
	if !gotCreated.Equal(created) {
		t.Fatalf("createdAt = %v, want %v", gotCreated, created)
	}
	if gotID != id {
		t.Fatalf("id = %v, want %v", gotID, id)
	}
}

func TestDecodeCursor_Invalid(t *testing.T) {
	for _, bad := range []string{"###", "", "bm90LWEtY3Vyc29y"} {
		if _, _, err := decodeCursor(bad); err != ErrInvalidCursor {
			t.Fatalf("decodeCursor(%q) error = %v, want %v", bad, err, ErrInvalidCursor)
		}
	}
}
