package db

import (
	"log"
	"os"
	"path/filepath"

	"greenlabelai/backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) *gorm.DB {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(
		&models.Product{},
		&models.SearchHistory{},
		&models.UserGoal{},
		&models.ShoppingBasket{},
		&models.BasketItem{},
		&models.Badge{},
		&models.UserBadge{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized and migrated")
	return db
}
