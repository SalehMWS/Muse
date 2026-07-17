package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/shared/authcontext"
	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
	"github.com/SalehMWS/Muse/internal/shared/response"
)

const bearerPrefix = "Bearer "

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

		authcontext.SetUser(c, claims.UserID, claims.SessionID)

		return c.Next()
	}
}

func CurrentUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	return authcontext.UserID(c)
}

func CurrentSessionID(c *fiber.Ctx) (uuid.UUID, bool) {
	return authcontext.SessionID(c)
}
