package response

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	apperrors "github.com/SalehMWS/Muse/internal/shared/errors"
)

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: data})
}

func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

func Fail(c *fiber.Ctx, err error) error {
	status, code, message := describe(err)
	return c.Status(status).JSON(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func StatusFromError(err error) int {
	status, _, _ := describe(err)
	return status
}

func describe(err error) (status int, code string, message string) {
	if appErr, ok := apperrors.As(err); ok {
		return appErr.HTTPStatus, string(appErr.Code), appErr.Message
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code, fiberErrorCode(fiberErr.Code), fiberErr.Message
	}

	internal := apperrors.NewInternal(err)
	return internal.HTTPStatus, string(internal.Code), internal.Message
}

func fiberErrorCode(status int) string {
	switch status {
	case fiber.StatusNotFound:
		return string(apperrors.CodeNotFound)
	case fiber.StatusMethodNotAllowed:
		return string(apperrors.CodeBadRequest)
	default:
		return string(apperrors.CodeInternal)
	}
}
