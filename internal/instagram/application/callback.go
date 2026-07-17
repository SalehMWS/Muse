package application

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type CallbackUseCase struct {
	oauth  OAuthClient
	signer StateSigner
	cipher TokenCipher
	repo   AccountRepository
}

func NewCallbackUseCase(oauth OAuthClient, signer StateSigner, cipher TokenCipher, repo AccountRepository) *CallbackUseCase {
	return &CallbackUseCase{oauth: oauth, signer: signer, cipher: cipher, repo: repo}
}

type CallbackInput struct {
	Code  string
	State string
}

func (uc *CallbackUseCase) Execute(ctx context.Context, in CallbackInput) (domain.ConnectedAccount, error) {
	userID, err := uc.signer.Verify(in.State)
	if err != nil {
		return domain.ConnectedAccount{}, ErrInvalidState
	}

	token, err := uc.oauth.ExchangeCode(ctx, in.Code)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	profile, err := uc.oauth.FetchProfile(ctx, token.AccessToken)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	if !domain.IsPublishableAccountType(profile.AccountType) {
		return domain.ConnectedAccount{}, ErrAccountNotPublishable
	}

	encrypted, err := uc.cipher.Encrypt(token.AccessToken)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	accountType := profile.AccountType
	account := domain.ConnectedAccount{
		ID:              uuid.New(),
		UserID:          userID,
		InstagramUserID: profile.UserID,
		Username:        profile.Username,
		AccountType:     &accountType,
		AccessToken:     encrypted,
		TokenExpiresAt:  time.Now().Add(token.ExpiresIn),
		Status:          domain.AccountStatusActive,
	}

	saved, err := uc.repo.Upsert(ctx, account)
	if err != nil {
		return domain.ConnectedAccount{}, err
	}

	saved.AccessToken = ""
	return saved, nil
}
