package migration

import (
	"fmt"
	"go-web-socket/config"
	"go-web-socket/internal/models"
	"log"
)

func RunMigration() {
	db, err := config.GetDatabaseConnection()

	if err != nil {
		panic(err.Error())
	}

	err = db.AutoMigrate(&models.User{}, &models.Message{})
	if err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}

	sqlDB, err := db.DB()

	if err == nil {
		defer sqlDB.Close()
	}

	fmt.Println("Migration executed successfully!")
}
