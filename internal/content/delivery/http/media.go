package http

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/domain"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type MediaRequest struct {
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Position  int    `json:"position"`
}

type MediaResponse struct {
	ID        string `json:"id"`
	ContentID string `json:"content_id"`
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	Position  int    `json:"position"`
	CreatedAt string `json:"created_at"`
}

func newMediaResponse(media domain.Media) MediaResponse {
	return MediaResponse{
		ID:        media.ID.String(),
		ContentID: media.ContentID.String(),
		URL:       media.URL,
		MediaType: string(media.MediaType),
		Position:  media.Position,
		CreatedAt: media.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) AttachMedia(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	var req MediaRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}
	if strings.TrimSpace(req.URL) == "" {
		return response.Fail(c, apperrors.NewValidation("url is required"))
	}

	media, err := h.attachMedia.Execute(c.UserContext(), userID, id, req.URL, req.MediaType, req.Position)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newMediaResponse(media))
}

func (h *Handler) ListMedia(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	items, err := h.listMedia.Execute(c.UserContext(), userID, id)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	out := make([]MediaResponse, 0, len(items))
	for _, media := range items {
		out = append(out, newMediaResponse(media))
	}
	return response.OK(c, out)
}

func (h *Handler) DeleteMedia(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	mediaID, err := uuid.Parse(c.Params("mediaId"))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid media id"))
	}

	if err := h.deleteMedia.Execute(c.UserContext(), userID, id, mediaID); err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.NoContent(c)
}
