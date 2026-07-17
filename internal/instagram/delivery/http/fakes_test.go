package http_test

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/instagram/domain"
)

type fakeAccountRepository struct {
	byID map[uuid.UUID]domain.ConnectedAccount
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{byID: map[uuid.UUID]domain.ConnectedAccount{}}
}

func (f *fakeAccountRepository) put(account domain.ConnectedAccount) {
	f.byID[account.ID] = account
}

func (f *fakeAccountRepository) Upsert(_ context.Context, account domain.ConnectedAccount) (domain.ConnectedAccount, error) {
	f.byID[account.ID] = account
	return account, nil
}

func (f *fakeAccountRepository) FindByIDForUser(_ context.Context, id, userID uuid.UUID) (domain.ConnectedAccount, error) {
	account, ok := f.byID[id]
	if !ok || account.UserID != userID {
		return domain.ConnectedAccount{}, application.ErrAccountNotFound
	}
	return account, nil
}

func (f *fakeAccountRepository) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.ConnectedAccount, error) {
	accounts := make([]domain.ConnectedAccount, 0)
	for _, account := range f.byID {
		if account.UserID == userID {
			accounts = append(accounts, account)
		}
	}
	return accounts, nil
}

func (f *fakeAccountRepository) UpdateToken(_ context.Context, id uuid.UUID, accessToken string, expiresAt time.Time, status domain.AccountStatus) (domain.ConnectedAccount, error) {
	account, ok := f.byID[id]
	if !ok {
		return domain.ConnectedAccount{}, application.ErrAccountNotFound
	}
	account.AccessToken = accessToken
	account.TokenExpiresAt = expiresAt
	account.Status = status
	f.byID[id] = account
	return account, nil
}

func (f *fakeAccountRepository) DeleteForUser(_ context.Context, id, userID uuid.UUID) error {
	if account, ok := f.byID[id]; ok && account.UserID == userID {
		delete(f.byID, id)
	}
	return nil
}

type fakeOAuthClient struct {
	exchangeToken  application.Token
	profile        application.Profile
	refreshedToken application.Token
}

func (f *fakeOAuthClient) AuthorizationURL(state string) string {
	return "https://www.instagram.com/oauth/authorize?state=" + state
}

func (f *fakeOAuthClient) ExchangeCode(context.Context, string) (application.Token, error) {
	return f.exchangeToken, nil
}

func (f *fakeOAuthClient) FetchProfile(context.Context, string) (application.Profile, error) {
	return f.profile, nil
}

func (f *fakeOAuthClient) RefreshToken(context.Context, string) (application.Token, error) {
	return f.refreshedToken, nil
}

type fakeTokenCipher struct{}

func (fakeTokenCipher) Encrypt(plaintext string) (string, error) {
	return "enc:" + plaintext, nil
}

func (fakeTokenCipher) Decrypt(ciphertext string) (string, error) {
	return strings.TrimPrefix(ciphertext, "enc:"), nil
}

type fakeStateSigner struct {
	userID    uuid.UUID
	verifyErr error
}

func (f fakeStateSigner) Sign(userID uuid.UUID) (string, error) {
	return "state:" + userID.String(), nil
}

func (f fakeStateSigner) Verify(string) (uuid.UUID, error) {
	if f.verifyErr != nil {
		return uuid.Nil, f.verifyErr
	}
	return f.userID, nil
}
