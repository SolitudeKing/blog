package service

import (
	"context"
	"time"

	"solitude-blog/server/internal/config"
	apperrors "solitude-blog/server/internal/errors"
)

type AuthService struct {
	cfg config.Config
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func NewAuthService(cfg config.Config) *AuthService {
	return &AuthService{cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	if req.Username == "" || req.Password == "" {
		return TokenPair{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if req.Username != "admin" || req.Password != "admin" {
		return TokenPair{}, apperrors.New(apperrors.CodeInvalidCredentials)
	}
	return s.fakeTokenPair(), nil
}

func (s *AuthService) Refresh() TokenPair {
	return s.fakeTokenPair()
}

func (s *AuthService) fakeTokenPair() TokenPair {
	return TokenPair{
		AccessToken:  "dev-access-token",
		RefreshToken: "dev-refresh-token",
		ExpiresAt:    time.Now().UTC().Add(30 * time.Minute),
	}
}
