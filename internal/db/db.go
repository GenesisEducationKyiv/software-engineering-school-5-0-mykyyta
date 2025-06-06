package db

import (
	"fmt"
	"log"

	"weatherApi/config"
	"weatherApi/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the globally accessible database instance used across the application.
// It is initialized once via ConnectDefaultDB or manually through InitDatabase.
var DB *gorm.DB

// InitDatabase initializes and returns a GORM DB connection based on the dbType and DSN provided.
// Supports "postgres" and "sqlite". Also handles schema migration and optional Postgres extension.
// This function should typically be called once at startup.
func InitDatabase(dbType, dsn string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch dbType {
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Enable pgcrypto (required for UUID generation, etc.)
	if dbType == "postgres" {
		err = db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`).Error
		if err != nil {
			return nil, fmt.Errorf("failed to enable pgcrypto: %w", err)
		}
	}

	// Run automatic schema migration for Subscription model
	err = db.AutoMigrate(&model.Subscription{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return db, nil
}

// ConnectDefaultDB reads DB_TYPE and DB_URL from environment variables,
// initializes the global DB instance, and applies migrations.
// Use this in main.go to ensure the DB is ready before handling requests.
func ConnectDefaultDB() {
	dbType := config.C.DBType // e.g., "postgres" or "sqlite"
	dsn := config.C.DBUrl

	var err error
	DB, err = InitDatabase(dbType, dsn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to DB and ran AutoMigrate")
}
