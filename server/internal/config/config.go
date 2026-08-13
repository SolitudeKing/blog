package config

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

type Config struct {
	AppEnv      string
	AppPort     int
	APIVersion  string
	SiteBaseURL string

	MySQLDSN      string
	RedisAddr     string
	RedisPassword string
	// RedisUsername 对应 REDIS_USER，用于 Redis 6+ 的 ACL 鉴权；
	// 留空时等价于仅密码，兼容旧版本 Redis。
	RedisUsername string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	AdminUsername string
	AdminPassword string
	AdminNickname string

	StorageDriver    string
	StorageLocalRoot string

	// 以下 STORAGE_S3_* 仅在 StorageDriver == "s3" 时使用。
	// Endpoint 支持 "host[:port]" 或 "http(s)://host[:port]"。
	// PublicURL 推荐形如 "https://cdn.example.com/<bucket>"，
	// 留空时按 endpoint + bucket 自动拼接。
	StorageS3Endpoint  string
	StorageS3AccessKey string
	StorageS3SecretKey string
	StorageS3Bucket    string
	StorageS3Region    string
	StorageS3UseSSL    bool
	StorageS3PublicURL string
}

type envLookup func(string) (string, bool)

func Load() (Config, error) {
	if err := loadFirstDotEnv(".env", "../.env"); err != nil {
		return Config{}, err
	}

	mysqlDSN, err := loadMySQLDSN()
	if err != nil {
		return Config{}, fmt.Errorf("load MySQL connection configuration: %w", err)
	}
	redisAddr, err := loadRedisAddr()
	if err != nil {
		return Config{}, fmt.Errorf("load Redis connection configuration: %w", err)
	}

	return Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		AppPort:          getEnvInt("APP_PORT", 8080),
		APIVersion:       getEnv("APP_API_VERSION", "v1"),
		SiteBaseURL:      strings.TrimRight(getEnv("SITE_BASE_URL", ""), "/"),
		MySQLDSN:         mysqlDSN,
		RedisAddr:        redisAddr,
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		RedisUsername:    getEnv("REDIS_USER", ""),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret"),
		JWTAccessTTL:     time.Duration(getEnvInt("JWT_ACCESS_TTL_MINUTES", 30)) * time.Minute,
		JWTRefreshTTL:    time.Duration(getEnvInt("JWT_REFRESH_TTL_DAYS", 14)) * 24 * time.Hour,
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "admin"),
		AdminNickname:    getEnv("ADMIN_NICKNAME", "Solitude King"),
		StorageDriver:    getEnv("STORAGE_DRIVER", "local"),
		StorageLocalRoot: getEnv("STORAGE_LOCAL_ROOT", "./storage/uploads"),
		StorageS3Endpoint:  getEnv("STORAGE_S3_ENDPOINT", ""),
		StorageS3AccessKey: getEnv("STORAGE_S3_ACCESS_KEY", ""),
		StorageS3SecretKey: getEnv("STORAGE_S3_SECRET_KEY", ""),
		StorageS3Bucket:    getEnv("STORAGE_S3_BUCKET", ""),
		StorageS3Region:    getEnv("STORAGE_S3_REGION", "us-east-1"),
		StorageS3UseSSL:    getEnvBool("STORAGE_S3_USE_SSL", false),
		StorageS3PublicURL: strings.TrimRight(getEnv("STORAGE_S3_PUBLIC_URL", ""), "/"),
	}, nil
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

