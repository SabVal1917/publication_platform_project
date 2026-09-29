package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB(dbc DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbc.Host, dbc.User, dbc.Password, dbc.Name, dbc.Port, dbc.SSLMode)

	fmt.Println("Connecting to DB at:", dbc.Host)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Printf("Ошибка открытия БД: %v \n", err)
		return nil, err
	}

	if err := db.Ping(); err != nil {
		log.Printf("Ошибка подключения к БД (Ping): %v \n", err)
		return nil, err
	}

	log.Println("Успешное подключение к БД!")
	return db, nil
}
