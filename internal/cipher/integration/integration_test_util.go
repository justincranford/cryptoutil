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

	"cryptoutil/internal/cipher/server/config"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
	cryptoutilE2E "cryptoutil/internal/template/testing/e2e"
)

// NewTestConfig returns an AppConfig with test-friendly settings.
// Reuses template helper for consistent ServerSettings across all cipher tests.
func NewTestConfig(serviceName string) *config.AppConfig {
	cfg := config.DefaultAppConfig()

	// Override with test-specific settings using template helper.
	serverSettings := cryptoutilE2E.NewTestServerSettingsWithService(serviceName)
	cfg.ServerSettings = *serverSettings

	return cfg
}

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
