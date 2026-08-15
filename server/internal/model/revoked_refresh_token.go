package model

import "time"

// RevokedRefreshToken 记录被显式撤销的 refresh token 的 jti。
// 启动时由 AutoMigrate 建表；后台在 auth/logout 时写入，
// auth/refresh 命中记录后返回 CodeInvalidToken。
//
// expires_at 与 refresh token 的有效期对齐；过期后无需保留，由
// `DELETE WHERE expires_at < now()` 周期清理（详见 docs/backend-optimization）。
type RevokedRefreshToken struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	JTI       string    `gorm:"size:64;not null;uniqueIndex" json:"jti"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	RevokedAt time.Time `gorm:"not null" json:"revoked_at"`
}
