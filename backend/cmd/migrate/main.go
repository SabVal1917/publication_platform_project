package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"backend/internal/config"
)

const defaultMigrationsURL = "file:///app/migrations"

func main() {
	if err := run(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func run() error {
	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		return fmt.Errorf("load database config: %w", err)
	}

	databaseURL := postgresURL(dbConfig)
	migrationsURL := os.Getenv("MIGRATIONS_URL")
	if migrationsURL == "" {
		migrationsURL = defaultMigrationsURL
	}

	migrator, err := migrate.New(migrationsURL, databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := migrator.Close()
		if sourceErr != nil {
			log.Printf("close migration source: %v", sourceErr)
		}
		if databaseErr != nil {
			log.Printf("close migration database: %v", databaseErr)
		}
	}()

	if err := migrator.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("database schema is up to date")
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	log.Println("database migrations applied successfully")
	return nil
}

func postgresURL(cfg config.DBConfig) string {
	query := url.Values{}
	query.Set("sslmode", cfg.SSLMode)

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: query.Encode(),
	}).String()
}
