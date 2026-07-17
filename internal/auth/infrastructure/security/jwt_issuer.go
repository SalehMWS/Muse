package security

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/application"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

type claims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

type JWTIssuer struct {
	cfg config.JWT
}

func NewJWTIssuer(cfg config.JWT) *JWTIssuer {
	return &JWTIssuer{cfg: cfg}
}

func (i *JWTIssuer) Issue(_ context.Context, userID, sessionID uuid.UUID) (application.Token, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(i.cfg.AccessTokenTTL)

	jti, err := uuid.NewV7()
	if err != nil {
		return application.Token{}, fmt.Errorf("jwt: generate jti: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    i.cfg.Issuer,
			Audience:  jwt.ClaimStrings{i.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        jti.String(),
		},
	})

	signed, err := token.SignedString([]byte(i.cfg.Secret))
	if err != nil {
		return application.Token{}, fmt.Errorf("jwt: sign token: %w", err)
	}

	return application.Token{Value: signed, ExpiresAt: expiresAt}, nil
}

func (i *JWTIssuer) Verify(_ context.Context, tokenString string) (application.Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method %v", token.Header["alg"])
		}
		return []byte(i.cfg.Secret), nil
	},
		jwt.WithIssuer(i.cfg.Issuer),
		jwt.WithAudience(i.cfg.Audience),
	)
	if err != nil {
		return application.Claims{}, err
	}

	claimsValue, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid {
		return application.Claims{}, fmt.Errorf("jwt: invalid token")
	}

	userID, err := uuid.Parse(claimsValue.Subject)
	if err != nil {
		return application.Claims{}, fmt.Errorf("jwt: invalid subject: %w", err)
	}

	return application.Claims{UserID: userID, SessionID: claimsValue.SessionID}, nil
}
