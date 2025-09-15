package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var DB *sql.DB

func Init() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден")
	}

	// Подключение к MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Panic("Ошибка подключения к MySQL:", err)
	}

	if err = DB.Ping(); err != nil {
		log.Panic("Ошибка ping MySQL:", err)
	}

	log.Println("Успешное подключение к MySQL")
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
