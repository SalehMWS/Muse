package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/knowledge/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	ingest *application.IngestUseCase
	query  *application.QueryUseCase
	list   *application.ListDocumentsUseCase
	delete *application.DeleteDocumentUseCase
}

func NewHandler(ingest *application.IngestUseCase, query *application.QueryUseCase, list *application.ListDocumentsUseCase, del *application.DeleteDocumentUseCase) *Handler {
	return &Handler{ingest: ingest, query: query, list: list, delete: del}
}

func (h *Handler) Ingest(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}

	var req IngestRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	document, err := h.ingest.Execute(c.UserContext(), application.IngestInput{
		UserID:  userID,
		Title:   req.Title,
		Source:  req.Source,
		Content: req.Content,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newDocumentResponse(document))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}

	documents, err := h.list.Execute(c.UserContext(), userID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	out := make([]DocumentResponse, 0, len(documents))
	for _, document := range documents {
		out = append(out, newDocumentResponse(document))
	}
	return response.OK(c, out)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}
	documentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid document id"))
	}

	if err := h.delete.Execute(c.UserContext(), userID, documentID); err != nil {
		return response.Fail(c, mapError(err))
	}
	return response.NoContent(c)
}

func (h *Handler) Query(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, missingUser())
	}

	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	out, err := h.query.Execute(c.UserContext(), application.QueryInput{
		UserID: userID,
		Query:  req.Query,
		TopK:   req.TopK,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newQueryResponse(out))
}

func missingUser() error {
	return apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user")
}

func mapError(err error) error {
	switch {
	case errors.Is(err, application.ErrDocumentNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrEmptyContent), errors.Is(err, application.ErrEmptyQuery):
		return apperrors.NewValidation(err.Error())
	case errors.Is(err, application.ErrEmbedding), errors.Is(err, application.ErrVectorStore):
		return apperrors.New(apperrors.CodeExternalAPI, err.Error())
	default:
		if _, ok := apperrors.As(err); ok {
			return err
		}
		return apperrors.NewInternal(err)
	}
}

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	group := router.Group("/knowledge")
	group.Post("/documents", requireAuth, h.Ingest)
	group.Get("/documents", requireAuth, h.List)
	group.Delete("/documents/:id", requireAuth, h.Delete)
	group.Post("/query", requireAuth, h.Query)
}
