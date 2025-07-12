package testutils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"
	"weatherApi/monolith/internal/infra"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestPostgres struct {
	DB        *infra.Gorm
	Container testcontainers.Container
	DSN       string
}

func StartPostgres(ctx context.Context) (*TestPostgres, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, nat.Port("5432/tcp"))
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	time.Sleep(2 * time.Second)

	gormDB, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect via gorm: %w", err)
	}

	log.Println("[Test] Running SQL migrations...")
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}

	return &TestPostgres{
		DB:        &infra.Gorm{Gorm: gormDB},
		Container: container,
		DSN:       dsn,
	}, nil
}

func (tp *TestPostgres) Terminate(ctx context.Context) error {
	log.Println("[Test] Terminating PostgreSQL container...")
	return tp.Container.Terminate(ctx)
}

func runMigrations(dsn string) error {
	absPath, err := filepath.Abs("../../../migrations")
	if err != nil {
		return fmt.Errorf("could not resolve migrations path: %w", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open failed: %w", err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			log.Printf("warning: failed to close db: %v", cerr)
		}
	}()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("postgres driver init failed: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+absPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("migrate init failed: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up failed: %w", err)
	}

	return nil
}
