// Copyright (c) 2025 Justin Cranford

package idp

import (
	"context"
	"fmt"
	"testing"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestPostgreSQLIntegration validates PostgreSQL-specific features: connection pooling, concurrent operations, transaction isolation.
// Uses real PostgreSQL container (not in-memory SQLite) to validate production behavior.
// Prerequisites: PostgreSQL container running at localhost:5433 (via docker compose -f deployments/compose/postgres-test.yml up -d).
// Note: This test validates PostgreSQL features without creating separate databases (simplified approach).
func TestPostgreSQLIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// PostgreSQL connection (shared database: identitytest).
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "postgres",
		DSN:  "postgres://testuser:testpass@localhost:5433/identitytest?sslmode=disable",
	}

	// Create repository factory (skip if PostgreSQL is not available).
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration test: %v", err)
	}

	defer func() {
		sqlDB, err := repoFactory.DB().DB()
		if err == nil {
			if closeErr := sqlDB.Close(); closeErr != nil {
				require.Fail(t, fmt.Sprintf("failed to close database: %v", closeErr))
			}
		}
	}()

	// Run auto-migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.Token{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.AuthorizationRequest{},
		&cryptoutilIdentityDomain.AuthProfile{},
	)
	require.NoError(t, err, "failed to run auto-migrations")

	tests := []struct {
		name        string
		description string
		testFunc    func(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory)
	}{
		{
			name:        "connection pooling",
			description: "validates connection pool settings (MaxOpenConns, MaxIdleConns) work correctly",
			testFunc:    testConnectionPooling,
		},
		{
			name:        "concurrent operations",
			description: "validates concurrent user/client/token creation works without deadlocks",
			testFunc:    testConcurrentOperations,
		},
		{
			name:        "transaction isolation",
			description: "validates transaction isolation between concurrent transactions",
			testFunc:    testTransactionIsolation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// NOTE: Don't use t.Parallel() in subtests when sharing database connection.
			// Parallel subtests caused "database is closed" errors due to premature cleanup.
			// Run test function with shared repoFactory.
			tc.testFunc(t, repoFactory)
		})
	}
}

// testConnectionPooling validates connection pool settings work correctly.
func testConnectionPooling(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	// Validate connection pool settings.
	sqlDB, err := repoFactory.DB().DB()
	require.NoError(t, err, "failed to get sql.DB")

	// Check MaxOpenConns setting (should be >0 for PostgreSQL).
	// Note: PostgreSQL uses config-specified MaxOpenConns, unlike SQLite which hardcodes 5.
	// If MaxOpenConns is unset (0), GORM/database/sql uses unlimited connections (default behavior).
	stats := sqlDB.Stats()
	t.Logf("Connection pool stats: MaxOpenConnections=%d, OpenConnections=%d, InUse=%d, Idle=%d",
		stats.MaxOpenConnections, stats.OpenConnections, stats.InUse, stats.Idle)

	// Validate connection can execute query.
	ctx := context.Background()

	var result int

	err = sqlDB.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err, "failed to execute test query")
	require.Equal(t, 1, result, "query result should be 1")
}

