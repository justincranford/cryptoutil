// Copyright (c) 2025-2026 Justin Cranford.
//
//

package container

import (
	"context"
	"database/sql"
	"fmt"
	"net"
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
	return startPostgres(ctx, telemetryService, dbName, dbUsername, dbPassword, StartContainer, GetContainerHostAndMappedPort)
}

func startPostgres(
	ctx context.Context,
	telemetryService *cryptoutilSharedTelemetry.TelemetryService,
	dbName, dbUsername, dbPassword string,
	startContainerFn func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, testcontainers.ContainerRequest) (testcontainers.Container, func(), error),
	getContainerHostAndMappedPortFn func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, testcontainers.Container, string) (string, string, error),
) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUsername, "POSTGRES_PASSWORD": dbPassword},
		// WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(postgresContainerStartupTimeout),
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(cryptoutilSharedMagic.DBPostgresContainerStartupTimeout),
	}

	postgresContainer, terminateContainer, err := startContainerFn(ctx, telemetryService, postgresContainerRequest)
	if err != nil {
		telemetryService.Slogger.Error("failed to start postgres container", cryptoutilSharedMagic.StringError, err)

		return "", nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	containerHost, containerMappedPort, err := getContainerHostAndMappedPortFn(ctx, telemetryService, postgresContainer, "5432")
	if err != nil {
		telemetryService.Slogger.Error("failed to get postgres container host and mapped port", cryptoutilSharedMagic.StringError, err)
		terminateContainer()

		return "", nil, fmt.Errorf("failed to get postgres container host and mapped port: %w", err)
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUsername, dbPassword, net.JoinHostPort(containerHost, containerMappedPort), dbName)

	return databaseURL, terminateContainer, nil
}

// SetupSharedPostgresContainer initializes a PostgreSQL test-container for integration tests.
// Returns the container and connection string for reuse across tests.
//
// This significantly reduces test execution time by amortizing container startup (3-4s)
// across all integration tests instead of per-test.
func SetupSharedPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	return setupSharedPostgresContainer(ctx, cryptoutilSharedUtilRandom.GeneratePasswordSimple, postgres.Run, postgresContainerConnectionString)
}

func postgresContainerConnectionString(ctx context.Context, container *postgres.PostgresContainer) (string, error) {
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get PostgreSQL container connection string: %w", err)
	}

	return connectionString, nil
}

func setupSharedPostgresContainer(
	ctx context.Context,
	generatePasswordFn func() (string, error),
	runPostgresFn func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error),
	connectionStringFn func(context.Context, *postgres.PostgresContainer) (string, error),
) (*postgres.PostgresContainer, string, error) {
	// Generate random database password.
	dbPassword, err := generatePasswordFn()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate database password: %w", err)
	}

	// Start PostgreSQL test-container ONCE for all integration tests.
	container, err := runPostgresFn(ctx,
		"postgres:latest",
		postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewString())),
		postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewString())),
		postgres.WithPassword(dbPassword),
	)
	if err != nil {
		startErr := fmt.Errorf("failed to start PostgreSQL container: %w", err)
		if container != nil {
			startErr = cleanupTestContainerAfterError(ctx, container, startErr, "failed to clean up PostgreSQL container after startup failure")
		}

		return nil, "", startErr
	}

	// Get connection string.
	connStr, err := connectionStringFn(ctx, container)
	if err != nil {
		connectionStringErr := fmt.Errorf("failed to get connection string: %w", err)
		connectionStringErr = cleanupTestContainerAfterError(ctx, container, connectionStringErr, "failed to clean up PostgreSQL container after connection string failure")

		return nil, "", connectionStringErr
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
	return verifyPostgresConnectionWithFns(connStr,
		func(dsn string) (*gorm.DB, error) {
			return gorm.Open(postgresDriver.Open(dsn), &gorm.Config{})
		},
		getSQLDBFromGORM,
		pingPostgresSQLDB,
	)
}

func getSQLDBFromGORM(db *gorm.DB) (*sql.DB, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL database from GORM: %w", err)
	}

	return sqlDB, nil
}

func pingPostgresSQLDB(sqlDB *sql.DB) error {
	if err := sqlDB.PingContext(context.Background()); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return nil
}

func verifyPostgresConnectionWithFns(
	connStr string,
	openDBFn func(string) (*gorm.DB, error),
	getSQLDBFn func(*gorm.DB) (*sql.DB, error),
	pingFn func(*sql.DB) error,
) error {
	db, err := openDBFn(connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := getSQLDBFn(db)
	if err != nil {
		return fmt.Errorf("failed to get SQL database instance: %w", err)
	}
	defer sqlDB.Close() //nolint:errcheck

	if err := pingFn(sqlDB); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return nil
}
