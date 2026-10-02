package config

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB(ctx context.Context, dbc DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbc.Host, dbc.User, dbc.Password, dbc.Name, dbc.Port, dbc.SSLMode)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	ConfigureDB(db, dbc)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func ConfigureDB(db *sql.DB, dbc DBConfig) {
	db.SetMaxOpenConns(dbc.MaxOpenConns)
	db.SetMaxIdleConns(dbc.MaxIdleConns)
	db.SetConnMaxLifetime(dbc.ConnMaxLifetime)
}