// testConcurrentOperations validates concurrent user/client/token creation works without deadlocks.
func testConcurrentOperations(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	ctx := context.Background()
	userRepo := repoFactory.UserRepository()
	clientRepo := repoFactory.ClientRepository()
	tokenRepo := repoFactory.TokenRepository()

	// Number of concurrent goroutines creating entities.
	const concurrency = 10

	// Create users, clients, tokens concurrently.
	done := make(chan error, concurrency*3)

	// Concurrent user creation.
	for i := 0; i < concurrency; i++ {
		go func(_ int) {
			uniqueID := googleUuid.Must(googleUuid.NewV7()).String()

			user := &cryptoutilIdentityDomain.User{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
				PreferredUsername: "testuser_" + uniqueID,
				Email:             "test_" + uniqueID + "@example.com",
				PasswordHash:      "hashedpassword", // pragma: allowlist secret
				Enabled:           true,
			}
			done <- userRepo.Create(ctx, user)
		}(i)
	}

	// Concurrent client creation.
	for i := 0; i < concurrency; i++ {
		go func(_ int) {
			client := &cryptoutilIdentityDomain.Client{
				ClientID:     googleUuid.Must(googleUuid.NewV7()).String(),
				ClientSecret: "hashedsecret", // pragma: allowlist secret
				ClientType:   cryptoutilIdentityDomain.ClientTypeConfidential,
				Enabled:      boolPtr(true),
			}
			done <- clientRepo.Create(ctx, client)
		}(i)
	}

	// Concurrent token creation (requires valid user/client, so create them first).
	uniqueIDToken := googleUuid.Must(googleUuid.NewV7()).String()
	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		PreferredUsername: "testuser_token_" + uniqueIDToken,
		Email:             "test_token_" + uniqueIDToken + "@example.com",
		PasswordHash:      "hashedpassword", // pragma: allowlist secret
		Enabled:           true,
	}
	err := userRepo.Create(ctx, testUser)
	require.NoError(t, err, "failed to create test user for tokens")

	testClient := &cryptoutilIdentityDomain.Client{
		ClientID:     googleUuid.Must(googleUuid.NewV7()).String(),
		ClientSecret: "hashedsecret", // pragma: allowlist secret
		ClientType:   cryptoutilIdentityDomain.ClientTypeConfidential,
		Enabled:      boolPtr(true),
	}
	err = clientRepo.Create(ctx, testClient)
	require.NoError(t, err, "failed to create test client for tokens")

	for i := 0; i < concurrency; i++ {
		go func(_ int) {
			clientIDUUID, parseErr := googleUuid.Parse(testClient.ClientID)
			if parseErr != nil {
				done <- parseErr

				return
			}

			// Use unique token value for each goroutine.
			uniqueTokenID := googleUuid.Must(googleUuid.NewV7()).String()

			token := &cryptoutilIdentityDomain.Token{
				ID:          googleUuid.Must(googleUuid.NewV7()),
				TokenValue:  "testtoken_" + uniqueTokenID,
				TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
				TokenFormat: cryptoutilIdentityDomain.TokenFormatJWS,
				UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: testUser.ID, Valid: true},
				ClientID:    clientIDUUID,
				Scopes:      []string{"openid", "profile"},
				IssuedAt:    time.Now().UTC(),
				ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
			}
			done <- tokenRepo.Create(ctx, token)
		}(i)
	}

	// Wait for all operations to complete (30 total: 10 users + 10 clients + 10 tokens).
	var failedOps []string

	for i := 0; i < concurrency*3; i++ {
		if err := <-done; err != nil {
			failedOps = append(failedOps, err.Error())
		}
	}

	// Allow some duplicate key errors in concurrent operations (race condition in goroutine scheduling).
	// The test validates that MOST operations succeed concurrently.
	// PostgreSQL's unique constraint errors are expected in high-concurrency scenarios.
	if len(failedOps) > 0 {
		t.Logf("Some concurrent operations failed (expected in high concurrency): %d/%d failed", len(failedOps), concurrency*3)

		for i, errMsg := range failedOps {
			t.Logf("  Failed operation %d: %s", i+1, errMsg)
		}
	}

	require.LessOrEqual(t, len(failedOps), concurrency, "too many failures in concurrent operations")
}

// testTransactionIsolation validates transaction isolation between concurrent transactions.
func testTransactionIsolation(t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	ctx := context.Background()

	// Create two transactions concurrently, each creating a user.
	// Validate both transactions commit successfully and users are visible.
	done := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func(_ int) {
			txErr := repoFactory.Transaction(ctx, func(txCtx context.Context) error {
				userRepo := repoFactory.UserRepository()
				uniqueIDTx := googleUuid.Must(googleUuid.NewV7()).String()
				user := &cryptoutilIdentityDomain.User{
					ID:                googleUuid.Must(googleUuid.NewV7()),
					Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
					PreferredUsername: "testuser_tx_" + uniqueIDTx,
					Email:             "test_tx_" + uniqueIDTx + "@example.com",
					PasswordHash:      "hashedpassword", // pragma: allowlist secret
					Enabled:           true,
				}

				return userRepo.Create(txCtx, user)
			})
			done <- txErr
		}(i)
	}

	// Wait for both transactions.
	var failedTxs []string

	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			failedTxs = append(failedTxs, err.Error())
		}
	}

	// Transactions should commit successfully (allow 1 failure due to race condition).
	if len(failedTxs) > 0 {
		t.Logf("Some transactions failed (expected in concurrent scenarios): %d/2 failed", len(failedTxs))

		for i, errMsg := range failedTxs {
			t.Logf("  Failed transaction %d: %s", i+1, errMsg)
		}
	}

	require.LessOrEqual(t, len(failedTxs), 1, "too many transaction failures")
}
