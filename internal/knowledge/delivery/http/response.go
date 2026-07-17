package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/knowledge/domain"
)

type DocumentResponse struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Source     string  `json:"source"`
	Status     string  `json:"status"`
	ChunkCount int     `json:"chunk_count"`
	LastError  *string `json:"last_error,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func newDocumentResponse(document domain.Document) DocumentResponse {
	return DocumentResponse{
		ID:         document.ID.String(),
		Title:      document.Title,
		Source:     document.Source,
		Status:     string(document.Status),
		ChunkCount: document.ChunkCount,
		LastError:  document.LastError,
		CreatedAt:  document.CreatedAt.Format(time.RFC3339),
	}
}

type HitResponse struct {
	DocumentID string  `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
	Content    string  `json:"content"`
	Score      float32 `json:"score"`
}

type QueryResponse struct {
	Hits    []HitResponse `json:"hits"`
	Context string        `json:"context"`
}

func newQueryResponse(out application.QueryOutput) QueryResponse {
	hits := make([]HitResponse, 0, len(out.Hits))
	for _, hit := range out.Hits {
		hits = append(hits, HitResponse{
			DocumentID: hit.DocumentID.String(),
			ChunkIndex: hit.ChunkIndex,
			Content:    hit.Content,
			Score:      hit.Score,
		})
	}
	return QueryResponse{Hits: hits, Context: out.Context}
}
