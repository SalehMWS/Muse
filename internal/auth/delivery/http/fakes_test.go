package http_test

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type fakeUserRepository struct {
	byEmail map[string]domain.User
	byID    map[uuid.UUID]domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{byEmail: map[string]domain.User{}, byID: map[uuid.UUID]domain.User{}}
}

func (f *fakeUserRepository) put(user domain.User) {
	f.byEmail[user.Email.String()] = user
	f.byID[user.ID] = user
}

func (f *fakeUserRepository) Create(_ context.Context, user domain.User) (domain.User, error) {
	if _, exists := f.byEmail[user.Email.String()]; exists {
		return domain.User{}, application.ErrEmailAlreadyExists
	}
	f.put(user)
	return user, nil
}

func (f *fakeUserRepository) FindByEmail(_ context.Context, email domain.Email) (domain.User, error) {
	if u, ok := f.byEmail[email.String()]; ok {
		return u, nil
	}
	return domain.User{}, application.ErrUserNotFound
}

func (f *fakeUserRepository) FindByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return domain.User{}, application.ErrUserNotFound
}

type fakeSessionRepository struct {
	byHash map[string]domain.Session
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{byHash: map[string]domain.Session{}}
}

func (f *fakeSessionRepository) Create(_ context.Context, session domain.Session) (domain.Session, error) {
	f.byHash[session.RefreshTokenHash] = session
	return session, nil
}

func (f *fakeSessionRepository) FindByRefreshTokenHash(_ context.Context, refreshTokenHash string) (domain.Session, error) {
	if s, ok := f.byHash[refreshTokenHash]; ok {
		return s, nil
	}
	return domain.Session{}, application.ErrSessionNotFound
}

func (f *fakeSessionRepository) Rotate(_ context.Context, sessionID uuid.UUID, newRefreshTokenHash string, newExpiresAt time.Time) (domain.Session, error) {
	for hash, s := range f.byHash {
		if s.ID != sessionID {
			continue
		}
		delete(f.byHash, hash)
		s.RefreshTokenHash = newRefreshTokenHash
		s.ExpiresAt = newExpiresAt
		f.byHash[newRefreshTokenHash] = s
		return s, nil
	}
	return domain.Session{}, application.ErrSessionNotFound
}

func (f *fakeSessionRepository) DeleteByRefreshTokenHash(_ context.Context, refreshTokenHash string) error {
	delete(f.byHash, refreshTokenHash)
	return nil
}

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (fakePasswordHasher) Verify(hash, plain string) (bool, error) {
	return hash == "hashed:"+plain, nil
}

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(_ context.Context, userID, sessionID uuid.UUID) (application.Token, error) {
	return application.Token{
		Value:     userID.String() + ":" + sessionID.String(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil
}

func (fakeTokenIssuer) Verify(_ context.Context, tokenString string) (application.Claims, error) {
	parts := strings.SplitN(tokenString, ":", 2)
	if len(parts) != 2 {
		return application.Claims{}, errors.New("invalid token")
	}

	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return application.Claims{}, err
	}

	sessionID, err := uuid.Parse(parts[1])
	if err != nil {
		return application.Claims{}, err
	}

	return application.Claims{UserID: userID, SessionID: sessionID}, nil
}
