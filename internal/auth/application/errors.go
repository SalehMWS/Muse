package application

import "errors"

var (
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountSuspended    = errors.New("account is suspended")
	ErrAccountDisabled     = errors.New("account is disabled")
	ErrUserNotFound        = errors.New("user not found")
	ErrSessionNotFound     = errors.New("session not found")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)
