package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv     string
	AppPort    int
	APIVersion string

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
