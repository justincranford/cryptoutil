// Copyright (c) 2025 Justin Cranford

// Package e2e_test provides cipher-specific E2E test utilities.
package e2e_test

import (
	"gorm.io/gorm"

	"cryptoutil/internal/cipher/integration"
	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/cipher/server/config"
)

// newTestAppConfig creates an AppConfig with test-friendly settings.
// Delegates to exported integration.NewTestAppConfig for consistency.
func newTestAppConfig(serviceName, jwtSecret string) *config.AppConfig {
	return integration.NewTestAppConfig(serviceName, jwtSecret)
}

// initTestDB creates an in-memory SQLite database with cipher schema.
// Delegates to exported integration.InitTestDB for consistency.
func initTestDB() (*gorm.DB, error) {
	return integration.InitTestDB()
}

// createTestCipherIMServerInternal creates a full CipherIMServer for testing.
// Delegates to exported integration.CreateTestCipherIMServerInternal for consistency.
func createTestCipherIMServerInternal(db *gorm.DB, cfg *config.AppConfig) (*server.CipherIMServer, string, string, error) {
	return integration.CreateTestCipherIMServerInternal(db, cfg, repository.DatabaseTypeSQLite)
}
