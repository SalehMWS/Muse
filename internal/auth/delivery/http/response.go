package http

import (
	"time"

	"github.com/SalehMWS/Muse/internal/auth/domain"
)

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

func newUserResponse(user domain.User) UserResponse {
	return UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email.String(),
		DisplayName:   user.DisplayName,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func expiresInSeconds(expiresAt time.Time) int64 {
	return int64(time.Until(expiresAt).Seconds())
}
