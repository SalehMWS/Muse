package application

import "errors"

var (
	ErrContentNotFound    = errors.New("content not found")
	ErrInvalidCursor      = errors.New("invalid pagination cursor")
	ErrCaptionUnavailable = errors.New("ai caption generation is unavailable")
	ErrMediaNotFound      = errors.New("media not found")
)
