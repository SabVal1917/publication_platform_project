package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPPort  string
	Database  DBConfig
	JWTSecret string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func envOrDefault(name, defaultValue string) string {
	result := os.Getenv(name)
	if result == "" {
		return defaultValue
	}
	return result
}

func requiredEnv(name string) (string, error) {
	result := os.Getenv(name)
	if result == "" {
		return "", fmt.Errorf("required %s", name)
	}
	return result, nil
}

func Load() (Config, error) {
	dbc := DBConfig{
		Host:    envOrDefault("DB_HOST", "localhost"),
		Port:    envOrDefault("DB_PORT", "5432"),
		User:    envOrDefault("DB_USER", "user"),
		Name:    envOrDefault("DB_NAME", "blog"),
		SSLMode: envOrDefault("DB_SSLMODE", "disable"),
	}
	password, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	dbc.Password = password
	cfg := Config{
		Database: dbc,
		HTTPPort: envOrDefault("HTTP_PORT", "8080"),
	}
	JWTSecret, err := requiredEnv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}
	if len(JWTSecret) < 32 {
		return Config{}, fmt.Errorf("len of JWT_SECRET must be at least 32")
	}
	cfg.JWTSecret = JWTSecret
	return cfg, nil
}
