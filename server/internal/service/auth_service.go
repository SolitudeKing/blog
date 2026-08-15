package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"solitude-blog/server/internal/config"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

type AuthService struct {
	cfg          config.Config
	db           *gorm.DB
	memoryAdmin  model.User
	memoryLoaded bool
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type TokenClaims struct {
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	JTI       string `json:"jti"`
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
	if err := s.ensureNotRevoked(claims); err != nil {
		return TokenPair{}, err
	}

	// 轮换：撤销当前 refresh jti 并签发新对，避免旧 token 继续可用。
	if err := s.revokeRefreshToken(claims); err != nil {
		return TokenPair{}, err
	}

	return s.issueTokenPair(model.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		Status:   "active",
	})
}

// LogoutRevoke 把当前 access / refresh token 的 jti 一并写入撤销表，
// 同一对 token 之后任何 Refresh / VerifyAccessToken 调用都会返回 CodeInvalidToken。
// Logout 调用方应传入从 Authorization 头解析的 access token 与登录时拿到的 refresh token；
// 解析失败时整体返回错误，由 handler 把响应降级为 200 + 不持久化撤销。
func (s *AuthService) LogoutRevoke(accessToken string, refreshToken string) error {
	if s.db == nil {
		return nil
	}
	now := time.Now().UTC()
	if refreshToken != "" {
		claims, err := s.parseToken(refreshToken, s.cfg.JWTRefreshSecret)
		if err == nil && claims.TokenType == "refresh" {
			if err := s.insertRevocation(claims.JTI, claims.UserID, claims.ExpiresAt.Time, now); err != nil {
				return err
			}
		}
	}
	if accessToken != "" {
		claims, err := s.parseToken(accessToken, s.cfg.JWTAccessSecret)
		if err == nil && claims.TokenType == "access" {
			if err := s.insertRevocation(claims.JTI, claims.UserID, claims.ExpiresAt.Time, now); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *AuthService) revokeRefreshToken(claims TokenClaims) error {
	if s.db == nil || claims.JTI == "" {
		return nil
	}
	return s.insertRevocation(claims.JTI, claims.UserID, claims.ExpiresAt.Time, time.Now().UTC())
}

func (s *AuthService) ensureNotRevoked(claims TokenClaims) error {
	if s.db == nil || claims.JTI == "" {
		return nil
	}
	var count int64
	if err := s.db.Model(&model.RevokedRefreshToken{}).
		Where("jti = ? AND expires_at > ?", claims.JTI, time.Now().UTC()).
		Count(&count).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if count > 0 {
		return apperrors.New(apperrors.CodeInvalidToken)
	}
	return nil
}

func (s *AuthService) insertRevocation(jti string, userID uint64, expiresAt time.Time, revokedAt time.Time) error {
	if jti == "" || !expiresAt.After(revokedAt) {
		return nil
	}
	record := model.RevokedRefreshToken{
		JTI:       jti,
		UserID:    userID,
		ExpiresAt: expiresAt,
		RevokedAt: revokedAt,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return nil
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
		return s.memoryAdminUser(), nil
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

// memoryAdminUser 在启动阶段就把无 DB 模式下的管理员密码算一次，
// 避免每次登录都执行 bcrypt.DefaultCost（一次约 50–150ms）。
// 返回值是值类型，调用方不应就地修改。
func (s *AuthService) memoryAdminUser() model.User {
	if s.memoryLoaded {
		return s.memoryAdmin
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(s.cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		// bcrypt 在正常环境下不会失败；失败时退化为不可登录的状态，
		// 让 Login 返回 CodeInvalidCredentials 而不是悄悄吞掉错误。
		s.memoryLoaded = true
		return model.User{Username: s.cfg.AdminUsername, Status: "disabled"}
	}
	s.memoryAdmin = model.User{
		ID:           1,
		Username:     s.cfg.AdminUsername,
		PasswordHash: string(passwordHash),
		Nickname:     s.cfg.AdminNickname,
		Role:         "owner",
		Status:       "active",
	}
	s.memoryLoaded = true
	return s.memoryAdmin
}

func (s *AuthService) issueTokenPair(user model.User) (TokenPair, error) {
	accessExpiresAt := time.Now().UTC().Add(s.cfg.JWTAccessTTL)
	refreshExpiresAt := time.Now().UTC().Add(s.cfg.JWTRefreshTTL)

	accessJTI := newTokenJTI()
	refreshJTI := newTokenJTI()

	accessToken, err := s.signToken(user, "access", accessJTI, accessExpiresAt, s.cfg.JWTAccessSecret)
	if err != nil {
		return TokenPair{}, apperrors.New(apperrors.CodeInternalServerError)
	}
	refreshToken, err := s.signToken(user, "refresh", refreshJTI, refreshExpiresAt, s.cfg.JWTRefreshSecret)
	if err != nil {
		return TokenPair{}, apperrors.New(apperrors.CodeInternalServerError)
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

func (s *AuthService) signToken(user model.User, tokenType string, jti string, expiresAt time.Time, secret string) (string, error) {
	claims := TokenClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: tokenType,
		JTI:       jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Subject:   user.Username,
			ID:        jti,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func newTokenJTI() string {
	return uuid.NewString()
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
