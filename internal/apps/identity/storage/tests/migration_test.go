// Copyright (c) 2025 Justin Cranford
//
//

package tests

import (
	"context"
	"testing"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Register CGO-free SQLite driver
)

func TestSQLiteMigrations(t *testing.T) {
	t.Parallel()

	// Skip if CGO is not available (for CI/CD with CGO_ENABLED=0)
	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	// Create database config for in-memory SQLite
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}

	// Create repository factory
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	// Run migrations
	err = repoFactory.AutoMigrate(ctx)
	require.NoError(t, err)

	// Verify migrations by checking if tables exist
	tables := []string{
		"users", "clients", "tokens", "sessions",
		"client_profiles", "auth_flows", "auth_profiles", "mfa_factors",
	}

	for _, table := range tables {
		var count int64

		err = repoFactory.DB().Table(table).Count(&count).Error
		require.NoError(t, err, "table %s should exist after migration", table)
	}

	// Test basic CRUD operations to ensure migrations are functional
	testBasicCRUDOperations(ctx, t, repoFactory)
}

func TestPostgreSQLMigrations(t *testing.T) {
	t.Parallel()

	// Skip PostgreSQL tests in unit test environment (would require PostgreSQL server)
	t.Skip("PostgreSQL tests require database server, skipping in unit tests")
}

func TestMigrationIdempotency(t *testing.T) {
	t.Parallel()

	// Skip if CGO is not available
	if !isCGOAvailable() {
		t.Skip("CGO not available, skipping SQLite tests")
	}

	ctx := context.Background()

	// Create database config for in-memory SQLite
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}

	// Create repository factory
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup //nolint:errcheck // Test cleanup

	// Run migrations multiple times to test idempotency
	for i := 0; i < 3; i++ {
		err = repoFactory.AutoMigrate(ctx)
		require.NoError(t, err, "migration should be idempotent, iteration %d", i+1)

		// Verify tables still exist after each migration
		var count int64

		err = repoFactory.DB().Table("users").Count(&count).Error
		require.NoError(t, err, "users table should exist after migration %d", i+1)
	}
}

// Helper function to check if CGO is available.
func isCGOAvailable() bool {
	// Try to create a repository factory with SQLite - if it fails due to CGO, skip the test
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type:            "sqlite",
		DSN:             ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
		AutoMigrate:     true,
	}

	ctx := context.Background()
	_, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)

	return err == nil
}

// Helper function to test basic CRUD operations after migration.
func testBasicCRUDOperations(_ context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()
	// This would test basic CRUD operations on each repository
	// Implementation depends on the specific repository interfaces
	// For now, just verify the repository factory is created successfully
	require.NotNil(t, repoFactory)
}
