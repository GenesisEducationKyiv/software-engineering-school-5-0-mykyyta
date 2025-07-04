package infra

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Gorm struct {
	Gorm *gorm.DB
}

func NewGorm(dsn string) (*Gorm, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &Gorm{Gorm: db}, nil
}

func (db *Gorm) Close() {
	if sqlDB, err := db.Gorm.DB(); err != nil {
		log.Printf("failed to get sql.DB: %v", err)
	} else if err := sqlDB.Close(); err != nil {
		log.Printf("failed to close DB: %v", err)
	} else {
		log.Println("DB connection closed")
	}
}
