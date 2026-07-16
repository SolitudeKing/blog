package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv      string
	AppPort     int
	APIVersion  string
	SiteBaseURL string

	MySQLDSN      string
	RedisAddr     string
	RedisPassword string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	AdminUsername string
	AdminPassword string
	AdminNickname string

	StorageDriver    string
	StorageLocalRoot string
}

func Load() Config {
	loadDotEnv(".env")
	loadDotEnv("../.env")

	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		AppPort:          getEnvInt("APP_PORT", 8080),
		APIVersion:       getEnv("APP_API_VERSION", "v1"),
		SiteBaseURL:      strings.TrimRight(getEnv("SITE_BASE_URL", ""), "/"),
		MySQLDSN:         getEnv("MYSQL_DSN", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret"),
		JWTAccessTTL:     time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 30)) * time.Minute,
		JWTRefreshTTL:    time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 14)) * 24 * time.Hour,
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin"),
		AdminNickname:    getEnv("ADMIN_NICKNAME", "Solitude King"),
		StorageDriver:    getEnv("STORAGE_DRIVER", "local"),
		StorageLocalRoot: getEnv("STORAGE_LOCAL_ROOT", "./storage/uploads"),
	}
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

// Validate 校验 API 启动所需的基础配置，并在生产环境阻止常见的弱密钥与默认口令。
// MySQL 是业务数据的唯一可信来源，因此不允许通过空 DSN 隐式切换到内存数据。
func (c Config) Validate() error {
	if strings.TrimSpace(c.MySQLDSN) == "" {
		return errors.New("MYSQL_DSN is required")
	}
	if c.JWTAccessTTL <= 0 {
		return errors.New("JWT_ACCESS_TTL_MINUTES must be greater than zero")
	}
	if c.JWTRefreshTTL <= 0 {
		return errors.New("JWT_REFRESH_TTL_DAYS must be greater than zero")
	}
	if strings.TrimSpace(c.JWTAccessSecret) == "" || strings.TrimSpace(c.JWTRefreshSecret) == "" {
		return errors.New("JWT secrets are required")
	}
	if strings.TrimSpace(c.AdminUsername) == "" || strings.TrimSpace(c.AdminPassword) == "" {
		return errors.New("admin username and password are required")
	}

	if !strings.EqualFold(strings.TrimSpace(c.AppEnv), "production") {
		return nil
	}

	accessSecret := strings.TrimSpace(c.JWTAccessSecret)
	refreshSecret := strings.TrimSpace(c.JWTRefreshSecret)
	if accessSecret == refreshSecret {
		return errors.New("JWT access and refresh secrets must be different in production")
	}
	if insecureJWTSecret(accessSecret) || insecureJWTSecret(refreshSecret) {
		return errors.New("development or placeholder JWT secrets are not allowed in production")
	}
	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		return errors.New("JWT secrets must contain at least 32 characters in production")
	}
	if insecureAdminPassword(c.AdminPassword) {
		return errors.New("default admin password is not allowed in production")
	}
	return nil
}

func insecureJWTSecret(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(normalized, "dev-") || strings.Contains(normalized, "change-me")
}

func insecureAdminPassword(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "admin" || strings.Contains(normalized, "change-me")
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}
