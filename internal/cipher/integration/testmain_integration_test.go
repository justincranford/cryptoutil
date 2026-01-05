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

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"

	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Shared test resources (initialized once per package).
var (
	sharedPGContainer *postgres.PostgresContainer
	sharedConnStr     string
	sharedDB          *gorm.DB
	sharedServer      *server.CipherIMServer
)

// TestMain initializes shared PostgreSQL container and server once for all integration tests.
// This significantly reduces test execution time by amortizing container startup (3-4s)
// across all integration tests instead of per-test.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup shared PostgreSQL container using utility function.
	var err error
	sharedPGContainer, sharedConnStr, err = SetupSharedPostgresContainer(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to setup PostgreSQL container: %v", err))
	}
	defer func() {
		if err := sharedPGContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate PostgreSQL container: %v\n", err)
		}
	}() // LIFO: cleanup container last.

	// Verify connection works before running tests.
	if err := VerifyPostgresConnection(sharedConnStr); err != nil {
		panic(fmt.Sprintf("failed to verify PostgreSQL connection: %v", err))
	}

	// Create shared database connection.
	sharedDB, err = gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %v", err))
	}

	// Create shared server instance (applies migrations via repository.ApplyMigrations).
	cfg := NewTestConfig("cipher-im-integration")
	sharedServer, err = server.New(ctx, cfg, sharedDB, repository.DatabaseTypePostgreSQL)
	if err != nil {
		panic(fmt.Sprintf("failed to create server: %v", err))
	}

	// Run all tests (defer statements execute cleanup AFTER m.Run() completes).
	exitCode := m.Run()

	os.Exit(exitCode)
}
