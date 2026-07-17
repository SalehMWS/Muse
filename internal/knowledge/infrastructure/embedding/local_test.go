package embedding_test

import (
	"context"
	"math"
	"testing"

	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/embedding"
)

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func TestLocalEmbedder_DeterministicAndDimension(t *testing.T) {
	e := embedding.NewLocalEmbedder(128)
	if e.Dimension() != 128 {
		t.Fatalf("Dimension() = %d, want 128", e.Dimension())
	}

	a, err := e.Embed(context.Background(), []string{"the brand voice is warm"})
	if err != nil {
		t.Fatalf("Embed() unexpected error: %v", err)
	}
	b, _ := e.Embed(context.Background(), []string{"the brand voice is warm"})
	if cosine(a[0], b[0]) < 0.999 {
		t.Fatal("identical text should embed to (near) identical vectors")
	}
}

func TestLocalEmbedder_LexicalSimilarity(t *testing.T) {
	e := embedding.NewLocalEmbedder(256)
	vecs, err := e.Embed(context.Background(), []string{
		"our coffee is roasted fresh daily",
		"fresh roasted coffee every day",
		"quarterly financial tax report",
	})
	if err != nil {
		t.Fatalf("Embed() unexpected error: %v", err)
	}

	related := cosine(vecs[0], vecs[1])
	unrelated := cosine(vecs[0], vecs[2])
	if related <= unrelated {
		t.Fatalf("related texts (%.3f) should score higher than unrelated (%.3f)", related, unrelated)
	}
}
