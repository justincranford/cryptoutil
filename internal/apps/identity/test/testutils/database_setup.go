// Copyright (c) 2025 Justin Cranford
//
//

// Package testutils provides testing utilities for identity service tests.
package testutils

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

var (
	// Global sql.DB connection for shared in-memory SQLite database.
	globalSQLDB     *sql.DB
	globalSQLDBOnce sync.Once
)

// TestDatabaseConfig holds test database configuration.
type TestDatabaseConfig struct {
	DB *gorm.DB
}

// SetupTestDatabase creates or reuses a shared in-memory SQLite database for tests.
func SetupTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Initialize shared sql.DB once for all tests in this package.
	globalSQLDBOnce.Do(func() {
		// Use shared in-memory SQLite database for all tests in this package.
		dsn := "file::memory:?cache=shared"

		// Open SQLite database with modernc driver (CGO-free) explicitly.
		var err error

		globalSQLDB, err = sql.Open("sqlite", dsn)
		require.NoError(t, err, "failed to open SQLite database")

		// Enable WAL mode for better concurrency.
		_, err = globalSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
		require.NoError(t, err, "failed to enable WAL mode")

		// Set busy timeout for handling concurrent write operations.
		_, err = globalSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
		require.NoError(t, err, "failed to set busy timeout")

		// Apply SQL migrations using the repository migration system (once for shared DB).
		err = cryptoutilIdentityRepository.Migrate(globalSQLDB, "sqlite")
		if err != nil {
			t.Logf("Migration error details: %+v", err)
		}

		require.NoError(t, err, "failed to migrate test database")
	})

	// Use GORM sqlite dialector with shared sql.DB connection.
	dialector := sqlite.Dialector{Conn: globalSQLDB}

	// Open database connection with GORM (each test gets its own GORM instance).
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true, // Disable automatic transactions (we manage explicitly).
	})
	require.NoError(t, err, "failed to connect to database")

	return db
}

// CleanupTestDatabase truncates all tables to cleanup between tests using shared database.
func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Truncate tables in reverse dependency order to avoid foreign key violations.
	tables := []string{"sessions", "tokens", "clients", "users"}
	for _, table := range tables {
		_ = db.Exec("DELETE FROM " + table).Error //nolint:errcheck // Best effort cleanup
	}
}

// CreateTestConfig creates a minimal test configuration.
func CreateTestConfig(t *testing.T, authzPort, idpPort, rsPort int) *cryptoutilIdentityConfig.Config {
	t.Helper()

	return &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:         "test-authz",
			BindAddress:  "127.0.0.1",
			Port:         authzPort,
			TLSEnabled:   false,
			ReadTimeout:  cryptoutilIdentityMagic.TestReadTimeout,
			WriteTimeout: cryptoutilIdentityMagic.TestWriteTimeout,
			IdleTimeout:  cryptoutilIdentityMagic.TestIdleTimeout,
		},
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:         "test-idp",
			BindAddress:  "127.0.0.1",
			Port:         idpPort,
			TLSEnabled:   false,
			ReadTimeout:  cryptoutilIdentityMagic.TestReadTimeout,
			WriteTimeout: cryptoutilIdentityMagic.TestWriteTimeout,
			IdleTimeout:  cryptoutilIdentityMagic.TestIdleTimeout,
		},
		RS: &cryptoutilIdentityConfig.ServerConfig{
			Name:         "test-rs",
			BindAddress:  "127.0.0.1",
			Port:         rsPort,
			TLSEnabled:   false,
			ReadTimeout:  cryptoutilIdentityMagic.TestReadTimeout,
			WriteTimeout: cryptoutilIdentityMagic.TestWriteTimeout,
			IdleTimeout:  cryptoutilIdentityMagic.TestIdleTimeout,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type:        "sqlite",
			DSN:         ":memory:",
			AutoMigrate: true,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime:  time.Hour,
			RefreshTokenLifetime: cryptoutilIdentityMagic.TestRefreshTokenLifetime,
			IDTokenLifetime:      time.Hour,
			AccessTokenFormat:    "jws",
			RefreshTokenFormat:   "uuid",
			IDTokenFormat:        "jws",
		},
	}
}

// WaitForServer waits for a server to be ready by checking repeatedly.
func WaitForServer(t *testing.T, url string, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(cryptoutilIdentityMagic.TestServerWaitTickerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for server at %s", url)
		case <-ticker.C:
			// In real implementation, would make HTTP request to check readiness
			// For now, this is a placeholder for the testing infrastructure
			return
		}
	}
}
