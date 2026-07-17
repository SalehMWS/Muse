package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

const (
	localsUserID    = "auth_user_id"
	localsSessionID = "auth_session_id"
	bearerPrefix    = "Bearer "
)

func RequireAuth(issuer application.TokenIssuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if !strings.HasPrefix(header, bearerPrefix) {
			return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "missing bearer token"))
		}

		tokenString := strings.TrimPrefix(header, bearerPrefix)

		claims, err := issuer.Verify(c.UserContext(), tokenString)
		if err != nil {
			return response.Fail(c, apperrors.New(apperrors.CodeAuthInvalidToken, "invalid or expired token"))
		}

		c.Locals(localsUserID, claims.UserID)
		c.Locals(localsSessionID, claims.SessionID)

		return c.Next()
	}
}

func CurrentUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsUserID).(uuid.UUID)
	return id, ok
}

func CurrentSessionID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsSessionID).(uuid.UUID)
	return id, ok
}
