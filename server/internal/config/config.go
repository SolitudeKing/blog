package config

import (
	"fmt"
	"os"
	"strconv"
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

	StorageDriver    string
	StorageLocalRoot string
}

func Load() Config {
	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		AppPort:          getEnvInt("APP_PORT", 8080),
		APIVersion:       getEnv("APP_API_VERSION", "v1"),
		MySQLDSN:         getEnv("MYSQL_DSN", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret"),
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
