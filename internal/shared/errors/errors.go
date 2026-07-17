package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type Code string

const (
	CodeValidation   Code = "VALIDATION_ERROR"
	CodeNotFound     Code = "NOT_FOUND"
	CodeConflict     Code = "CONFLICT"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeForbidden    Code = "FORBIDDEN"
	CodeInternal     Code = "INTERNAL_ERROR"
	CodeUnavailable  Code = "SERVICE_UNAVAILABLE"
	CodeRateLimited  Code = "RATE_LIMITED"
	CodeExternalAPI  Code = "EXTERNAL_API_ERROR"
	CodeBadRequest   Code = "BAD_REQUEST"

	CodeAuthInvalidCredentials  Code = "AUTH_INVALID_CREDENTIALS" //nolint:gosec
	CodeAuthTokenExpired        Code = "AUTH_TOKEN_EXPIRED"       //nolint:gosec
	CodeAuthInvalidToken        Code = "AUTH_INVALID_TOKEN"       //nolint:gosec
	CodeAuthSessionExpired      Code = "AUTH_SESSION_EXPIRED"
	CodeAuthPermissionDenied    Code = "AUTH_PERMISSION_DENIED"
	CodeAuthAccountDisabled     Code = "AUTH_ACCOUNT_DISABLED"
	CodeAuthAccountSuspended    Code = "AUTH_ACCOUNT_SUSPENDED"
	CodeAuthRefreshRequired     Code = "AUTH_REFRESH_REQUIRED"
	CodeAuthInvalidRefreshToken Code = "AUTH_INVALID_REFRESH_TOKEN" //nolint:gosec
)

var httpStatusByCode = map[Code]int{
	CodeValidation:   http.StatusUnprocessableEntity,
	CodeNotFound:     http.StatusNotFound,
	CodeConflict:     http.StatusConflict,
	CodeUnauthorized: http.StatusUnauthorized,
	CodeForbidden:    http.StatusForbidden,
	CodeInternal:     http.StatusInternalServerError,
	CodeUnavailable:  http.StatusServiceUnavailable,
	CodeRateLimited:  http.StatusTooManyRequests,
	CodeExternalAPI:  http.StatusBadGateway,
	CodeBadRequest:   http.StatusBadRequest,

	CodeAuthInvalidCredentials:  http.StatusUnauthorized,
	CodeAuthTokenExpired:        http.StatusUnauthorized,
	CodeAuthInvalidToken:        http.StatusUnauthorized,
	CodeAuthSessionExpired:      http.StatusUnauthorized,
	CodeAuthPermissionDenied:    http.StatusForbidden,
	CodeAuthAccountDisabled:     http.StatusForbidden,
	CodeAuthAccountSuspended:    http.StatusForbidden,
	CodeAuthRefreshRequired:     http.StatusUnauthorized,
	CodeAuthInvalidRefreshToken: http.StatusUnauthorized,
}

type Error struct {
	Code       Code
	Message    string
	HTTPStatus int
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(code Code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusByCode[code],
	}
}

func Wrap(err error, code Code, message string) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatusByCode[code],
		Err:        err,
	}
}

func NewValidation(message string) *Error {
	return New(CodeValidation, message)
}

func NewNotFound(message string) *Error {
	return New(CodeNotFound, message)
}

func NewConflict(message string) *Error {
	return New(CodeConflict, message)
}

func NewUnauthorized(message string) *Error {
	return New(CodeUnauthorized, message)
}

func NewForbidden(message string) *Error {
	return New(CodeForbidden, message)
}

func NewInternal(err error) *Error {
	return Wrap(err, CodeInternal, "internal server error")
}

func As(err error) (*Error, bool) {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
