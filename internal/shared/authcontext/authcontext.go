package authcontext

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	localsUserID    = "auth_user_id"
	localsSessionID = "auth_session_id"
)

func SetUser(c *fiber.Ctx, userID, sessionID uuid.UUID) {
	c.Locals(localsUserID, userID)
	c.Locals(localsSessionID, sessionID)
}

func UserID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsUserID).(uuid.UUID)
	return id, ok
}

func SessionID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsSessionID).(uuid.UUID)
	return id, ok
}
