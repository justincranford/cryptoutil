// Copyright (c) 2025 Justin Cranford
//
//

package container

import (
	"context"
	"fmt"
	"strings"

	googleUuid "github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"

	postgresDriver "gorm.io/driver/postgres"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// StartPostgres starts a PostgreSQL test container and returns the DSN connection string.
func StartPostgres(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, dbName, dbUsername, dbPassword string) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:18",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUsername, "POSTGRES_PASSWORD": dbPassword},
		// WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(postgresContainerStartupTimeout),
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(cryptoutilSharedMagic.DBPostgresContainerStartupTimeout),
	}

	postgresContainer, terminateContainer, err := StartContainer(ctx, telemetryService, postgresContainerRequest)
	if err != nil {
		telemetryService.Slogger.Error("failed to start postgres container", "error", err)

		return "", nil, fmt.Errorf("failed to start sqlite container: %w", err)
	}

	containerHost, containerMappedPort, err := GetContainerHostAndMappedPort(ctx, telemetryService, postgresContainer, "5432")
	if err != nil {
		telemetryService.Slogger.Error("failed to get postgres container host and mapped port", "error", err)
		terminateContainer()

		return "", nil, fmt.Errorf("failed to get postgres container host and mapped port: %w", err)
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, containerHost, containerMappedPort, dbName)

	return databaseURL, terminateContainer, nil
}

// SetupSharedPostgresContainer initializes a PostgreSQL test-container for integration tests.
// Returns the container and connection string for reuse across tests.
//
// This significantly reduces test execution time by amortizing container startup (3-4s)
// across all integration tests instead of per-test.
func SetupSharedPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	// Generate random database password.
	dbPassword, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate database password: %w", err)
	}

	// Start PostgreSQL test-container ONCE for all integration tests.
	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewString())),
		postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewString())),
		postgres.WithPassword(dbPassword),
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get connection string.
	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get connection string: %w", err)
	}

	// Disable SSL for test containers (testcontainers doesn't configure SSL by default).
	if !strings.Contains(connStr, "?") {
		connStr += "?sslmode=disable"
	} else {
		connStr += "&sslmode=disable"
	}

	return container, connStr, nil
}

// VerifyPostgresConnection verifies the PostgreSQL connection works.
// Returns error if connection or ping fails.
func VerifyPostgresConnection(connStr string) error {
	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL database instance: %w", err)
	}
	defer sqlDB.Close() //nolint:errcheck

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return nil
}
