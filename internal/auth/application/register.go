package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterOutput struct {
	User domain.User
}

type RegisterUseCase struct {
	users  UserRepository
	hasher PasswordHasher
}

func NewRegisterUseCase(users UserRepository, hasher PasswordHasher) *RegisterUseCase {
	return &RegisterUseCase{users: users, hasher: hasher}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (RegisterOutput, error) {
	email, err := domain.NewEmail(input.Email)
	if err != nil {
		return RegisterOutput{}, err
	}

	if err := domain.ValidatePassword(input.Password); err != nil {
		return RegisterOutput{}, err
	}

	_, err = uc.users.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return RegisterOutput{}, ErrEmailAlreadyExists
	case errors.Is(err, ErrUserNotFound):
	default:
		return RegisterOutput{}, fmt.Errorf("register: check existing email: %w", err)
	}

	hash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: hash password: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("register: generate id: %w", err)
	}

	user := domain.User{
		ID:            id,
		Email:         email,
		PasswordHash:  hash,
		DisplayName:   strings.TrimSpace(input.DisplayName),
		Status:        domain.StatusActive,
		EmailVerified: false,
	}

	created, err := uc.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			return RegisterOutput{}, ErrEmailAlreadyExists
		}
		return RegisterOutput{}, fmt.Errorf("register: create user: %w", err)
	}

	return RegisterOutput{User: created}, nil
}
