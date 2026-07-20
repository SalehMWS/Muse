package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	milvus "github.com/milvus-io/milvus-sdk-go/v2/client"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	httpdelivery "github.com/SalehMWS/Muse/internal/knowledge/delivery/http"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/embedding"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/instrumented"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/knowledge/infrastructure/vectorstore"
	"github.com/SalehMWS/Muse/internal/shared/config"
	"github.com/SalehMWS/Muse/internal/shared/metrics"
)

type Module struct {
	Handler *httpdelivery.Handler
	closer  func()
	store   application.VectorStore
}

func New(pool *pgxpool.Pool, cfg config.Knowledge, ai config.AI, recorder *metrics.Metrics) (*Module, error) {
	var (
		knowledgeRecorder *metrics.Knowledge
		businessRecorder  *metrics.Business
	)
	if recorder != nil {
		knowledgeRecorder = recorder.Knowledge
		businessRecorder = recorder.Business
	}

	embedderName := cfg.Embedder
	if embedderName == "" {
		embedderName = "local"
	}
	storeName := cfg.VectorStore
	if storeName == "" {
		storeName = "memory"
	}

	embedder := instrumented.NewEmbedder(newEmbedder(cfg, ai), embedderName, knowledgeRecorder)

	rawStore, closer, err := newStore(cfg)
	if err != nil {
		return nil, err
	}
	store := instrumented.NewVectorStore(rawStore, storeName, knowledgeRecorder)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := store.EnsureReady(ctx, embedder.Dimension()); err != nil {
		if closer != nil {
			closer()
		}
		return nil, fmt.Errorf("knowledge: vector store: %w", err)
	}

	repo := postgres.NewDocumentRepository(pool)
	return &Module{
		Handler: httpdelivery.NewHandler(
			application.NewIngestUseCase(repo, embedder, store, cfg.ChunkSize, cfg.ChunkOverlap, knowledgeRecorder, businessRecorder),
			application.NewQueryUseCase(embedder, store, cfg.TopK, knowledgeRecorder),
			application.NewListDocumentsUseCase(repo),
			application.NewDeleteDocumentUseCase(repo, store),
		),
		closer: closer,
		store:  store,
	}, nil
}

func newEmbedder(cfg config.Knowledge, ai config.AI) application.Embedder {
	if cfg.Embedder == "openai" {
		return embedding.NewOpenAIEmbedder(ai.BaseURL, ai.APIKey, cfg.EmbeddingModel, cfg.EmbeddingDim, ai.HTTPTimeout)
	}
	return embedding.NewLocalEmbedder(cfg.EmbeddingDim)
}

func newStore(cfg config.Knowledge) (application.VectorStore, func(), error) {
	if cfg.VectorStore == "milvus" {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		client, err := milvus.NewClient(ctx, milvus.Config{Address: cfg.MilvusAddr})
		if err != nil {
			return nil, nil, fmt.Errorf("knowledge: connect milvus: %w", err)
		}
		return vectorstore.NewMilvusStore(client, cfg.MilvusCollection), func() { _ = client.Close() }, nil
	}
	return vectorstore.NewMemoryStore(), nil, nil
}

func (m *Module) Ping(ctx context.Context) error {
	if m.store == nil {
		return nil
	}
	return m.store.Ping(ctx)
}

func (m *Module) Close() {
	if m.closer != nil {
		m.closer()
	}
}

func (m *Module) RegisterRoutes(router fiber.Router, requireAuth fiber.Handler) {
	httpdelivery.RegisterRoutes(router, m.Handler, requireAuth)
}
