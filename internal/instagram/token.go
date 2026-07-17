package instagram

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/postgres"
	"github.com/SalehMWS/Muse/internal/instagram/infrastructure/security"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

type ConnectedAccount struct {
	ID              uuid.UUID
	InstagramUserID string
	AccessToken     string
}

type TokenService struct {
	repo   *postgres.AccountRepository
	cipher *security.AESTokenCipher
}

func NewTokenService(pool *pgxpool.Pool, cfg config.Instagram) (*TokenService, error) {
	cipher, err := security.NewAESTokenCipher(cfg.TokenEncryptionKey)
	if err != nil {
		return nil, err
	}
	return &TokenService{
		repo:   postgres.NewAccountRepository(pool),
		cipher: cipher,
	}, nil
}

func (s *TokenService) Resolve(ctx context.Context, userID, accountID uuid.UUID) (ConnectedAccount, error) {
	account, err := s.repo.FindByIDForUser(ctx, accountID, userID)
	if err != nil {
		return ConnectedAccount{}, err
	}

	token, err := s.cipher.Decrypt(account.AccessToken)
	if err != nil {
		return ConnectedAccount{}, err
	}

	return ConnectedAccount{
		ID:              account.ID,
		InstagramUserID: account.InstagramUserID,
		AccessToken:     token,
	}, nil
}
