package domain

import "errors"

var (
	ErrInvalidEmail = errors.New("invalid email address")
	ErrWeakPassword = errors.New("password does not meet the security policy")
)
