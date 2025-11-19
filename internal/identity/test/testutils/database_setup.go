// Copyright (c) 2025 Justin Cranford
//
//

package testutils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// TestDatabaseConfig holds test database configuration.
type TestDatabaseConfig struct {
	DB *gorm.DB
}

// SetupTestDatabase creates an in-memory SQLite database for testing.
func SetupTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	// Create in-memory SQLite database.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		// If this is a CGO error, skip the test instead of failing.
		if err.Error() == "Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub" {
			t.Skip("Skipping database tests: CGO not available (required for SQLite)")
		}

		require.NoError(t, err, "failed to open test database")
	}

	// Auto-migrate all domain models.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientProfile{},
		&cryptoutilIdentityDomain.AuthFlow{},
		&cryptoutilIdentityDomain.Token{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.AuthProfile{},
		&cryptoutilIdentityDomain.MFAFactor{},
	)
	require.NoError(t, err, "failed to migrate test database")

	return db
}

// CleanupTestDatabase closes the test database connection.
func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
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
