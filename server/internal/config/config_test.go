package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateRequiresPersistentDatabaseAndPositiveTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "missing mysql", mutate: func(cfg *Config) { cfg.MySQLDSN = "" }},
		{name: "zero access ttl", mutate: func(cfg *Config) { cfg.JWTAccessTTL = 0 }},
		{name: "negative refresh ttl", mutate: func(cfg *Config) { cfg.JWTRefreshTTL = -time.Hour }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want validation error")
			}
		})
	}
}

func TestValidateRejectsInsecureProductionCredentials(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "same secrets", mutate: func(cfg *Config) { cfg.JWTRefreshSecret = cfg.JWTAccessSecret }},
		{name: "development secret", mutate: func(cfg *Config) { cfg.JWTAccessSecret = "dev-" + strings.Repeat("a", 40) }},
		{name: "placeholder secret", mutate: func(cfg *Config) { cfg.JWTRefreshSecret = "change-me-" + strings.Repeat("b", 40) }},
		{name: "short secret", mutate: func(cfg *Config) { cfg.JWTAccessSecret = "too-short" }},
		{name: "default admin password", mutate: func(cfg *Config) { cfg.AdminPassword = "admin" }},
		{name: "placeholder admin password", mutate: func(cfg *Config) { cfg.AdminPassword = "change-me-admin" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validTestConfig()
			tt.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate() error = nil, want production credential error")
			}
		})
	}
}

func TestValidateAcceptsSecureProductionConfig(t *testing.T) {
	t.Parallel()

	if err := validTestConfig().Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func validTestConfig() Config {
	return Config{
		AppEnv:           "production",
		MySQLDSN:         "blog:secret@tcp(mysql:3306)/blog",
		JWTAccessSecret:  "access-" + strings.Repeat("a", 40),
		JWTRefreshSecret: "refresh-" + strings.Repeat("b", 40),
		JWTAccessTTL:     30 * time.Minute,
		JWTRefreshTTL:    14 * 24 * time.Hour,
		AdminUsername:    "owner",
		AdminPassword:    "a-strong-owner-password",
	}
}
