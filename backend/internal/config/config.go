package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	App      AppConfig
}

type ServerConfig struct {
	Host         string
	Port         int
	Timeout      int
	MaxBodySize  int64
	TrustedProxy string
}

type DatabaseConfig struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	DBName                 string
	SSLMode                string
	MaxConn                int
	MaxIdleConns           int
	ConnMaxIdleTimeSeconds int
	ConnMaxLifetimeSeconds int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type AuthConfig struct {
	JWTSecret             string
	AccessTokenTTLMinutes int
}

type AppConfig struct {
	Env   string
	Debug bool
}

func Load() *Config {
	godotenv.Load("configs/.env")

	cfg := &Config{
		Server: ServerConfig{
			Host:        getEnv("SERVER_HOST", "localhost"),
			Port:        getEnvInt("SERVER_PORT", 8080),
			Timeout:     getEnvInt("SERVER_TIMEOUT", 30),
			MaxBodySize: int64(getEnvInt("SERVER_MAX_BODY_SIZE", 1048576)),
		},
		Database: DatabaseConfig{
			Host:                   getEnv("DB_HOST", "localhost"),
			Port:                   getEnvInt("DB_PORT", 5432),
			User:                   getEnv("DB_USER", "postgres"),
			Password:               getEnv("DB_PASSWORD", ""),
			DBName:                 getEnv("DB_NAME", "relay"),
			SSLMode:                getEnv("DB_SSL_MODE", "disable"),
			MaxConn:                getEnvInt("DB_MAX_CONN", 25),
			MaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxIdleTimeSeconds: getEnvInt("DB_CONN_MAX_IDLE_TIME", 300),
			ConnMaxLifetimeSeconds: getEnvInt("DB_CONN_MAX_LIFETIME", 3600),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:             getEnv("JWT_SECRET", "dev-only-change-me"),
			AccessTokenTTLMinutes: getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 60),
		},
		App: AppConfig{
			Env:   getEnv("APP_ENV", "development"),
			Debug: getEnvBool("DEBUG", true),
		},
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	return cfg
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required (DB_NAME env var)")
	}

	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.IsProd() && c.Auth.JWTSecret == "dev-only-change-me" {
		return fmt.Errorf("JWT_SECRET must be changed in production")
	}

	if c.Auth.AccessTokenTTLMinutes < 1 {
		return fmt.Errorf("ACCESS_TOKEN_TTL_MINUTES must be at least 1")
	}

	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[c.App.Env] {
		return fmt.Errorf("invalid APP_ENV: %s (must be development, staging, or production)", c.App.Env)
	}

	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// Helper functions to read environment variables with type conversion

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	valStr := os.Getenv(key)
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "yes"
}

func (c *Config) IsProd() bool {
	return c.App.Env == "production"
}

func (c *Config) IsDev() bool {
	return c.App.Env == "development"
}
