package http

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/content/application"
	"github.com/SalehMWS/Muse/internal/content/domain"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	create          *application.CreateUseCase
	get             *application.GetUseCase
	update          *application.UpdateUseCase
	archive         *application.ArchiveUseCase
	duplicate       *application.DuplicateUseCase
	list            *application.ListUseCase
	generateCaption *application.GenerateCaptionUseCase
}

func NewHandler(
	create *application.CreateUseCase,
	get *application.GetUseCase,
	update *application.UpdateUseCase,
	archive *application.ArchiveUseCase,
	duplicate *application.DuplicateUseCase,
	list *application.ListUseCase,
	generateCaption *application.GenerateCaptionUseCase,
) *Handler {
	return &Handler{
		create:          create,
		get:             get,
		update:          update,
		archive:         archive,
		duplicate:       duplicate,
		list:            list,
		generateCaption: generateCaption,
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}

	var req CreateContentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	content, err := h.create.Execute(c.UserContext(), userID, domain.NewContentInput{
		Title:       req.Title,
		Caption:     req.Caption,
		Language:    req.Language,
		ContentType: req.ContentType,
		Visibility:  req.Visibility,
		Tags:        req.Tags,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newContentResponse(content))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}

	in := application.ListInput{
		UserID:      userID,
		Status:      optionalQuery(c, "status"),
		Language:    optionalQuery(c, "language"),
		ContentType: optionalQuery(c, "type"),
		Tag:         optionalQuery(c, "tag"),
		Cursor:      c.Query("cursor"),
	}

	if limitStr := strings.TrimSpace(c.Query("limit")); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return response.Fail(c, apperrors.NewValidation("invalid limit"))
		}
		in.Limit = limit
	}

	out, err := h.list.Execute(c.UserContext(), in)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	items := make([]ContentResponse, 0, len(out.Items))
	for _, content := range out.Items {
		items = append(items, newContentResponse(content))
	}

	return response.OK(c, ContentListResponse{Items: items, NextCursor: out.NextCursor})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	content, err := h.get.Execute(c.UserContext(), userID, id)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newContentResponse(content))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	var req UpdateContentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	content, err := h.update.Execute(c.UserContext(), userID, id, domain.UpdateContentInput{
		Title:       req.Title,
		Caption:     req.Caption,
		Status:      req.Status,
		Language:    req.Language,
		ContentType: req.ContentType,
		Visibility:  req.Visibility,
		Tags:        req.Tags,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newContentResponse(content))
}

func (h *Handler) Archive(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	content, err := h.archive.Execute(c.UserContext(), userID, id)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newContentResponse(content))
}

func (h *Handler) Duplicate(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	content, err := h.duplicate.Execute(c.UserContext(), userID, id)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newContentResponse(content))
}

func (h *Handler) GenerateCaption(c *fiber.Ctx) error {
	userID, id, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	var req GenerateCaptionRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.Fail(c, apperrors.NewValidation("invalid request body"))
		}
	}

	content, err := h.generateCaption.Execute(c.UserContext(), userID, id, req.Prompt)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newContentResponse(content))
}

func identify(c *fiber.Ctx) (uuid.UUID, uuid.UUID, error) {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return uuid.Nil, uuid.Nil, missingUser()
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperrors.NewValidation("invalid content id")
	}
	return userID, id, nil
}

func missingUser() error {
	return apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user")
}

func optionalQuery(c *fiber.Ctx, key string) *string {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil
	}
	return &value
}

func mapError(err error) error {
	switch {
	case errors.Is(err, application.ErrContentNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrCaptionUnavailable):
		return apperrors.New(apperrors.CodeExternalAPI, err.Error())
	case errors.Is(err, application.ErrInvalidCursor),
		errors.Is(err, domain.ErrTitleTooLong),
		errors.Is(err, domain.ErrCaptionTooLong),
		errors.Is(err, domain.ErrInvalidType),
		errors.Is(err, domain.ErrInvalidVisibility),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrTooManyTags),
		errors.Is(err, domain.ErrTagTooLong):
		return apperrors.NewValidation(err.Error())
	default:
		if _, ok := apperrors.As(err); ok {
			return err
		}
		return apperrors.NewInternal(err)
	}
}
