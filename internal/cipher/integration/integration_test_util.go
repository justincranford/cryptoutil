// Copyright (c) 2025 Justin Cranford

//go:build !windows

package integration

import (
	"context"
	"fmt"
	"strings"

	googleUuid "github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

// SetupSharedPostgresContainer initializes a PostgreSQL test-container for integration tests.
// Returns the container and connection string for reuse across tests.
//
// This significantly reduces test execution time by amortizing container startup (3-4s)
// across all integration tests instead of per-test.
func SetupSharedPostgresContainer(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	// Generate random database password.
	dbPassword, err := cryptoutilRandom.GeneratePasswordSimple()
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
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	return nil
}

// InitSharedCipherIMServer creates a full CipherIMServer with PostgreSQL for integration tests.
// This should be called from TestMain to amortize server startup cost across all tests.
func InitSharedCipherIMServer(ctx context.Context, db *gorm.DB) (*server.CipherIMServer, error) {
	cfg := NewTestConfig("cipher-im-integration")

	// Create full server instance (applies migrations via repository.ApplyMigrations).
	cipherServer, err := server.New(ctx, cfg, db, repository.DatabaseTypePostgreSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher server: %w", err)
	}

	return cipherServer, nil
}

// RunTestMainSetup performs standard TestMain setup for cipher integration tests with PostgreSQL.
// Returns the initialized server and cleanup function.
//
// This is the reusable core of TestMain logic extracted for consistency.
// Caller should invoke cleanup via defer and call os.Exit(m.Run()).
func RunTestMainSetup(ctx context.Context) (*server.CipherIMServer, *gorm.DB, func(), error) {
	// PostgreSQL test-container setup.
	container, connStr, err := SetupSharedPostgresContainer(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup PostgreSQL container: %w", err)
	}

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	if err := VerifyPostgresConnection(connStr); err != nil {
		cleanup()
		return nil, nil, nil, fmt.Errorf("failed to verify PostgreSQL connection: %w", err)
	}

	db, err := gorm.Open(postgresDriver.Open(connStr), &gorm.Config{})
	if err != nil {
		cleanup()
		return nil, nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Create full cipher-im server.
	srv, err := InitSharedCipherIMServer(ctx, db)
	if err != nil {
		cleanup()
		return nil, nil, nil, fmt.Errorf("failed to create cipher server: %w", err)
	}

	return srv, db, cleanup, nil
}
