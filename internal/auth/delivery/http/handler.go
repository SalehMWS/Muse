package http

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	auditapp "github.com/SalehMWS/Muse/internal/audit/application"
	auditdomain "github.com/SalehMWS/Muse/internal/audit/domain"
	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

type Handler struct {
	register       *application.RegisterUseCase
	login          *application.LoginUseCase
	refresh        *application.RefreshUseCase
	logout         *application.LogoutUseCase
	getCurrentUser *application.GetCurrentUserUseCase
	audit          *auditapp.Recorder
}

func NewHandler(
	register *application.RegisterUseCase,
	login *application.LoginUseCase,
	refresh *application.RefreshUseCase,
	logout *application.LogoutUseCase,
	getCurrentUser *application.GetCurrentUserUseCase,
	audit *auditapp.Recorder,
) *Handler {
	return &Handler{
		register:       register,
		login:          login,
		refresh:        refresh,
		logout:         logout,
		getCurrentUser: getCurrentUser,
		audit:          audit,
	}
}

func (h *Handler) recordAudit(c *fiber.Ctx, action auditdomain.Action, result auditdomain.Result, userID *uuid.UUID, metadata map[string]string) {
	h.audit.Record(c.UserContext(), auditapp.Entry{
		UserID:    userID,
		Action:    action,
		Result:    result,
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
		Metadata:  metadata,
	})
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.DisplayName) == "" {
		return response.Fail(c, apperrors.NewValidation("email, password, and display_name are required"))
	}

	out, err := h.register.Execute(c.UserContext(), application.RegisterInput{
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	h.recordAudit(c, auditdomain.ActionUserRegistered, auditdomain.ResultSuccess, &out.User.ID, nil)

	return response.Created(c, newUserResponse(out.User))
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return response.Fail(c, apperrors.NewValidation("email and password are required"))
	}

	out, err := h.login.Execute(c.UserContext(), application.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		Device:    c.Get("X-Device-Name"),
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		h.recordAudit(c, auditdomain.ActionUserLoginFailed, auditdomain.ResultFailure, nil, map[string]string{
			"email": strings.TrimSpace(req.Email),
		})
		return response.Fail(c, mapError(err))
	}

	h.recordAudit(c, auditdomain.ActionUserLoggedIn, auditdomain.ResultSuccess, &out.User.ID, nil)

	return response.OK(c, AuthResponse{
		User:         newUserResponse(out.User),
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    expiresInSeconds(out.AccessTokenExpiresAt),
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		return response.Fail(c, apperrors.NewValidation("refresh_token is required"))
	}

	out, err := h.refresh.Execute(c.UserContext(), application.RefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, TokenResponse{
		AccessToken:  out.AccessToken,
		RefreshToken: out.RefreshToken,
		ExpiresIn:    expiresInSeconds(out.AccessTokenExpiresAt),
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, apperrors.NewValidation("invalid request body"))
	}

	if strings.TrimSpace(req.RefreshToken) == "" {
		return response.Fail(c, apperrors.NewValidation("refresh_token is required"))
	}

	if err := h.logout.Execute(c.UserContext(), application.LogoutInput{RefreshToken: req.RefreshToken}); err != nil {
		return response.Fail(c, mapError(err))
	}

	var userID *uuid.UUID
	if id, ok := CurrentUserID(c); ok {
		userID = &id
	}
	h.recordAudit(c, auditdomain.ActionUserLoggedOut, auditdomain.ResultSuccess, userID, nil)

	return response.NoContent(c)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, ok := CurrentUserID(c)
	if !ok {
		return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing authenticated user"))
	}

	user, err := h.getCurrentUser.Execute(c.UserContext(), userID)
	if err != nil {
		return response.Fail(c, mapError(err))
	}

	return response.OK(c, newUserResponse(user))
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrWeakPassword):
		return apperrors.NewValidation(err.Error())
	case errors.Is(err, application.ErrEmailAlreadyExists):
		return apperrors.NewConflict(err.Error())
	case errors.Is(err, application.ErrInvalidCredentials):
		return apperrors.New(apperrors.CodeAuthInvalidCredentials, err.Error())
	case errors.Is(err, application.ErrAccountSuspended):
		return apperrors.New(apperrors.CodeAuthAccountSuspended, err.Error())
	case errors.Is(err, application.ErrAccountDisabled):
		return apperrors.New(apperrors.CodeAuthAccountDisabled, err.Error())
	case errors.Is(err, application.ErrUserNotFound):
		return apperrors.NewNotFound(err.Error())
	case errors.Is(err, application.ErrSessionNotFound), errors.Is(err, application.ErrRefreshTokenExpired):
		return apperrors.New(apperrors.CodeAuthInvalidRefreshToken, err.Error())
	default:
		return apperrors.NewInternal(err)
	}
}
