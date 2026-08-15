package config

import (
	"strings"
	"testing"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

func TestValidateRequiresPersistentDatabaseAndPositiveTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "missing mysql", mutate: func(cfg *Config) { cfg.MySQLDSN = "" }},
		{name: "missing mysql database", mutate: func(cfg *Config) {
			cfg.MySQLDSN = "blog:secret@tcp(mysql:3306)/"
		}},
		{name: "invalid redis address", mutate: func(cfg *Config) { cfg.RedisAddr = "redis" }},
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

func TestResolveBuildsConnectionConfigFromAtomicEnvironment(t *testing.T) {
	t.Parallel()

	values := atomicConnectionEnv()
	mysqlDSN, err := resolveMySQLDSN(mapEnvLookup(values))
	if err != nil {
		t.Fatalf("resolveMySQLDSN() error = %v", err)
	}

	mysqlConfig, err := mysqlDriver.ParseDSN(mysqlDSN)
	if err != nil {
		t.Fatalf("ParseDSN() error = %v", err)
	}
	if mysqlConfig.User != "blog" {
		t.Fatalf("MySQL user = %q, want blog", mysqlConfig.User)
	}
	if mysqlConfig.Passwd != "p@ss:/#?%=value" {
		t.Fatal("MySQL password was not preserved")
	}
	if mysqlConfig.Net != "tcp" {
		t.Fatalf("MySQL network = %q, want tcp", mysqlConfig.Net)
	}
	if mysqlConfig.Addr != "[2001:db8::10]:3307" {
		t.Fatalf("MySQL address = %q, want [2001:db8::10]:3307", mysqlConfig.Addr)
	}
	if mysqlConfig.DBName != "blog" {
		t.Fatalf("MySQL database = %q, want blog", mysqlConfig.DBName)
	}
	if !mysqlConfig.ParseTime {
		t.Fatal("MySQL ParseTime = false, want true")
	}
	if mysqlConfig.Loc != time.UTC {
		t.Fatalf("MySQL location = %v, want UTC", mysqlConfig.Loc)
	}
	if mysqlConfig.Params["charset"] != "utf8mb4" {
		t.Fatalf("MySQL charset = %q, want utf8mb4", mysqlConfig.Params["charset"])
	}
	redisAddr, err := resolveRedisAddr(mapEnvLookup(values))
	if err != nil {
		t.Fatalf("resolveRedisAddr() error = %v", err)
	}
	if redisAddr != "[2001:db8::20]:6380" {
		t.Fatalf("Redis address = %q, want [2001:db8::20]:6380", redisAddr)
	}
}

func TestResolveConnectionConfigSupportsLegacyValues(t *testing.T) {
	t.Parallel()

	legacyMySQLDSN := "blog:secret@tcp(mysql:3306)/blog?parseTime=true"
	legacyRedisAddr := "redis:6379"
	lookup := mapEnvLookup(map[string]string{
		"MYSQL_DSN":      legacyMySQLDSN,
		"REDIS_ADDR":     legacyRedisAddr,
		"REDIS_PASSWORD": "shared-secret",
	})

	mysqlDSN, err := resolveMySQLDSN(lookup)
	if err != nil {
		t.Fatalf("resolveMySQLDSN() error = %v", err)
	}
	if mysqlDSN != legacyMySQLDSN {
		t.Fatalf("resolveMySQLDSN() = %q, want legacy value", mysqlDSN)
	}
	redisAddr, err := resolveRedisAddr(lookup)
	if err != nil {
		t.Fatalf("resolveRedisAddr() error = %v", err)
	}
	if redisAddr != legacyRedisAddr {
		t.Fatalf("resolveRedisAddr() = %q, want legacy value", redisAddr)
	}
}

func TestResolveConnectionConfigRejectsLegacyAndAtomicValues(t *testing.T) {
	t.Parallel()

	for _, key := range []string{
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_DATABASE",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
	} {
		t.Run("mysql "+key, func(t *testing.T) {
			values := map[string]string{
				"MYSQL_DSN": "blog:secret@tcp(mysql:3306)/blog",
				key:         "",
			}
			if _, err := resolveMySQLDSN(mapEnvLookup(values)); err == nil {
				t.Fatal("resolveMySQLDSN() error = nil, want conflicting configuration error")
			}
		})
	}

	for _, key := range []string{"REDIS_HOST", "REDIS_PORT"} {
		t.Run("redis "+key, func(t *testing.T) {
			values := map[string]string{
				"REDIS_ADDR": "redis:6379",
				key:          "",
			}
			if _, err := resolveRedisAddr(mapEnvLookup(values)); err == nil {
				t.Fatal("resolveRedisAddr() error = nil, want conflicting configuration error")
			}
		})
	}
}

