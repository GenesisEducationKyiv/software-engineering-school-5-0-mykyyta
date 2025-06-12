package db

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"weatherApi/config"
)

func InitDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable pgcrypto extension for UUID support
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error; err != nil {
		return nil, fmt.Errorf("failed to enable pgcrypto: %w", err)
	}

	return db, nil
}

func ConnectDefaultDB() *gorm.DB {
	dsn := config.C.DBUrl

	db, err := InitDatabase(dsn)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	fmt.Println("Connected to PostgreSQL")
	return db
}

func CloseDB(db *gorm.DB) {
	if db == nil {
		log.Println("CloseDB: no DB to close (db is nil)")
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Could not get sql.DB from GORM: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close DB: %v", err)
	} else {
		log.Println("DB connection closed")
	}
}
