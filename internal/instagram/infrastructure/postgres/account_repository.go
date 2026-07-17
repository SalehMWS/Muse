package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/instagram/domain"
	"github.com/SalehMWS/Muse/internal/shared/database/sqlc"
)

type AccountRepository struct {
	queries *sqlc.Queries
}

func NewAccountRepository(db sqlc.DBTX) *AccountRepository {
	return &AccountRepository{queries: sqlc.New(db)}
}

func (r *AccountRepository) Upsert(ctx context.Context, account domain.ConnectedAccount) (domain.ConnectedAccount, error) {
	row, err := r.queries.UpsertInstagramAccount(ctx, sqlc.UpsertInstagramAccountParams{
		ID:              account.ID,
		UserID:          account.UserID,
		InstagramUserID: account.InstagramUserID,
		Username:        account.Username,
		AccountType:     account.AccountType,
		AccessToken:     account.AccessToken,
		TokenExpiresAt:  pgtype.Timestamptz{Time: account.TokenExpiresAt, Valid: true},
		Scopes:          account.Scopes,
		Status:          string(account.Status),
	})
	if err != nil {
		return domain.ConnectedAccount{}, err
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) FindByIDForUser(ctx context.Context, id, userID uuid.UUID) (domain.ConnectedAccount, error) {
	row, err := r.queries.GetInstagramAccountByIDForUser(ctx, sqlc.GetInstagramAccountByIDForUserParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ConnectedAccount{}, application.ErrAccountNotFound
		}
		return domain.ConnectedAccount{}, err
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ConnectedAccount, error) {
	rows, err := r.queries.ListInstagramAccountsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	accounts := make([]domain.ConnectedAccount, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, toDomainAccount(row))
	}
	return accounts, nil
}

func (r *AccountRepository) UpdateToken(ctx context.Context, id uuid.UUID, accessToken string, expiresAt time.Time, status domain.AccountStatus) (domain.ConnectedAccount, error) {
	row, err := r.queries.UpdateInstagramAccountToken(ctx, sqlc.UpdateInstagramAccountTokenParams{
		ID:             id,
		AccessToken:    accessToken,
		TokenExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
		Status:         string(status),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ConnectedAccount{}, application.ErrAccountNotFound
		}
		return domain.ConnectedAccount{}, err
	}
	return toDomainAccount(row), nil
}

func (r *AccountRepository) DeleteForUser(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.DeleteInstagramAccountForUser(ctx, sqlc.DeleteInstagramAccountForUserParams{
		ID:     id,
		UserID: userID,
	})
}

func toDomainAccount(row sqlc.InstagramAccount) domain.ConnectedAccount {
	account := domain.ConnectedAccount{
		ID:              row.ID,
		UserID:          row.UserID,
		InstagramUserID: row.InstagramUserID,
		Username:        row.Username,
		AccountType:     row.AccountType,
		AccessToken:     row.AccessToken,
		TokenExpiresAt:  row.TokenExpiresAt.Time,
		Scopes:          row.Scopes,
		Status:          domain.AccountStatus(row.Status),
		ConnectedAt:     row.ConnectedAt.Time,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
	if row.LastRefreshedAt.Valid {
		refreshed := row.LastRefreshedAt.Time
		account.LastRefreshedAt = &refreshed
	}
	return account
}
