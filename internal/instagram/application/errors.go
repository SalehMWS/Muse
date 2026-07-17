package application

import "errors"

var (
	ErrInvalidState          = errors.New("invalid or expired oauth state")
	ErrAccountNotFound       = errors.New("instagram account not found")
	ErrAccountNotPublishable = errors.New("instagram account must be a business or creator account")
	ErrTokenExchange         = errors.New("failed to exchange instagram authorization code")
	ErrInstagramAPI          = errors.New("instagram api request failed")
)
