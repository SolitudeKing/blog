package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"solitude-blog/server/internal/config"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

type AuthService struct {
	cfg config.Config
	db  *gorm.DB
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type TokenClaims struct {
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{cfg: cfg, db: db}
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (TokenPair, error) {
	if req.Username == "" || req.Password == "" {
		return TokenPair{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}

	user, err := s.findUser(ctx, req.Username)
	if err != nil {
		return TokenPair{}, err
	}
	if user.Status != "active" {
		return TokenPair{}, apperrors.New(apperrors.CodeAccountDisabled)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return TokenPair{}, apperrors.New(apperrors.CodeInvalidCredentials)
	}

	return s.issueTokenPair(user)
}

func (s *AuthService) Refresh(refreshToken string) (TokenPair, error) {
	claims, err := s.parseToken(refreshToken, s.cfg.JWTRefreshSecret)
	if err != nil {
		return TokenPair{}, err
	}
	if claims.TokenType != "refresh" {
		return TokenPair{}, apperrors.New(apperrors.CodeInvalidToken)
	}

	return s.issueTokenPair(model.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		Status:   "active",
	})
}

func (s *AuthService) VerifyAccessToken(accessToken string) (TokenClaims, error) {
	claims, err := s.parseToken(accessToken, s.cfg.JWTAccessSecret)
	if err != nil {
		return TokenClaims{}, err
	}
	if claims.TokenType != "access" {
		return TokenClaims{}, apperrors.New(apperrors.CodeInvalidToken)
	}
	return claims, nil
}

func (s *AuthService) findUser(ctx context.Context, username string) (model.User, error) {
	if s.db == nil {
		if username != s.cfg.AdminUsername {
			return model.User{}, apperrors.New(apperrors.CodeInvalidCredentials)
		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return model.User{}, apperrors.New(apperrors.CodeInternalServerError)
		}
		return model.User{
			ID:           1,
			Username:     s.cfg.AdminUsername,
			PasswordHash: string(passwordHash),
			Nickname:     s.cfg.AdminNickname,
			Role:         "owner",
			Status:       "active",
		}, nil
	}

	var user model.User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, apperrors.New(apperrors.CodeInvalidCredentials)
	}
	if err != nil {
		return model.User{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return user, nil
}

func (s *AuthService) issueTokenPair(user model.User) (TokenPair, error) {
	accessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessTTL)
	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshTTL)

	accessToken, err := s.signToken(user, "access", accessExpiresAt, s.cfg.JWTAccessSecret)
	if err != nil {
		return TokenPair{}, apperrors.New(apperrors.CodeInternalServerError)
	}
	refreshToken, err := s.signToken(user, "refresh", refreshExpiresAt, s.cfg.JWTRefreshSecret)
	if err != nil {
		return TokenPair{}, apperrors.New(apperrors.CodeInternalServerError)
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func (s *AuthService) signToken(user model.User, tokenType string, expiresAt time.Time, secret string) (string, error) {
	claims := TokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   user.Username,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (s *AuthService) parseToken(rawToken string, secret string) (TokenClaims, error) {
	var claims TokenClaims
	token, err := jwt.ParseWithClaims(rawToken, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, apperrors.New(apperrors.CodeInvalidToken)
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return TokenClaims{}, apperrors.New(apperrors.CodeTokenExpired)
		}
		return TokenClaims{}, apperrors.New(apperrors.CodeInvalidToken)
	}
	if !token.Valid {
		return TokenClaims{}, apperrors.New(apperrors.CodeInvalidToken)
	}
	return claims, nil
}
