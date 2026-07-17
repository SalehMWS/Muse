package http

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/publishing/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	publish *application.PublishUseCase
	list    *application.ListPublicationsUseCase
}

func NewHandler(publish *application.PublishUseCase, list *application.ListPublicationsUseCase) *Handler {
	return &Handler{publish: publish, list: list}
}

func (h *Handler) Publish(c *fiber.Ctx) error {
	userID, contentID, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	var req PublishRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}
	if strings.TrimSpace(req.InstagramAccountID) == "" {
		return response.Fail(c, apperrors.NewValidation("instagram_account_id is required"))
	}
	accountID, err := uuid.Parse(strings.TrimSpace(req.InstagramAccountID))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid instagram_account_id"))
	}

	publication, err := h.publish.Execute(c.UserContext(), application.PublishInput{
		UserID:             userID,
		ContentID:          contentID,
		InstagramAccountID: accountID,
		MediaType:          req.MediaType,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newPublicationResponse(publication))
}

func (h *Handler) ListPublications(c *fiber.Ctx) error {
	userID, contentID, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	items, err := h.list.Execute(c.UserContext(), userID, contentID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	out := make([]PublicationResponse, 0, len(items))
	for _, publication := range items {
		out = append(out, newPublicationResponse(publication))
	}
	return response.OK(c, out)
}

func identify(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user")
	}
	contentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperrors.NewValidation("invalid content id")
	}
	return userID, contentID, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, application.ErrAccountNotFound), errors.Is(err, application.ErrContentNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrNoMedia), errors.Is(err, application.ErrInvalidMediaType):
		return apperrors.NewValidation(err.Error())
	case errors.Is(err, application.ErrPublishFailed):
		return apperrors.New(apperrors.CodeExternalAPI, err.Error())
	default:
		if _, ok := apperrors.As(err); ok {
			return err
		}
		return apperrors.NewInternal(err)
	}
}
