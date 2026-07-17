package domain

import "errors"

var (
	ErrTitleTooLong      = errors.New("title exceeds maximum length")
	ErrCaptionTooLong    = errors.New("caption exceeds maximum length")
	ErrInvalidType       = errors.New("invalid content type")
	ErrInvalidVisibility = errors.New("invalid visibility")
	ErrInvalidStatus     = errors.New("invalid content status")
	ErrTooManyTags       = errors.New("too many tags")
	ErrTagTooLong        = errors.New("tag exceeds maximum length")
)
