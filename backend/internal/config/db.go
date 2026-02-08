package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost" 
	}
	
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "user"
	}
	
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "blog"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", 
		host, user, password, dbname)

	fmt.Println("Connecting to DB at:", host) 
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Ошибка открытия БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Ошибка подключения к БД (Ping): %v", err)
	}

	log.Println("Успешное подключение к БД!")
	return db
}