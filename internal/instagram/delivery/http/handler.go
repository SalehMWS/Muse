package http

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	auditapp "github.com/SalehMWS/Muse/internal/audit/application"
	auditdomain "github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	connect    *application.ConnectUseCase
	callback   *application.CallbackUseCase
	list       *application.ListUseCase
	refresh    *application.RefreshUseCase
	disconnect *application.DisconnectUseCase
	audit      *auditapp.Recorder
}

func NewHandler(
	connect *application.ConnectUseCase,
	callback *application.CallbackUseCase,
	list *application.ListUseCase,
	refresh *application.RefreshUseCase,
	disconnect *application.DisconnectUseCase,
	audit *auditapp.Recorder,
) *Handler {
	return &Handler{
		connect:    connect,
		callback:   callback,
		list:       list,
		refresh:    refresh,
		disconnect: disconnect,
		audit:      audit,
	}
}

func (h *Handler) recordAudit(c *fiber.Ctx, action auditdomain.Action, userID *uuid.UUID, accountID string) {
	h.audit.Record(c.UserContext(), auditapp.Entry{
		UserID:       userID,
		Action:       action,
		Result:       auditdomain.ResultSuccess,
		ResourceType: "instagram_account",
		ResourceID:   accountID,
		IPAddress:    c.IP(),
		UserAgent:    c.Get(fiber.HeaderUserAgent),
	})
}

func (h *Handler) Connect(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}

	out, err := h.connect.Execute(c.UserContext(), userID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, ConnectResponse{
		AuthorizationURL: out.AuthorizationURL,
		State:            out.State,
	})
}

func (h *Handler) Callback(c *fiber.Ctx) error {
	if oauthErr := c.Query("error"); oauthErr != "" {
		message := c.Query("error_description")
		if message == "" {
			message = oauthErr
		}
		return response.Fail(c, apperrors.New(apperrors.CodeBadRequest, message))
	}

	code := c.Query("code")
	state := c.Query("state")
	if strings.TrimSpace(code) == "" || strings.TrimSpace(state) == "" {
		return response.Fail(c, apperrors.NewValidation("code and state are required"))
	}

	account, err := h.callback.Execute(c.UserContext(), application.CallbackInput{Code: code, State: state})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	h.recordAudit(c, auditdomain.ActionInstagramConnected, &account.UserID, account.ID.String())

	return response.Created(c, newAccountResponse(account, time.Now()))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}

	accounts, err := h.list.Execute(c.UserContext(), userID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	now := time.Now()
	items := make([]AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		items = append(items, newAccountResponse(account, now))
	}

	return response.OK(c, items)
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}

	accountID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid account id"))
	}

	account, err := h.refresh.Execute(c.UserContext(), userID, accountID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newAccountResponse(account, time.Now()))
}

func (h *Handler) Disconnect(c *fiber.Ctx) error {
	userID, ok := authcontext.UserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}

	accountID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid account id"))
	}

	if err := h.disconnect.Execute(c.UserContext(), userID, accountID); err != nil {
		return response.Fail(c, mapError(err))
	}

	h.recordAudit(c, auditdomain.ActionInstagramDisconnected, &userID, accountID.String())

	return response.NoContent(c)
}

func mapError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidState):
		return apperrors.New(apperrors.CodeBadRequest, err.Error())
	case errors.Is(err, application.ErrAccountNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrAccountNotPublishable):
		return apperrors.NewValidation(err.Error())
	case errors.Is(err, application.ErrInstagramAPI), errors.Is(err, application.ErrTokenExchange):
		return apperrors.New(apperrors.CodeExternalAPI, err.Error())
	default:
		if _, ok := apperrors.As(err); ok {
			return err
		}
		return apperrors.NewInternal(err)
	}
}
