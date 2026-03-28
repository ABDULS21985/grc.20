// Package config provides centralized configuration management for ComplianceForge.
// It supports loading from .env files, YAML, and environment variables using Viper.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	Storage  StorageConfig
	Security SecurityConfig
	Logging  LoggingConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name        string   `mapstructure:"APP_NAME"`
	Env         string   `mapstructure:"APP_ENV"`
	Port        int      `mapstructure:"APP_PORT"`
	Host        string   `mapstructure:"APP_HOST"`
	Debug       bool     `mapstructure:"APP_DEBUG"`
	Version     string   `mapstructure:"APP_VERSION"`
	BaseURL     string   `mapstructure:"APP_BASE_URL"`
	CORSOrigins []string `mapstructure:"APP_CORS_ORIGINS"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string `mapstructure:"DB_HOST"`
	Port            int    `mapstructure:"DB_PORT"`
	User            string `mapstructure:"DB_USER"`
	Password        string `mapstructure:"DB_PASSWORD"`
	Name            string `mapstructure:"DB_NAME"`
	SSLMode         string `mapstructure:"DB_SSLMODE"`
	MaxOpenConns    int    `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int    `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime int    `mapstructure:"DB_CONN_MAX_LIFETIME"` // seconds
	ConnMaxIdleTime int    `mapstructure:"DB_CONN_MAX_IDLE_TIME"` // seconds
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     int    `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret              string        `mapstructure:"JWT_SECRET"`
	AccessTokenExpiry   time.Duration `mapstructure:"JWT_ACCESS_TOKEN_EXPIRY"`
	RefreshTokenExpiry  time.Duration `mapstructure:"JWT_REFRESH_TOKEN_EXPIRY"`
}

// SMTPConfig holds email server settings.
type SMTPConfig struct {
	Host     string `mapstructure:"SMTP_HOST"`
	Port     int    `mapstructure:"SMTP_PORT"`
	User     string `mapstructure:"SMTP_USER"`
	Password string `mapstructure:"SMTP_PASSWORD"`
	From     string `mapstructure:"SMTP_FROM"`
	FromName string `mapstructure:"SMTP_FROM_NAME"`
}

// StorageConfig holds file storage settings.
type StorageConfig struct {
	Driver    string `mapstructure:"STORAGE_DRIVER"` // "local" or "s3"
	LocalPath string `mapstructure:"STORAGE_LOCAL_PATH"`
	S3Bucket  string `mapstructure:"STORAGE_S3_BUCKET"`
	S3Region  string `mapstructure:"STORAGE_S3_REGION"`
}

