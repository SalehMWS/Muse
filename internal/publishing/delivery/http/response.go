package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/publishing/domain"
)

type PublicationResponse struct {
	ID                 string  `json:"id"`
	ContentID          string  `json:"content_id"`
	InstagramAccountID string  `json:"instagram_account_id"`
	Platform           string  `json:"platform"`
	Status             string  `json:"status"`
	PlatformPostID     *string `json:"platform_post_id,omitempty"`
	Permalink          *string `json:"permalink,omitempty"`
	PublishedAt        *string `json:"published_at,omitempty"`
	CreatedAt          string  `json:"created_at"`
}

func newPublicationResponse(publication domain.Publication) PublicationResponse {
	resp := PublicationResponse{
		ID:                 publication.ID.String(),
		ContentID:          publication.ContentID.String(),
		InstagramAccountID: publication.InstagramAccountID.String(),
		Platform:           publication.Platform,
		Status:             string(publication.Status),
		PlatformPostID:     publication.PlatformPostID,
		Permalink:          publication.Permalink,
		CreatedAt:          publication.CreatedAt.Format(time.RFC3339),
	}
	if publication.PublishedAt != nil {
		published := publication.PublishedAt.Format(time.RFC3339)
		resp.PublishedAt = &published
	}
	return resp
}
