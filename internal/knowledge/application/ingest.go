package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/domain"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type IngestUseCase struct {
	repo         DocumentRepository
	embedder     Embedder
	store        VectorStore
	chunkSize    int
	chunkOverlap int
	recorder     *metrics.Knowledge
	business     *metrics.Business
}

func NewIngestUseCase(repo DocumentRepository, embedder Embedder, store VectorStore, chunkSize, chunkOverlap int, recorder *metrics.Knowledge, business *metrics.Business) *IngestUseCase {
	return &IngestUseCase{
		repo:         repo,
		embedder:     embedder,
		store:        store,
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		recorder:     recorder,
		business:     business,
	}
}

type IngestInput struct {
	UserID  uuid.UUID
	Title   string
	Source  string
	Content string
}

func (uc *IngestUseCase) Execute(ctx context.Context, in IngestInput) (domain.Document, error) {
	content := strings.TrimSpace(in.Content)
	if content == "" {
		return domain.Document{}, ErrEmptyContent
	}

	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = "manual"
	}

	created, err := uc.repo.Create(ctx, domain.Document{
		ID:     uuid.New(),
		UserID: in.UserID,
		Title:  strings.TrimSpace(in.Title),
		Source: source,
		Status: domain.StatusPending,
	})
	if err != nil {
		return domain.Document{}, err
	}

	chunks := domain.SplitIntoChunks(content, uc.chunkSize, uc.chunkOverlap)
	if len(chunks) == 0 {
		chunks = []string{content}
	}

	embeddings, err := uc.embedder.Embed(ctx, chunks)
	if err != nil {
		return uc.fail(ctx, created.ID, ErrEmbedding, err)
	}
	if len(embeddings) != len(chunks) {
		return uc.fail(ctx, created.ID, ErrEmbedding, fmt.Errorf("embedder returned %d vectors for %d chunks", len(embeddings), len(chunks)))
	}

	records := make([]VectorRecord, 0, len(chunks))
	for i, chunk := range chunks {
		records = append(records, VectorRecord{
			ID:         created.ID.String() + "-" + strconv.Itoa(i),
			UserID:     in.UserID,
			DocumentID: created.ID,
			ChunkIndex: i,
			Content:    chunk,
			Embedding:  embeddings[i],
		})
	}

	if err := uc.store.Upsert(ctx, records); err != nil {
		return uc.fail(ctx, created.ID, ErrVectorStore, err)
	}

	indexed, err := uc.repo.UpdateStatus(ctx, created.ID, domain.StatusIndexed, len(chunks), nil)
	if err != nil {
		return domain.Document{}, err
	}

	uc.recorder.Ingested(metrics.OutcomeSuccess, len(chunks))
	uc.business.Record(metrics.EventDocumentIngested)

	return indexed, nil
}

func (uc *IngestUseCase) fail(ctx context.Context, id uuid.UUID, sentinel, cause error) (domain.Document, error) {
	uc.recorder.Ingested(metrics.OutcomeFailure, 0)

	message := cause.Error()
	_, _ = uc.repo.UpdateStatus(ctx, id, domain.StatusFailed, 0, &message)
	return domain.Document{}, fmt.Errorf("%w: %v", sentinel, cause)
}
