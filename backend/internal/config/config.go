package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpiryHours int    `yaml:"expiry_hours"`
}

type AlgorithmConfig struct {
	ServiceURL string `yaml:"service_url"`
}

type RedisConfig struct {
	URL string `yaml:"url"`
}

type StripeConfig struct {
	SecretKey     string `yaml:"secret_key"`
	WebhookSecret string `yaml:"webhook_secret"`
}

type FrontendConfig struct {
	URL string `yaml:"url"`
}

type EncryptionConfig struct {
	Key string `yaml:"key"`
}

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	Algorithm  AlgorithmConfig  `yaml:"algorithm"`
	Redis      RedisConfig      `yaml:"redis"`
	Stripe     StripeConfig     `yaml:"stripe"`
	Frontend   FrontendConfig   `yaml:"frontend"`
	Encryption EncryptionConfig `yaml:"encryption"`
}

func Load() (*Config, error) {
	configFile := flag.String("f", "configs/config.dev.yaml", "config file path")
	flag.Parse()

	return LoadFromFile(*configFile)
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	cfg.Server.Port = getEnvInt("APP_PORT", cfg.Server.Port)
	cfg.Database.Host = getEnv("DB_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvInt("DB_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("DB_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("DB_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnv("DB_NAME", cfg.Database.Name)
	cfg.JWT.Secret = getEnv("JWT_SECRET", cfg.JWT.Secret)
	cfg.JWT.ExpiryHours = getEnvInt("JWT_EXPIRY_HOURS", cfg.JWT.ExpiryHours)
	cfg.Algorithm.ServiceURL = getEnv("ALGORITHM_SERVICE_URL", cfg.Algorithm.ServiceURL)
	cfg.Redis.URL = getEnv("REDIS_URL", cfg.Redis.URL)
	cfg.Stripe.SecretKey = getEnv("STRIPE_SECRET_KEY", cfg.Stripe.SecretKey)
	cfg.Stripe.WebhookSecret = getEnv("STRIPE_WEBHOOK_SECRET", cfg.Stripe.WebhookSecret)
	cfg.Frontend.URL = getEnv("FRONTEND_URL", cfg.Frontend.URL)
	cfg.Encryption.Key = getEnv("ENCRYPTION_KEY", cfg.Encryption.Key)
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password, c.Database.Name,
	)
}

func (c *Config) AppAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

// Convenience accessors — keep existing caller code working.
func (c *Config) JWTSecret() string          { return c.JWT.Secret }
func (c *Config) JWTExpiryHours() int        { return c.JWT.ExpiryHours }
func (c *Config) AlgorithmServiceURL() string { return c.Algorithm.ServiceURL }
func (c *Config) StripeSecretKey() string     { return c.Stripe.SecretKey }
func (c *Config) StripeWebhookSecret() string { return c.Stripe.WebhookSecret }
func (c *Config) FrontendURL() string         { return c.Frontend.URL }
func (c *Config) EncryptionKey() string       { return c.Encryption.Key }

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

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
