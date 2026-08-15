package service

import (
	"testing"
	"time"

	"solitude-blog/server/internal/config"
	apperrors "solitude-blog/server/internal/errors"
)

// TestAuthServiceSignsTokensWithDistinctJTI 验证 access/refresh 各自拿到不同的 jti。
func TestAuthServiceSignsTokensWithDistinctJTI(t *testing.T) {
	cfg := config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessTTL:     30 * time.Minute,
		JWTRefreshTTL:    24 * time.Hour,
	}
	auth := NewAuthService(cfg, nil)

	// 触发内存 admin 路径
	user, _ := auth.findUser(t.Context(), cfg.AdminUsername)
	if user.ID == 0 {
		// 配置未提供 AdminUsername，触发 fallback：admin/admin
		user, _ = auth.findUser(t.Context(), "admin")
	}
	if user.ID == 0 {
		t.Fatal("memory admin user not initialised")
	}

	pair, err := auth.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair() error = %v", err)
	}

	access, err := auth.VerifyAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if access.JTI == "" {
		t.Fatal("access token JTI is empty")
	}

	refresh, err := auth.parseToken(pair.RefreshToken, cfg.JWTRefreshSecret)
	if err != nil {
		t.Fatalf("parseToken(refresh) error = %v", err)
	}
	if refresh.JTI == "" {
		t.Fatal("refresh token JTI is empty")
	}
	if access.JTI == refresh.JTI {
		t.Fatalf("access and refresh share the same JTI %q", access.JTI)
	}
}

// TestAuthServiceLogoutRevokeIsNoopWithoutDB 验证内存模式下 logout 不报错。
func TestAuthServiceLogoutRevokeIsNoopWithoutDB(t *testing.T) {
	cfg := config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
	}
	auth := NewAuthService(cfg, nil)
	if err := auth.LogoutRevoke("dummy", "dummy"); err != nil {
		t.Fatalf("LogoutRevoke() error = %v, want nil", err)
	}
}

// TestAuthServiceParseTokenRejectsWrongType 验证 token_type=access 的 token
// 不能被当作 refresh token 刷新。
func TestAuthServiceParseTokenRejectsWrongType(t *testing.T) {
	cfg := config.Config{
		JWTAccessSecret:  "test-access-secret",
		JWTRefreshSecret: "test-refresh-secret",
		JWTAccessTTL:     30 * time.Minute,
		JWTRefreshTTL:    24 * time.Hour,
	}
	auth := NewAuthService(cfg, nil)
	user, _ := auth.findUser(t.Context(), "admin")
	pair, err := auth.issueTokenPair(user)
	if err != nil {
		t.Fatalf("issueTokenPair() error = %v", err)
	}
	// 把 access token 当 refresh token 提交，应该返回 CodeInvalidToken
	_, err = auth.Refresh(pair.AccessToken)
	if err == nil {
		t.Fatal("Refresh(access token) returned nil error")
	}
	appErr, ok := err.(apperrors.AppError)
	if !ok || appErr.Code != apperrors.CodeInvalidToken {
		t.Fatalf("Refresh(access token) error = %v, want CodeInvalidToken", err)
	}
}
