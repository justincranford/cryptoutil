// Copyright (c) 2025 Justin Cranford
//
//

package testutils

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TestDatabaseConfig holds test database configuration.
type TestDatabaseConfig struct {
	DB *gorm.DB
}

// SetupTestDatabase creates an in-memory SQLite database for testing.
func SetupTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Use shared in-memory SQLite database with modernc driver (CGO-free).
	dsn := "file::memory:?cache=shared"

	// Open SQLite database with modernc driver (CGO-free) explicitly.
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open SQLite database")

	// Enable WAL mode for better concurrency.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err, "failed to enable WAL mode")

	// Set busy timeout for handling concurrent write operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err, "failed to set busy timeout")

	// Use GORM sqlite dialector with existing sql.DB connection from modernc driver.
	dialector := sqlite.Dialector{Conn: sqlDB}

	// Open database connection with GORM.
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true, // Disable automatic transactions (we manage explicitly).
	})
	require.NoError(t, err, "failed to connect to database")

	// Apply SQL migrations using the repository migration system.
	err = cryptoutilIdentityRepository.Migrate(sqlDB)
	require.NoError(t, err, "failed to migrate test database")

	return db
}

// CleanupTestDatabase closes the test database connection.
func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close() //nolint:errcheck // Test helper cleanup
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
