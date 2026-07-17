package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const defaultTopK = 5

type QueryUseCase struct {
	embedder Embedder
	store    VectorStore
	topK     int
}

func NewQueryUseCase(embedder Embedder, store VectorStore, topK int) *QueryUseCase {
	if topK <= 0 {
		topK = defaultTopK
	}
	return &QueryUseCase{embedder: embedder, store: store, topK: topK}
}

type QueryInput struct {
	UserID uuid.UUID
	Query  string
	TopK   int
}

type QueryOutput struct {
	Hits    []SearchHit
	Context string
}

func (uc *QueryUseCase) Execute(ctx context.Context, in QueryInput) (QueryOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return QueryOutput{}, ErrEmptyContent
	}

	topK := in.TopK
	if topK <= 0 {
		topK = uc.topK
	}

	embeddings, err := uc.embedder.Embed(ctx, []string{query})
	if err != nil {
		return QueryOutput{}, fmt.Errorf("%w: %v", ErrEmbedding, err)
	}
	if len(embeddings) == 0 {
		return QueryOutput{}, fmt.Errorf("%w: no query embedding produced", ErrEmbedding)
	}

	hits, err := uc.store.Search(ctx, in.UserID, embeddings[0], topK)
	if err != nil {
		return QueryOutput{}, fmt.Errorf("%w: %v", ErrVectorStore, err)
	}

	return QueryOutput{Hits: hits, Context: buildContext(hits)}, nil
}

func buildContext(hits []SearchHit) string {
	parts := make([]string, 0, len(hits))
	for _, hit := range hits {
		parts = append(parts, hit.Content)
	}
	return strings.Join(parts, "\n\n---\n\n")
}
