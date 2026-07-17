package application

import "errors"

var (
	ErrAccountNotFound  = errors.New("connected instagram account not found")
	ErrContentNotFound  = errors.New("content not found")
	ErrNoMedia          = errors.New("content has no media to publish")
	ErrInvalidMediaType = errors.New("invalid publish media type")
	ErrPublishFailed    = errors.New("instagram publish failed")
)
