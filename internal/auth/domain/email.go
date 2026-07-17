package domain

import (
	"net/mail"
	"strings"
)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return Email{}, ErrInvalidEmail
	}

	return Email{value: strings.ToLower(addr.Address)}, nil
}

func (e Email) String() string {
	return e.value
}
