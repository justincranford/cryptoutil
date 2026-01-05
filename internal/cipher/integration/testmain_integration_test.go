// Copyright (c) 2025 Justin Cranford
//
//

//go:build !windows

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	"cryptoutil/internal/shared/container"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

// Shared test resources (initialized once per package).
var (
	sharedPGContainer *postgres.PostgresContainer
	sharedConnStr     string
	sharedDB          *gorm.DB
	sharedServer      *server.CipherIMServer
)

// TestMain initializes shared PostgreSQL container and full cipher-im server once for all integration tests.
// This significantly reduces test execution time by amortizing container startup (3-4s)
// and server initialization across all integration tests instead of per-test.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup shared PostgreSQL container using utility function.
	var err error
	sharedPGContainer, sharedConnStr, err = container.SetupSharedPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to setup PostgreSQL container: %v", err))
	}
	defer func() {
		if err := sharedPGContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate PostgreSQL container: %v\n", err)
		}
	}() // LIFO: cleanup container last.

	// Verify connection works before running tests.
	if err := container.VerifyPostgresConnection(sharedConnStr); err != nil {
		panic(fmt.Sprintf("failed to verify PostgreSQL connection: %v", err))
	}

	// Create shared database connection.
	sharedDB, err = gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %v", err))
	}

	// Create full cipher-im server instance (applies migrations, initializes all services).
	// Uses utility function for consistency.
	sharedServer, err = InitSharedCipherIMServer(ctx, sharedDB)
	if err != nil {
		panic(fmt.Sprintf("failed to create cipher-im server: %v", err))
	}

	// Run all tests (defer statements execute cleanup AFTER m.Run() completes).
	exitCode := m.Run()

	os.Exit(exitCode)
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
