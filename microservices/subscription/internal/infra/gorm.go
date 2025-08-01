package infra

import (
	"context"
	"fmt"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

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

func (db *Gorm) Close(ctx context.Context) {
	logger := loggerPkg.From(ctx)
	if sqlDB, err := db.Gorm.DB(); err != nil {
		logger.Error("failed to get sql.DB: %v", err)
	} else if err := sqlDB.Close(); err != nil {
		logger.Error("failed to close DB: %v", err)
	} else {
		logger.Info("DB connection closed")
	}
}
