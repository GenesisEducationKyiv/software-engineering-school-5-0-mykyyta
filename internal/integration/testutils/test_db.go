package testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"weatherApi/internal/db"
)

type TestPostgres struct {
	DB        *db.DB
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

	port, err := container.MappedPort(ctx, "5432/tcp")
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
		DB:        &db.DB{Gorm: gormDB},
		Container: container,
		DSN:       dsn,
	}, nil
}

func (tp *TestPostgres) Terminate(ctx context.Context) error {
	log.Println("[Test] Terminating PostgreSQL container...")
	return tp.Container.Terminate(ctx)
}

func runMigrations(dsn string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd error: %w", err)
	}
	migrationsPath := filepath.Join(wd, "../../../migrations")

	cmd := exec.Command("migrate", "-path", migrationsPath, "-database", dsn, "up")
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("migration error: %w\nOutput: %s", err, string(output))
	}
	return nil
}
