package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTP      HTTPConfig
	Database  DBConfig
	JWTSecret string
}

type HTTPConfig struct {
	Port              string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	DrainDelay        time.Duration
}

type DBConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
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

func intEnv(name string, defaultValue int) (int, error) {
	raw := envOrDefault(name, strconv.Itoa(defaultValue))
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return value, nil
}

func durationEnv(name string, defaultValue time.Duration) (time.Duration, error) {
	raw := envOrDefault(name, defaultValue.String())
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	return value, nil
}

func positiveDurationEnv(name string, defaultValue time.Duration) (time.Duration, error) {
	value, err := durationEnv(name, defaultValue)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", name)
	}
	return value, nil
}

func Load() (Config, error) {
	dbc, err := LoadDBConfig()
	if err != nil {
		return Config{}, err
	}
	httpConfig, err := LoadHTTPConfig()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Database: dbc,
		HTTP:     httpConfig,
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

func LoadHTTPConfig() (HTTPConfig, error) {
	readHeaderTimeout, err := positiveDurationEnv("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	readTimeout, err := positiveDurationEnv("HTTP_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	writeTimeout, err := positiveDurationEnv("HTTP_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	idleTimeout, err := positiveDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	shutdownTimeout, err := positiveDurationEnv("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	drainDelay, err := durationEnv("HTTP_DRAIN_DELAY", 2*time.Second)
	if err != nil {
		return HTTPConfig{}, err
	}
	if drainDelay < 0 {
		return HTTPConfig{}, fmt.Errorf("HTTP_DRAIN_DELAY must not be negative")
	}

	return HTTPConfig{
		Port:              envOrDefault("HTTP_PORT", "8080"),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ShutdownTimeout:   shutdownTimeout,
		DrainDelay:        drainDelay,
	}, nil
}

func LoadDBConfig() (DBConfig, error) {
	maxOpenConns, err := intEnv("DB_MAX_OPEN_CONNS", 10)
	if err != nil || maxOpenConns <= 0 {
		return DBConfig{}, fmt.Errorf("DB_MAX_OPEN_CONNS must be greater than zero")
	}
	maxIdleConns, err := intEnv("DB_MAX_IDLE_CONNS", 5)
	if err != nil || maxIdleConns < 0 || maxIdleConns > maxOpenConns {
		return DBConfig{}, fmt.Errorf("DB_MAX_IDLE_CONNS must be between zero and DB_MAX_OPEN_CONNS")
	}
	connMaxLifetime, err := durationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil || connMaxLifetime <= 0 {
		return DBConfig{}, fmt.Errorf("DB_CONN_MAX_LIFETIME must be a positive duration")
	}

	dbc := DBConfig{
		Host:            envOrDefault("DB_HOST", "localhost"),
		Port:            envOrDefault("DB_PORT", "5432"),
		User:            envOrDefault("DB_USER", "user"),
		Name:            envOrDefault("DB_NAME", "blog"),
		SSLMode:         envOrDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}
	password, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return DBConfig{}, err
	}
	dbc.Password = password
	return dbc, nil
}