// SecurityConfig holds encryption and security settings.
type SecurityConfig struct {
	EncryptionKey     string `mapstructure:"ENCRYPTION_KEY"`
	RateLimitRequests int    `mapstructure:"RATE_LIMIT_REQUESTS"`
	RateLimitWindow   string `mapstructure:"RATE_LIMIT_WINDOW"`
	DataResidency     string `mapstructure:"DATA_RESIDENCY_REGION"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
	Output string `mapstructure:"LOG_OUTPUT"`
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("APP_NAME", "complianceforge")
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_HOST", "0.0.0.0")
	v.SetDefault("APP_DEBUG", true)
	v.SetDefault("APP_VERSION", "1.0.0")
	v.SetDefault("APP_BASE_URL", "http://localhost:8080")
	v.SetDefault("APP_CORS_ORIGINS", "http://localhost:3000")

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "complianceforge")
	v.SetDefault("DB_PASSWORD", "complianceforge")
	v.SetDefault("DB_NAME", "complianceforge")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", 300)
	v.SetDefault("DB_CONN_MAX_IDLE_TIME", 30)

	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 6379)
	v.SetDefault("REDIS_DB", 0)

	v.SetDefault("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	v.SetDefault("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	v.SetDefault("SMTP_PORT", 1025)
	v.SetDefault("SMTP_FROM", "noreply@complianceforge.io")
	v.SetDefault("SMTP_FROM_NAME", "ComplianceForge")

	v.SetDefault("STORAGE_DRIVER", "local")
	v.SetDefault("STORAGE_LOCAL_PATH", "./uploads")

	v.SetDefault("RATE_LIMIT_REQUESTS", 100)
	v.SetDefault("RATE_LIMIT_WINDOW", "60s")
	v.SetDefault("DATA_RESIDENCY_REGION", "EU")

	v.SetDefault("LOG_LEVEL", "debug")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("LOG_OUTPUT", "stdout")

	// Read from .env file
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		// .env file is optional — environment variables take precedence
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only log, don't fail — env vars may be set directly
		}
	}

	// Environment variables override .env values
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Parse configuration
	cfg := &Config{}

	// App config
	cfg.App = AppConfig{
		Name:    v.GetString("APP_NAME"),
		Env:     v.GetString("APP_ENV"),
		Port:    v.GetInt("APP_PORT"),
		Host:    v.GetString("APP_HOST"),
		Debug:   v.GetBool("APP_DEBUG"),
		Version: v.GetString("APP_VERSION"),
		BaseURL: v.GetString("APP_BASE_URL"),
		CORSOrigins: strings.Split(v.GetString("APP_CORS_ORIGINS"), ","),
	}

	// Database config
	cfg.Database = DatabaseConfig{
		Host:            v.GetString("DB_HOST"),
		Port:            v.GetInt("DB_PORT"),
		User:            v.GetString("DB_USER"),
		Password:        v.GetString("DB_PASSWORD"),
		Name:            v.GetString("DB_NAME"),
		SSLMode:         v.GetString("DB_SSLMODE"),
		MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		ConnMaxLifetime: v.GetInt("DB_CONN_MAX_LIFETIME"),
		ConnMaxIdleTime: v.GetInt("DB_CONN_MAX_IDLE_TIME"),
	}

	// Redis config
	cfg.Redis = RedisConfig{
		Host:     v.GetString("REDIS_HOST"),
		Port:     v.GetInt("REDIS_PORT"),
		Password: v.GetString("REDIS_PASSWORD"),
		DB:       v.GetInt("REDIS_DB"),
	}

	// JWT config
	accessExpiry, _ := time.ParseDuration(v.GetString("JWT_ACCESS_TOKEN_EXPIRY"))
	refreshExpiry, _ := time.ParseDuration(v.GetString("JWT_REFRESH_TOKEN_EXPIRY"))
	cfg.JWT = JWTConfig{
		Secret:             v.GetString("JWT_SECRET"),
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
	}

	// SMTP config
	cfg.SMTP = SMTPConfig{
		Host:     v.GetString("SMTP_HOST"),
		Port:     v.GetInt("SMTP_PORT"),
		User:     v.GetString("SMTP_USER"),
		Password: v.GetString("SMTP_PASSWORD"),
		From:     v.GetString("SMTP_FROM"),
		FromName: v.GetString("SMTP_FROM_NAME"),
	}

	// Storage config
	cfg.Storage = StorageConfig{
		Driver:    v.GetString("STORAGE_DRIVER"),
		LocalPath: v.GetString("STORAGE_LOCAL_PATH"),
		S3Bucket:  v.GetString("STORAGE_S3_BUCKET"),
		S3Region:  v.GetString("STORAGE_S3_REGION"),
	}

	// Security config
	cfg.Security = SecurityConfig{
		EncryptionKey:     v.GetString("ENCRYPTION_KEY"),
		RateLimitRequests: v.GetInt("RATE_LIMIT_REQUESTS"),
		RateLimitWindow:   v.GetString("RATE_LIMIT_WINDOW"),
		DataResidency:     v.GetString("DATA_RESIDENCY_REGION"),
	}

	// Logging config
	cfg.Logging = LoggingConfig{
		Level:  v.GetString("LOG_LEVEL"),
		Format: v.GetString("LOG_FORMAT"),
		Output: v.GetString("LOG_OUTPUT"),
	}

	return cfg, nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
