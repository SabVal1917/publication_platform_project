package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDBConfigPoolDefaults(t *testing.T) {
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("DB_MAX_OPEN_CONNS", "")
	t.Setenv("DB_MAX_IDLE_CONNS", "")
	t.Setenv("DB_CONN_MAX_LIFETIME", "")

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Fatalf("LoadDBConfig() error = %v", err)
	}

	if cfg.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 30*time.Minute {
		t.Errorf("ConnMaxLifetime = %s, want 30m", cfg.ConnMaxLifetime)
	}
}

func TestLoadDBConfigPoolFromEnvironment(t *testing.T) {
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("DB_MAX_OPEN_CONNS", "24")
	t.Setenv("DB_MAX_IDLE_CONNS", "8")
	t.Setenv("DB_CONN_MAX_LIFETIME", "45m")

	cfg, err := LoadDBConfig()
	if err != nil {
		t.Fatalf("LoadDBConfig() error = %v", err)
	}

	if cfg.MaxOpenConns != 24 || cfg.MaxIdleConns != 8 || cfg.ConnMaxLifetime != 45*time.Minute {
		t.Fatalf("unexpected pool config: %+v", cfg)
	}
}

func TestLoadDBConfigRejectsInvalidPool(t *testing.T) {
	tests := []struct {
		name      string
		maxOpen   string
		maxIdle   string
		lifetime  string
		wantError string
	}{
		{name: "invalid max open", maxOpen: "many", maxIdle: "5", lifetime: "30m", wantError: "DB_MAX_OPEN_CONNS"},
		{name: "non-positive max open", maxOpen: "0", maxIdle: "0", lifetime: "30m", wantError: "DB_MAX_OPEN_CONNS"},
		{name: "negative max idle", maxOpen: "10", maxIdle: "-1", lifetime: "30m", wantError: "DB_MAX_IDLE_CONNS"},
		{name: "idle exceeds open", maxOpen: "3", maxIdle: "4", lifetime: "30m", wantError: "DB_MAX_IDLE_CONNS"},
		{name: "invalid lifetime", maxOpen: "10", maxIdle: "5", lifetime: "forever", wantError: "DB_CONN_MAX_LIFETIME"},
		{name: "non-positive lifetime", maxOpen: "10", maxIdle: "5", lifetime: "0s", wantError: "DB_CONN_MAX_LIFETIME"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_PASSWORD", "password")
			t.Setenv("DB_MAX_OPEN_CONNS", tt.maxOpen)
			t.Setenv("DB_MAX_IDLE_CONNS", tt.maxIdle)
			t.Setenv("DB_CONN_MAX_LIFETIME", tt.lifetime)

			_, err := LoadDBConfig()
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("LoadDBConfig() error = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestLoadHTTPConfigFromEnvironment(t *testing.T) {
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("JWT_SECRET", "test-secret-at-least-thirty-two-bytes")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("HTTP_READ_HEADER_TIMEOUT", "3s")
	t.Setenv("HTTP_READ_TIMEOUT", "12s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "45s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "9s")
	t.Setenv("HTTP_DRAIN_DELAY", "1500ms")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := HTTPConfig{
		Port:              "9090",
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       12 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       45 * time.Second,
		ShutdownTimeout:   9 * time.Second,
		DrainDelay:        1500 * time.Millisecond,
	}
	if cfg.HTTP != want {
		t.Fatalf("HTTP config = %+v, want %+v", cfg.HTTP, want)
	}
}

func TestLoadRejectsInvalidHTTPTimeout(t *testing.T) {
	t.Setenv("DB_PASSWORD", "password")
	t.Setenv("JWT_SECRET", "test-secret-at-least-thirty-two-bytes")
	t.Setenv("HTTP_READ_TIMEOUT", "0s")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "HTTP_READ_TIMEOUT") {
		t.Fatalf("Load() error = %v, want HTTP_READ_TIMEOUT error", err)
	}
}

func TestLoadRequiresStrongJWTSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{name: "missing", secret: ""},
		{name: "too short", secret: "short-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_PASSWORD", "password")
			t.Setenv("JWT_SECRET", tt.secret)

			_, err := Load()
			if err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
				t.Fatalf("Load() error = %v, want JWT_SECRET error", err)
			}
		})
	}
}
