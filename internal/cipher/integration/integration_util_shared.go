// Copyright (c) 2025 Justin Cranford

// Package integration provides shared test utilities for cipher integration and E2E tests.
// This file contains platform-independent utilities (no build constraints).
package integration

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
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

// NewTestAppConfig creates an AppConfig with test-friendly settings and JWT secret.
// Exported version for reuse in both E2E and integration tests.
func NewTestAppConfig(serviceName, jwtSecret string) *config.AppConfig {
	return &config.AppConfig{
		ServerSettings: *cryptoutilE2E.NewTestServerSettingsWithService(serviceName),
		JWTSecret:      jwtSecret,
	}
}

// InitTestDB creates an in-memory SQLite database with cipher schema.
// Exported version for reuse in E2E tests.
func InitTestDB() (*gorm.DB, error) {
	ctx := context.Background()

	applyMigrations := func(sqlDB *sql.DB) error {
		return repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	}

	return cryptoutilE2E.InitTestDB(ctx, applyMigrations)
}

// CreateTestCipherIMServerInternal creates a full CipherIMServer with custom config for testing.
// Returns the server instance, public URL, and admin URL.
//
// This helper encapsulates:
// - Server creation with test config
// - Background startup
// - Port allocation waiting
// - URL construction
//
// Exported version for reuse across test packages.
func CreateTestCipherIMServerInternal(db *gorm.DB, cfg *config.AppConfig, dbType repository.DatabaseType) (*server.CipherIMServer, string, string, error) {
	ctx := context.Background()

	// Create full server.
	cipherServer, err := server.New(ctx, cfg, db, dbType)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create cipher server: %w", err)
	}

	// Start server in background.
	errChan := cryptoutilE2E.StartServerAsync(ctx, cipherServer)

	// Wait for public server to bind.
	publicURL, err := cryptoutilE2E.WaitForServerPort(cipherServer, cryptoutilE2E.DefaultServerWaitParams())
	if err != nil {
		return nil, "", "", fmt.Errorf("public server failed to bind: %w", err)
	}

	// Wait for admin server to bind (cipher has separate admin port).
	adminPort, _ := cipherServer.AdminPort()
	if adminPort == 0 {
		// Admin server might still be binding, wait a bit more.
		const maxAdminWaitAttempts = 50
		for i := 0; i < maxAdminWaitAttempts; i++ {
			adminPort, _ = cipherServer.AdminPort()
			if adminPort > 0 {
				break
			}

			select {
			case err := <-errChan:
				return nil, "", "", fmt.Errorf("server startup error: %w", err)
			default:
			}
		}

		if adminPort == 0 {
			return nil, "", "", fmt.Errorf("admin server did not bind to port")
		}
	}

	adminURL := fmt.Sprintf("https://%s:%d", cfg.BindPrivateAddress, adminPort)

	return cipherServer, publicURL, adminURL, nil
}

// CreateTestCipherIMServer creates a full CipherIMServer for testing with default SQLite config.
// Returns the server instance, public URL, and admin URL.
//
// Exported version for reuse in E2E tests.
func CreateTestCipherIMServer(db *gorm.DB, jwtSecret string) (*server.CipherIMServer, string, string, error) {
	cfg := NewTestAppConfig("cipher-im-e2e-test", jwtSecret)
	return CreateTestCipherIMServerInternal(db, cfg, repository.DatabaseTypeSQLite)
}