func TestResolveConnectionConfigRejectsInvalidAtomicValues(t *testing.T) {
	t.Parallel()

	t.Run("mysql port", func(t *testing.T) {
		values := atomicConnectionEnv()
		values["MYSQL_PORT"] = "70000"

		if _, err := resolveMySQLDSN(mapEnvLookup(values)); err == nil {
			t.Fatal("resolveMySQLDSN() error = nil, want invalid port error")
		}
	})

	t.Run("mysql password", func(t *testing.T) {
		values := atomicConnectionEnv()
		values["MYSQL_PASSWORD"] = ""

		if _, err := resolveMySQLDSN(mapEnvLookup(values)); err == nil {
			t.Fatal("resolveMySQLDSN() error = nil, want missing password error")
		}
	})

	t.Run("redis port", func(t *testing.T) {
		values := atomicConnectionEnv()
		values["REDIS_PORT"] = "not-a-port"

		if _, err := resolveRedisAddr(mapEnvLookup(values)); err == nil {
			t.Fatal("resolveRedisAddr() error = nil, want invalid port error")
		}
	})

	t.Run("redis partial", func(t *testing.T) {
		if _, err := resolveRedisAddr(mapEnvLookup(map[string]string{"REDIS_HOST": "redis"})); err == nil {
			t.Fatal("resolveRedisAddr() error = nil, want missing port error")
		}
	})
}

func TestParsePortRejectsOutOfRangeAndMalformedValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "0", "65536", "-1", "not-a-port"} {
		t.Run(value, func(t *testing.T) {
			if _, err := parsePort("TEST_PORT", value); err == nil {
				t.Fatalf("parsePort(%q) error = nil, want validation error", value)
			}
		})
	}
	for value, want := range map[string]int{"1": 1, "65535": 65535} {
		t.Run(value, func(t *testing.T) {
			got, err := parsePort("TEST_PORT", value)
			if err != nil {
				t.Fatalf("parsePort(%q) error = %v", value, err)
			}
			if got != want {
				t.Fatalf("parsePort(%q) = %d, want %d", value, got, want)
			}
		})
	}
}

func TestResolveMySQLDSNAppendsTLSParam(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		tlsValue     string
		wantContains string
		wantOmit     bool
	}{
		{name: "false omitted", tlsValue: "false", wantOmit: true},
		{name: "empty omitted", tlsValue: "", wantOmit: true},
		{name: "true appends tls=true", tlsValue: "true", wantContains: "tls=true"},
		{name: "skip-verify appends tls=skip-verify", tlsValue: "skip-verify", wantContains: "tls=skip-verify"},
		{name: "ON alias", tlsValue: "ON", wantContains: "tls=true"},
		{name: "OFF alias", tlsValue: "OFF", wantOmit: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			values := atomicConnectionEnv()
			if tc.tlsValue != "" {
				values["MYSQL_TLS"] = tc.tlsValue
			}

			dsn, err := resolveMySQLDSN(mapEnvLookup(values))
			if err != nil {
				t.Fatalf("resolveMySQLDSN() error = %v", err)
			}

			if tc.wantOmit {
				if strings.Contains(dsn, "tls=") {
					t.Fatalf("DSN = %q, want no tls= param", dsn)
				}
				return
			}
			if !strings.Contains(dsn, tc.wantContains) {
				t.Fatalf("DSN = %q, want substring %q", dsn, tc.wantContains)
			}
		})
	}
}

func TestResolveMySQLDSNRejectsUnsupportedTLS(t *testing.T) {
	t.Parallel()

	values := atomicConnectionEnv()
	values["MYSQL_TLS"] = "custom-name"

	if _, err := resolveMySQLDSN(mapEnvLookup(values)); err == nil {
		t.Fatal("resolveMySQLDSN() error = nil, want unsupported value error")
	}
}

