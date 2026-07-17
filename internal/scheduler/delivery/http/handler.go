package http

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/scheduler/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	create *application.CreateScheduleUseCase
	list   *application.ListSchedulesUseCase
	cancel *application.CancelScheduleUseCase
}

func NewHandler(create *application.CreateScheduleUseCase, list *application.ListSchedulesUseCase, cancel *application.CancelScheduleUseCase) *Handler {
	return &Handler{create: create, list: list, cancel: cancel}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	userID, contentID, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	accountID, err := uuid.Parse(strings.TrimSpace(req.InstagramAccountID))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("valid instagram_account_id is required"))
	}

	in := application.CreateScheduleInput{
		UserID:             userID,
		ContentID:          contentID,
		InstagramAccountID: accountID,
		CronExpression:     req.CronExpression,
		Timezone:           req.Timezone,
		MediaType:          req.MediaType,
		MaxRetries:         req.MaxRetries,
	}

	if scheduledFor := strings.TrimSpace(req.ScheduledFor); scheduledFor != "" {
		parsed, err := time.Parse(time.RFC3339, scheduledFor)
		if err != nil {
			return response.Fail(c, apperrors.NewValidation("scheduled_for must be an RFC3339 timestamp"))
		}
		in.ScheduledFor = &parsed
	}

	schedule, err := h.create.Execute(c.UserContext(), in)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.Created(c, newScheduleResponse(schedule))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, contentID, err := identify(c)
	if err != nil {
		return response.Fail(c, err)
	}

	schedules, err := h.list.Execute(c.UserContext(), userID, contentID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	out := make([]ScheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		out = append(out, newScheduleResponse(schedule))
	}
	return response.OK(c, out)
}

func (h *Handler) Cancel(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}
	scheduleID, err := uuid.Parse(c.Params("scheduleId"))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid schedule id"))
	}

	if err := h.cancel.Execute(c.UserContext(), userID, scheduleID); err != nil {
		return response.Fail(c, mapError(err))
	}
	return response.NoContent(c)
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
	case errors.Is(err, application.ErrScheduleNotFound), errors.Is(err, application.ErrContentNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrInvalidCron),
		errors.Is(err, application.ErrInvalidTimezone),
		errors.Is(err, application.ErrScheduleInPast),
		errors.Is(err, application.ErrScheduleTimeRequired):
		return apperrors.NewValidation(err.Error())
	default:
		if _, ok := apperrors.As(err); ok {
			return err
		}
		return apperrors.NewInternal(err)
	}
}