// Validate 校验 API 启动所需的基础配置，并在生产环境阻止常见的弱密钥与默认口令。
// MySQL 是业务数据的唯一可信来源，因此不允许通过空 DSN 隐式切换到内存数据。
func (c Config) Validate() error {
	if strings.TrimSpace(c.MySQLDSN) == "" {
		return errors.New("MySQL connection configuration is required")
	}
	mysqlConfig, err := mysqlDriver.ParseDSN(c.MySQLDSN)
	if err != nil {
		return fmt.Errorf("invalid MySQL connection configuration: %w", err)
	}
	if strings.TrimSpace(mysqlConfig.DBName) == "" {
		return errors.New("MySQL database name is required")
	}
	if strings.TrimSpace(c.RedisAddr) != "" {
		if err := validateHostPort("Redis address", c.RedisAddr); err != nil {
			return err
		}
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

	// STORAGE_DRIVER 必须在所有环境都校验：错误的 driver 会让上传静默走错分支。
	switch strings.ToLower(strings.TrimSpace(c.StorageDriver)) {
	case "", "local":
		// 默认 local，无需额外校验。
	case "s3":
		for _, field := range []struct{ name, value string }{
			{"STORAGE_S3_ENDPOINT", c.StorageS3Endpoint},
			{"STORAGE_S3_ACCESS_KEY", c.StorageS3AccessKey},
			{"STORAGE_S3_SECRET_KEY", c.StorageS3SecretKey},
			{"STORAGE_S3_BUCKET", c.StorageS3Bucket},
		} {
			if strings.TrimSpace(field.value) == "" {
				return fmt.Errorf("%s is required when STORAGE_DRIVER=s3", field.name)
			}
		}
	default:
		return fmt.Errorf("unsupported STORAGE_DRIVER %q (allowed: local, s3)", c.StorageDriver)
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

func loadMySQLDSN() (string, error) {
	return resolveMySQLDSN(os.LookupEnv)
}

func resolveMySQLDSN(lookup envLookup) (string, error) {
	legacyDSN, _ := lookup("MYSQL_DSN")
	hasAtomicConfig := hasAnyEnv(
		lookup,
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_DATABASE",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
	)
	if strings.TrimSpace(legacyDSN) != "" {
		if hasAtomicConfig {
			return "", errors.New("MYSQL_DSN cannot be combined with atomic MYSQL_* connection variables")
		}
		parsed, err := mysqlDriver.ParseDSN(legacyDSN)
		if err != nil {
			return "", fmt.Errorf("MYSQL_DSN is invalid: %w", err)
		}
		if strings.TrimSpace(parsed.DBName) == "" {
			return "", errors.New("MYSQL_DSN must include a database name")
		}
		return legacyDSN, nil
	}

	host, err := getRequiredEnv(lookup, "MYSQL_HOST")
	if err != nil {
		return "", err
	}
	port, err := getRequiredPort(lookup, "MYSQL_PORT")
	if err != nil {
		return "", err
	}
	database, err := getRequiredEnv(lookup, "MYSQL_DATABASE")
	if err != nil {
		return "", err
	}
	user, err := getRequiredEnv(lookup, "MYSQL_USER")
	if err != nil {
		return "", err
	}
	password, err := getRequiredRawEnv(lookup, "MYSQL_PASSWORD")
	if err != nil {
		return "", err
	}

	driverConfig := mysqlDriver.NewConfig()
	driverConfig.User = user
	driverConfig.Passwd = password
	driverConfig.Net = "tcp"
	driverConfig.Addr = net.JoinHostPort(host, strconv.Itoa(port))
	driverConfig.DBName = database
	driverConfig.ParseTime = true
	driverConfig.Loc = time.UTC
	driverConfig.Params = map[string]string{"charset": "utf8mb4"}
	return driverConfig.FormatDSN(), nil
}

func loadRedisAddr() (string, error) {
	return resolveRedisAddr(os.LookupEnv)
}

func resolveRedisAddr(lookup envLookup) (string, error) {
	legacyAddr, _ := lookup("REDIS_ADDR")
	hasAtomicConfig := hasAnyEnv(lookup, "REDIS_HOST", "REDIS_PORT")
	if strings.TrimSpace(legacyAddr) != "" {
		if hasAtomicConfig {
			return "", errors.New("REDIS_ADDR cannot be combined with REDIS_HOST or REDIS_PORT")
		}
		if err := validateHostPort("REDIS_ADDR", legacyAddr); err != nil {
			return "", err
		}
		return legacyAddr, nil
	}
	if !hasAtomicConfig {
		return "localhost:6379", nil
	}

	host, err := getRequiredEnv(lookup, "REDIS_HOST")
	if err != nil {
		return "", err
	}
	port, err := getRequiredPort(lookup, "REDIS_PORT")
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(host, strconv.Itoa(port)), nil
}

func hasAnyEnv(lookup envLookup, keys ...string) bool {
	for _, key := range keys {
		if _, exists := lookup(key); exists {
			return true
		}
	}
	return false
}

func getRequiredEnv(lookup envLookup, key string) (string, error) {
	value, exists := lookup(key)
	value = strings.TrimSpace(value)
	if !exists || value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func getRequiredRawEnv(lookup envLookup, key string) (string, error) {
	value, exists := lookup(key)
	if !exists || value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func getRequiredPort(lookup envLookup, key string) (int, error) {
	value, err := getRequiredEnv(lookup, key)
	if err != nil {
		return 0, err
	}
	return parsePort(key, value)
}

func parsePort(name string, value string) (int, error) {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("%s must be an integer between 1 and 65535", name)
	}
	return port, nil
}

func validateHostPort(name string, address string) error {
	host, portValue, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("%s must use host:port format: %w", name, err)
	}
	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("%s host is required", name)
	}
	if _, err := parsePort(name+" port", portValue); err != nil {
		return err
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

// getEnvBool 解析布尔环境变量：接受 "1"/"true"/"TRUE"/"yes"/"on" 为 true，
// 其他非空值返回 fallback。空值或未设置返回 fallback。
func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func loadFirstDotEnv(paths ...string) error {
	for _, path := range paths {
		loaded, err := loadDotEnv(path)
		if err != nil {
			return err
		}
		if loaded {
			return nil
		}
	}
	return nil
}

func loadDotEnv(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("open dotenv file %q: %w", path, err)
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
			if err := os.Setenv(key, value); err != nil {
				return false, fmt.Errorf("set %s from dotenv file %q: %w", key, path, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("read dotenv file %q: %w", path, err)
	}
	return true, nil
}
