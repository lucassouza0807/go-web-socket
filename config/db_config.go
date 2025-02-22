package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetDatabaseConnection() (*gorm.DB, error) {
	err := godotenv.Load()

	if err != nil {
		log.Printf("Erro ao carregar o arquivo .env: %v", err)
		return nil, fmt.Errorf("erro ao carregar o arquivo .env: %w", err)
	}

	connection := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(connection), &gorm.Config{})

	if err != nil {
		log.Printf("Erro ao conectar ao banco de dados: %v", err)
		return nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}

	return db, nil

}
