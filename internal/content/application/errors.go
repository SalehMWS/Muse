package application

import "errors"

var (
	ErrContentNotFound = errors.New("content not found")
	ErrInvalidCursor   = errors.New("invalid pagination cursor")
)
