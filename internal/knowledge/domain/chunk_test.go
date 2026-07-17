package domain_test

import (
	"strings"
	"testing"

	"github.com/SalehMWS/Muse/internal/knowledge/domain"
)

func TestSplitIntoChunks(t *testing.T) {
	words := make([]string, 500)
	for i := range words {
		words[i] = "w"
	}
	text := strings.Join(words, " ")

	chunks := domain.SplitIntoChunks(text, 200, 40)
	// step = 160; windows [0:200], [160:360], [320:500] -> 3 chunks.
	if len(chunks) != 3 {
		t.Fatalf("chunks = %d, want 3 (size 200, overlap 40, 500 words)", len(chunks))
	}

	if got := domain.SplitIntoChunks("", 200, 40); got != nil {
		t.Fatalf("empty text should yield no chunks, got %v", got)
	}

	short := domain.SplitIntoChunks("only a few words", 200, 40)
	if len(short) != 1 {
		t.Fatalf("short text chunks = %d, want 1", len(short))
	}
}
