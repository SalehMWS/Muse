package http

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/SalehMWS/Muse/internal/audit/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	list *application.ListEventsUseCase
}

func NewHandler(list *application.ListEventsUseCase) *Handler {
	return &Handler{list: list}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.NewUnauthorized("authentication required"))
	}

	limit, err := parseLimit(c.Query("limit"))
	if err != nil {
		return response.Fail(c, err)
	}

	events, err := h.list.Execute(c.UserContext(), userID, limit)
	if err != nil {
		return response.Fail(c, apperrors.NewInternal(err))
	}

	return response.OK(c, newEventResponses(events))
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, apperrors.NewValidation("limit must be an integer")
	}
	if limit < 0 {
		return 0, apperrors.NewValidation("limit must not be negative")
	}
	return limit, nil
}

func RegisterRoutes(router fiber.Router, h *Handler, requireAuth fiber.Handler) {
	group := router.Group("/audit")
	group.Get("/events", requireAuth, h.List)
}
