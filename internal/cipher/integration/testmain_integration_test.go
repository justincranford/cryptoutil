// Copyright (c) 2025 Justin Cranford
//
//

//go:build !windows

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

// Shared test resources (initialized once per package).
var (
	sharedPGContainer *postgres.PostgresContainer
	sharedConnStr     string
)

// TestMain initializes shared PostgreSQL container once for all integration tests.
// This significantly reduces test execution time by amortizing container startup (3-4s)
// across all integration tests instead of per-test.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Generate random database password.
	dbPassword, err := cryptoutilRandom.GeneratePasswordSimple()
	if err != nil {
		panic(fmt.Sprintf("failed to generate database password: %v", err))
	}

	// Start PostgreSQL test-container ONCE for all integration tests.
	sharedPGContainer, err = postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(fmt.Sprintf("test_%s", googleUuid.NewString())),
		postgres.WithUsername(fmt.Sprintf("user_%s", googleUuid.NewString())),
		postgres.WithPassword(dbPassword),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start PostgreSQL container: %v", err))
	}
	defer func() {
		if err := sharedPGContainer.Terminate(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to terminate PostgreSQL container: %v\n", err)
		}
	}() // LIFO: cleanup container last.

	// Get connection string.
	connStr, err := sharedPGContainer.ConnectionString(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %v", err))
	}

	// Disable SSL for test containers (testcontainers doesn't configure SSL by default).
	if !strings.Contains(connStr, "?") {
		connStr += "?sslmode=disable"
	} else {
		connStr += "&sslmode=disable"
	}

	sharedConnStr = connStr

	// Verify connection works before running tests.
	db, err := gorm.Open(postgresDriver.Open(sharedConnStr), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to PostgreSQL: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get SQL database instance: %v", err))
	}
	defer sqlDB.Close() // LIFO: close database connection.

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Sprintf("failed to ping PostgreSQL: %v", err))
	}

	// Run all tests (defer statements execute cleanup AFTER m.Run() completes).
	exitCode := m.Run()

	os.Exit(exitCode)
}
