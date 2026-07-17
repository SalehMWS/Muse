package instagram

import (
	"context"
	"errors"

	"github.com/google/uuid"

	igmodule "github.com/SalehMWS/Muse/internal/instagram"
	igapp "github.com/SalehMWS/Muse/internal/instagram/application"
	"github.com/SalehMWS/Muse/internal/publishing/application"
)

type AccountReader struct {
	tokens *igmodule.TokenService
}

func NewAccountReader(tokens *igmodule.TokenService) *AccountReader {
	return &AccountReader{tokens: tokens}
}

func (r *AccountReader) AccountForUser(ctx context.Context, userID, accountID uuid.UUID) (application.Account, error) {
	account, err := r.tokens.Resolve(ctx, userID, accountID)
	if err != nil {
		if errors.Is(err, igapp.ErrAccountNotFound) {
			return application.Account{}, application.ErrAccountNotFound
		}
		return application.Account{}, err
	}

	return application.Account{
		ID:              account.ID,
		InstagramUserID: account.InstagramUserID,
		AccessToken:     account.AccessToken,
	}, nil
}
