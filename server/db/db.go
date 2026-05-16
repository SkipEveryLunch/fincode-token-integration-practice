package db

import (
	infrapostgres "fincode-token-practice/server/infrastructure/postgres"
	"log"
	"os"

	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = gorm.Open(gormpg.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(
		&infrapostgres.Customer{},
		&infrapostgres.Card{},
		&infrapostgres.Payment{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	log.Println("database connected and migrated")
}
