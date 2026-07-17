package security_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/SalehMWS/Muse/internal/auth/infrastructure/security"
	"github.com/SalehMWS/Muse/internal/shared/config"
)

func testJWTConfig() config.JWT {
	return config.JWT{
		Secret:         "test-secret-at-least-32-characters-long",
		Issuer:         "novaflow-test",
		Audience:       "novaflow-test-api",
		AccessTokenTTL: 15 * time.Minute,
	}
}

type testClaims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

func TestJWTIssuer_IssueAndVerify(t *testing.T) {
	issuer := security.NewJWTIssuer(testJWTConfig())
	userID, _ := uuid.NewV7()
	sessionID, _ := uuid.NewV7()

	token, err := issuer.Issue(context.Background(), userID, sessionID)
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}
	if token.Value == "" {
		t.Fatal("Issue() returned an empty token")
	}

	claims, err := issuer.Verify(context.Background(), token.Value)
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("Verify() UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.SessionID != sessionID {
		t.Fatalf("Verify() SessionID = %v, want %v", claims.SessionID, sessionID)
	}
}

func TestJWTIssuer_Verify_WrongSignature(t *testing.T) {
	issuer := security.NewJWTIssuer(testJWTConfig())

	otherCfg := testJWTConfig()
	otherCfg.Secret = "a-totally-different-secret-value-here-too"
	otherIssuer := security.NewJWTIssuer(otherCfg)

	userID, _ := uuid.NewV7()
	sessionID, _ := uuid.NewV7()
	token, err := otherIssuer.Issue(context.Background(), userID, sessionID)
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}

	if _, err := issuer.Verify(context.Background(), token.Value); err == nil {
		t.Fatal("Verify() expected an error for a token signed with a different secret")
	}
}

func TestJWTIssuer_Verify_ExpiredToken(t *testing.T) {
	cfg := testJWTConfig()
	issuer := security.NewJWTIssuer(cfg)

	past := time.Now().Add(-time.Hour)
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims{
		SessionID: uuid.Must(uuid.NewV7()),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   uuid.Must(uuid.NewV7()).String(),
			Issuer:    cfg.Issuer,
			Audience:  jwt.ClaimStrings{cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(past.Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(past),
		},
	})
	signed, err := expired.SignedString([]byte(cfg.Secret))
	if err != nil {
		t.Fatalf("SignedString() unexpected error: %v", err)
	}

	if _, err := issuer.Verify(context.Background(), signed); err == nil {
		t.Fatal("Verify() expected an error for an expired token")
	}
}

func TestJWTIssuer_Verify_TamperedToken(t *testing.T) {
	issuer := security.NewJWTIssuer(testJWTConfig())
	userID, _ := uuid.NewV7()
	sessionID, _ := uuid.NewV7()

	token, err := issuer.Issue(context.Background(), userID, sessionID)
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}

	tampered := token.Value[:len(token.Value)-2] + "xx"

	if _, err := issuer.Verify(context.Background(), tampered); err == nil {
		t.Fatal("Verify() expected an error for a tampered token")
	}
}
