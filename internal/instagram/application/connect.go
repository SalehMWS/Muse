package application

import (
	"context"

	"github.com/google/uuid"
)

type ConnectUseCase struct {
	oauth  OAuthClient
	signer StateSigner
}

func NewConnectUseCase(oauth OAuthClient, signer StateSigner) *ConnectUseCase {
	return &ConnectUseCase{oauth: oauth, signer: signer}
}

type ConnectOutput struct {
	AuthorizationURL string
	State            string
}

func (uc *ConnectUseCase) Execute(_ context.Context, userID uuid.UUID) (ConnectOutput, error) {
	state, err := uc.signer.Sign(userID)
	if err != nil {
		return ConnectOutput{}, err
	}

	return ConnectOutput{
		AuthorizationURL: uc.oauth.AuthorizationURL(state),
		State:            state,
	}, nil
}
