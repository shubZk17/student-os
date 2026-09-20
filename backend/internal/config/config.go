package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                  string
	Env                   string
	AllowedOrigins        []string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	DBSSLMode             string
	JWTSecret             string
	JWTExpirationMinutes  int
	RefreshExpirationDays int
	TrustedPlatform       string
	DatabaseURL           string
	// IngestToken guards the scheduled-ingest route. Empty disables the route entirely.
	IngestToken string
}

const devJWTSecret = "dev_only_jwt_secret_do_not_use_in_production"

// Load reads configuration from the environment. In production it refuses to start
// with dev defaults for secrets.
func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	env := getEnv("ENV", "development")
	origins := strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:3000"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "15"))
	if err != nil || jwtExp <= 0 {
		return nil, fmt.Errorf("JWT_EXPIRATION_MINUTES must be a positive integer")
	}
	refExp, err := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRATION_DAYS", "7"))
	if err != nil || refExp <= 0 {
		return nil, fmt.Errorf("REFRESH_TOKEN_EXPIRATION_DAYS must be a positive integer")
	}

	cfg := &Config{
		Port:                  port,
		Env:                   env,
		AllowedOrigins:        origins,
		DBHost:                getEnv("DB_HOST", "localhost"),
		DBPort:                getEnv("DB_PORT", "5432"),
		DBUser:                getEnv("DB_USER", "studentos"),
		DBPassword:            getEnv("DB_PASSWORD", "studentos_dev_password"),
		DBName:                getEnv("DB_NAME", "studentos_db"),
		DBSSLMode:             getEnv("DB_SSLMODE", "disable"),
		IngestToken:           getEnv("INGEST_TOKEN", ""),
		JWTSecret:             getEnv("JWT_SECRET", devJWTSecret),
		JWTExpirationMinutes:  jwtExp,
		RefreshExpirationDays: refExp,
		TrustedPlatform:       os.Getenv("TRUSTED_PLATFORM"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
	}

	if cfg.Env == "production" {
		if cfg.DatabaseURL == "" && os.Getenv("DB_PASSWORD") == "" {
			return nil, fmt.Errorf("DB_PASSWORD (or DATABASE_URL) must be set in production")
		}
		if cfg.DatabaseURL == "" && cfg.DBSSLMode == "disable" {
			return nil, fmt.Errorf("DB_SSLMODE=disable is not allowed in production")
		}
		for _, o := range cfg.AllowedOrigins {
			if o == "*" {
				return nil, fmt.Errorf("ALLOWED_ORIGINS=* is not allowed in production")
			}
		}
	}
	return cfg, nil
}

// ValidateJWT is called by the API server only; the worker never signs tokens.
func (c *Config) ValidateJWT() error {
	if c.Env == "production" && (c.JWTSecret == devJWTSecret || len(c.JWTSecret) < 32) {
		return fmt.Errorf("JWT_SECRET must be set to at least 32 characters in production")
	}
	return nil
}

// DatabaseDSN prefers DATABASE_URL (as given by hosted Postgres providers) over the DB_* parts.
func (c *Config) DatabaseDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.DBUser, c.DBPassword),
		Host:     c.DBHost + ":" + c.DBPort,
		Path:     c.DBName,
		RawQuery: "sslmode=" + url.QueryEscape(c.DBSSLMode),
	}
	return u.String()
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