func TestMySQLTLSParam(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		raw        string
		wantParam  string
		wantOK     bool
		wantErrMsg string
	}{
		{raw: "", wantParam: "", wantOK: false},
		{raw: "false", wantParam: "", wantOK: false},
		{raw: "TRUE", wantParam: "true", wantOK: true},
		{raw: "skip-verify", wantParam: "skip-verify", wantOK: true},
		{raw: "bogus", wantParam: "", wantOK: false, wantErrMsg: "unsupported value"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			param, ok, err := mysqlTLSParam(tc.raw)
			if tc.wantErrMsg != "" {
				if err == nil {
					t.Fatalf("mysqlTLSParam(%q) error = nil, want error containing %q", tc.raw, tc.wantErrMsg)
				}
				if !strings.Contains(err.Error(), tc.wantErrMsg) {
					t.Fatalf("mysqlTLSParam(%q) error = %v, want contains %q", tc.raw, err, tc.wantErrMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("mysqlTLSParam(%q) unexpected error: %v", tc.raw, err)
			}
			if ok != tc.wantOK {
				t.Fatalf("mysqlTLSParam(%q) ok = %v, want %v", tc.raw, ok, tc.wantOK)
			}
			if param != tc.wantParam {
				t.Fatalf("mysqlTLSParam(%q) param = %q, want %q", tc.raw, param, tc.wantParam)
			}
		})
	}
}

func TestValidateRejectsUnsupportedTLS(t *testing.T) {
	t.Parallel()

	t.Run("mysql", func(t *testing.T) {
		cfg := validTestConfig()
		cfg.MySQLTLS = "custom-name"
		if err := cfg.Validate(); err == nil {
			t.Fatal("Validate() error = nil, want MYSQL_TLS validation error")
		}
	})

	t.Run("redis", func(t *testing.T) {
		cfg := validTestConfig()
		cfg.RedisTLS = "custom-name"
		if err := cfg.Validate(); err == nil {
			t.Fatal("Validate() error = nil, want REDIS_TLS validation error")
		}
	})
}

func TestValidateAcceptsSupportedTLS(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"false", "true", "skip-verify"} {
		t.Run(value, func(t *testing.T) {
			cfg := validTestConfig()
			cfg.MySQLTLS = value
			cfg.RedisTLS = value
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() error = %v, want nil for TLS=%q", err, value)
			}
		})
	}
}

func TestValidateRejectsNegativeRedisDB(t *testing.T) {
	t.Parallel()

	cfg := validTestConfig()
	cfg.RedisDB = -1
	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want REDIS_DB validation error")
	}
}

func TestValidateRequiresRedisAddrInProduction(t *testing.T) {
	t.Parallel()

	cfg := validTestConfig()
	cfg.RedisAddr = ""
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want Redis addr required in production")
	}
	if !strings.Contains(err.Error(), "REDIS_HOST") {
		t.Fatalf("Validate() error = %v, want mention of REDIS_HOST", err)
	}
}

func TestValidateAcceptsMissingRedisAddrInDevelopment(t *testing.T) {
	t.Parallel()

	cfg := validTestConfig()
	cfg.AppEnv = "development"
	cfg.RedisAddr = ""
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil in development with no Redis", err)
	}
}

func TestLoadReadsRedisDB(t *testing.T) {
	t.Parallel()

	// 通过 Validate() 路径验证正整数 REDIS_DB 可被接受，确保配置加载后
	// 不会被默认的 0 静默覆盖，也不会被误认为负数。
	cfg := validTestConfig()
	cfg.RedisDB = 3
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil for RedisDB=3", err)
	}
}

func atomicConnectionEnv() map[string]string {
	return map[string]string{
		"MYSQL_HOST":     "2001:db8::10",
		"MYSQL_PORT":     "3307",
		"MYSQL_DATABASE": "blog",
		"MYSQL_USER":     "blog",
		"MYSQL_PASSWORD": "p@ss:/#?%=value",
		"REDIS_HOST":     "2001:db8::20",
		"REDIS_PORT":     "6380",
	}
}

func mapEnvLookup(values map[string]string) envLookup {
	return func(key string) (string, bool) {
		value, exists := values[key]
		return value, exists
	}
}

func validTestConfig() Config {
	return Config{
		AppEnv:           "production",
		MySQLDSN:         "blog:secret@tcp(mysql:3306)/blog",
		RedisAddr:        "redis:6379",
		JWTAccessSecret:  "access-" + strings.Repeat("a", 40),
		JWTRefreshSecret: "refresh-" + strings.Repeat("b", 40),
		JWTAccessTTL:     30 * time.Minute,
		JWTRefreshTTL:    14 * 24 * time.Hour,
		AdminUsername:    "owner",
		AdminPassword:    "a-strong-owner-password",
	}
}
