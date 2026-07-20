package vectorstore

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
)

const (
	fieldID         = "id"
	fieldUserID     = "user_id"
	fieldDocumentID = "document_id"
	fieldChunkIndex = "chunk_index"
	fieldContent    = "content"
	fieldEmbedding  = "embedding"
	maxContentBytes = 60000
)

type MilvusStore struct {
	client     client.Client
	collection string
}

var _ application.VectorStore = (*MilvusStore)(nil)

func NewMilvusStore(c client.Client, collection string) *MilvusStore {
	return &MilvusStore{client: c, collection: collection}
}

func (s *MilvusStore) Ping(ctx context.Context) error {
	if _, err := s.client.HasCollection(ctx, s.collection); err != nil {
		return fmt.Errorf("has collection: %w", err)
	}
	return nil
}

func (s *MilvusStore) EnsureReady(ctx context.Context, dimension int) error {
	has, err := s.client.HasCollection(ctx, s.collection)
	if err != nil {
		return fmt.Errorf("has collection: %w", err)
	}

	if !has {
		schema := entity.NewSchema().WithName(s.collection).
			WithField(entity.NewField().WithName(fieldID).WithDataType(entity.FieldTypeVarChar).WithIsPrimaryKey(true).WithMaxLength(128)).
			WithField(entity.NewField().WithName(fieldUserID).WithDataType(entity.FieldTypeVarChar).WithMaxLength(64)).
			WithField(entity.NewField().WithName(fieldDocumentID).WithDataType(entity.FieldTypeVarChar).WithMaxLength(64)).
			WithField(entity.NewField().WithName(fieldChunkIndex).WithDataType(entity.FieldTypeInt64)).
			WithField(entity.NewField().WithName(fieldContent).WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535)).
			WithField(entity.NewField().WithName(fieldEmbedding).WithDataType(entity.FieldTypeFloatVector).WithDim(int64(dimension)))

		if err := s.client.CreateCollection(ctx, schema, 1); err != nil {
			return fmt.Errorf("create collection: %w", err)
		}

		index, err := entity.NewIndexAUTOINDEX(entity.COSINE)
		if err != nil {
			return fmt.Errorf("build index: %w", err)
		}
		if err := s.client.CreateIndex(ctx, s.collection, fieldEmbedding, index, false); err != nil {
			return fmt.Errorf("create index: %w", err)
		}
	}

	if err := s.client.LoadCollection(ctx, s.collection, false); err != nil {
		return fmt.Errorf("load collection: %w", err)
	}
	return nil
}

func (s *MilvusStore) Upsert(ctx context.Context, records []application.VectorRecord) error {
	if len(records) == 0 {
		return nil
	}

	ids := make([]string, len(records))
	userIDs := make([]string, len(records))
	documentIDs := make([]string, len(records))
	chunkIndexes := make([]int64, len(records))
	contents := make([]string, len(records))
	embeddings := make([][]float32, len(records))

	for i, record := range records {
		ids[i] = record.ID
		userIDs[i] = record.UserID.String()
		documentIDs[i] = record.DocumentID.String()
		chunkIndexes[i] = int64(record.ChunkIndex)
		contents[i] = truncate(record.Content, maxContentBytes)
		embeddings[i] = record.Embedding
	}

	dim := 0
	if len(embeddings) > 0 {
		dim = len(embeddings[0])
	}

	_, err := s.client.Insert(ctx, s.collection, "",
		entity.NewColumnVarChar(fieldID, ids),
		entity.NewColumnVarChar(fieldUserID, userIDs),
		entity.NewColumnVarChar(fieldDocumentID, documentIDs),
		entity.NewColumnInt64(fieldChunkIndex, chunkIndexes),
		entity.NewColumnVarChar(fieldContent, contents),
		entity.NewColumnFloatVector(fieldEmbedding, dim, embeddings),
	)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	if err := s.client.Flush(ctx, s.collection, false); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return nil
}

func (s *MilvusStore) Search(ctx context.Context, userID uuid.UUID, embedding []float32, topK int) ([]application.SearchHit, error) {
	searchParam, err := entity.NewIndexAUTOINDEXSearchParam(1)
	if err != nil {
		return nil, fmt.Errorf("search param: %w", err)
	}

	expr := fmt.Sprintf(`%s == "%s"`, fieldUserID, userID.String())
	results, err := s.client.Search(ctx, s.collection, nil, expr,
		[]string{fieldDocumentID, fieldChunkIndex, fieldContent},
		[]entity.Vector{entity.FloatVector(embedding)},
		fieldEmbedding, entity.COSINE, topK, searchParam,
	)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	if len(results) == 0 {
		return nil, nil
	}

	result := results[0]
	documentCol := result.Fields.GetColumn(fieldDocumentID)
	chunkCol := result.Fields.GetColumn(fieldChunkIndex)
	contentCol := result.Fields.GetColumn(fieldContent)

	hits := make([]application.SearchHit, 0, result.ResultCount)
	for i := 0; i < result.ResultCount; i++ {
		documentRaw, err := documentCol.GetAsString(i)
		if err != nil {
			continue
		}
		documentID, err := uuid.Parse(documentRaw)
		if err != nil {
			continue
		}
		chunkIndex, _ := chunkCol.GetAsInt64(i)
		content, _ := contentCol.GetAsString(i)

		var score float32
		if i < len(result.Scores) {
			score = result.Scores[i]
		}

		hits = append(hits, application.SearchHit{
			DocumentID: documentID,
			ChunkIndex: int(chunkIndex),
			Content:    content,
			Score:      score,
		})
	}
	return hits, nil
}

func (s *MilvusStore) DeleteByDocument(ctx context.Context, userID, documentID uuid.UUID) error {
	expr := fmt.Sprintf(`%s == "%s" && %s == "%s"`, fieldUserID, userID.String(), fieldDocumentID, documentID.String())
	if err := s.client.Delete(ctx, s.collection, "", expr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return s.client.Flush(ctx, s.collection, false)
}

func truncate(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	return value[:maxBytes]
}
