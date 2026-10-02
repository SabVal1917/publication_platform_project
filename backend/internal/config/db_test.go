package config

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestConfigureDBPool(t *testing.T) {
	db, err := sql.Open("pgx", "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	cfg := DBConfig{
		MaxOpenConns:    17,
		MaxIdleConns:    6,
		ConnMaxLifetime: 25 * time.Minute,
	}
	ConfigureDB(db, cfg)

	stats := db.Stats()
	if stats.MaxOpenConnections != cfg.MaxOpenConns {
		t.Fatalf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, cfg.MaxOpenConns)
	}
}
