package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Neo4j    Neo4jConfig
	Redis    RedisConfig
	MinIO    MinIOConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Host        string
	Port        int
	Env         string   // "development", "production"
	CORSOrigins []string // allowed CORS origins
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

type Neo4jConfig struct {
	URI      string
	User     string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Bucket    string
}

type AuthConfig struct {
	JWTSecret          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	// EncryptionKey is a 64-char hex string (32 raw bytes) used for AES-256-GCM.
	EncryptionKey string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Port:        getEnvInt("SERVER_PORT", 8080),
			Env:         getEnv("APP_ENV", "development"),
			CORSOrigins: getEnvStringSlice("CORS_ORIGINS", "http://localhost:3000"),
		},
		Postgres: PostgresConfig{
			Host:     getEnv("PG_HOST", "localhost"),
			Port:     getEnvInt("PG_PORT", 5432),
			User:     getEnv("PG_USER", "tc_app"),
			Password: getEnv("PG_PASSWORD", ""),
			DBName:   getEnv("PG_DBNAME", "trueconnect"),
			SSLMode:  getEnv("PG_SSLMODE", "disable"),
		},
		Neo4j: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			User:     getEnv("NEO4J_USER", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey: getEnv("MINIO_SECRET_KEY", ""),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			Bucket:    getEnv("MINIO_BUCKET", "trueconnect"),
		},
		Auth: AuthConfig{
			JWTSecret:          getEnv("JWT_SECRET", ""),
			AccessTokenExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			EncryptionKey:      getEnv("ENCRYPTION_KEY", ""),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Postgres.Password == "" {
		return fmt.Errorf("PG_PASSWORD is required")
	}
	if c.Neo4j.Password == "" {
		return fmt.Errorf("NEO4J_PASSWORD is required")
	}
	if c.Redis.Password == "" && c.Server.Env == "production" {
		return fmt.Errorf("REDIS_PASSWORD is required in production")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Auth.EncryptionKey == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if len(c.Auth.EncryptionKey) != 64 {
		return fmt.Errorf("ENCRYPTION_KEY must be a 64-character hex string (32 raw bytes)")
	}
	if c.MinIO.AccessKey == "" || c.MinIO.SecretKey == "" {
		return fmt.Errorf("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required")
	}
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvStringSlice(key, fallback string) []string {
	v := getEnv(key, fallback)
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
