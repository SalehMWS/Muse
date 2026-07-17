package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type RefreshUseCase struct {
	oauth  OAuthClient
	cipher TokenCipher
	repo   AccountRepository
}

func NewRefreshUseCase(oauth OAuthClient, cipher TokenCipher, repo AccountRepository) *RefreshUseCase {
	return &RefreshUseCase{oauth: oauth, cipher: cipher, repo: repo}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, userID, accountID uuid.UUID) (domain.ConnectedAccount, error) {
	account, err := uc.repo.FindByIDForUser(ctx, accountID, userID)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	current, err := uc.cipher.Decrypt(account.AccessToken)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	token, err := uc.oauth.RefreshToken(ctx, current)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	encrypted, err := uc.cipher.Encrypt(token.AccessToken)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	updated, err := uc.repo.UpdateToken(ctx, account.ID, encrypted, time.Now().Add(token.ExpiresIn), domain.AccountStatusActive)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	updated.AccessToken = ""
	return updated, nil
}
