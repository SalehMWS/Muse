package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/content/domain"
)

type ContentResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Caption     string   `json:"caption"`
	Status      string   `json:"status"`
	Language    string   `json:"language"`
	ContentType string   `json:"content_type"`
	Visibility  string   `json:"visibility"`
	Tags        []string `json:"tags"`
	PublishedAt *string  `json:"published_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func newContentResponse(content domain.Content) ContentResponse {
	tags := content.Tags
	if tags == nil {
		tags = []string{}
	}

	resp := ContentResponse{
		ID:          content.ID.String(),
		Title:       content.Title,
		Caption:     content.Caption,
		Status:      string(content.Status),
		Language:    content.Language,
		ContentType: string(content.ContentType),
		Visibility:  string(content.Visibility),
		Tags:        tags,
		CreatedAt:   content.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   content.UpdatedAt.Format(time.RFC3339),
	}
	if content.PublishedAt != nil {
		published := content.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &published
	}
	return resp
}

type ContentListResponse struct {
	Items      []ContentResponse `json:"items"`
	NextCursor string            `json:"next_cursor,omitempty"`
}
