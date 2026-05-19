package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	DBHost              string
	DBPort              int
	DBUser              string
	DBPassword          string
	DBName              string
	JWTSecret           string
	JWTExpiryHours      int
	AlgorithmServiceURL string
	RedisURL            string
	AppPort             int
	StripeSecretKey     string
	StripeWebhookSecret string
	FrontendURL         string
	EncryptionKey       string
}

// Load reads configuration from environment variables and .env file.
// Environment variables take precedence over .env file values.
func Load() (*Config, error) {
	// Attempt to load .env file; ignore error if file does not exist.
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getEnvInt("DB_PORT", 5432),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          getEnv("DB_PASSWORD", "123456"),
		DBName:              getEnv("DB_NAME", "crm0_db"),
		JWTSecret:           getEnv("JWT_SECRET", "crm0-secret-key"),
		JWTExpiryHours:      getEnvInt("JWT_EXPIRY_HOURS", 72),
		AlgorithmServiceURL: getEnv("ALGORITHM_SERVICE_URL", "http://localhost:8001"),
		RedisURL:            getEnv("REDIS_URL", "localhost:6379"),
		AppPort:             getEnvInt("APP_PORT", 3000),
			StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
			FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:5173"),
			EncryptionKey:       getEnv("ENCRYPTION_KEY", "default-encryption-key-change-in-prod"),
	}

	return cfg, nil
}

// DSN returns the PostgreSQL data source name constructed from config values.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

// AppAddr returns the address the application server should listen on.
func (c *Config) AppAddr() string {
	return fmt.Sprintf(":%d", c.AppPort)
}

// getEnv reads an environment variable or returns the provided default value.
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// getEnvInt reads an integer environment variable or returns the provided default value.
func getEnvInt(key string, defaultVal int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}
