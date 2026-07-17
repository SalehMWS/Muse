package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

const uniqueViolationCode = "23505"

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(db sqlc.DBTX) *UserRepository {
	return &UserRepository{queries: sqlc.New(db)}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:            user.ID,
		Email:         user.Email.String(),
		PasswordHash:  user.PasswordHash,
		DisplayName:   user.DisplayName,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.User{}, application.ErrEmailAlreadyExists
		}
		return domain.User{}, err
	}
	return toDomainUser(row)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, application.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(row)
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, application.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return toDomainUser(row)
}

func toDomainUser(row sqlc.User) (domain.User, error) {
	email, err := domain.NewEmail(row.Email)
	if err != nil {
		return domain.User{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		deletedAt = &t
	}

	return domain.User{
		ID:            row.ID,
		Email:         email,
		PasswordHash:  row.PasswordHash,
		DisplayName:   row.DisplayName,
		AvatarURL:     row.AvatarUrl,
		Status:        domain.Status(row.Status),
		EmailVerified: row.EmailVerified,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		DeletedAt:     deletedAt,
	}, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}
