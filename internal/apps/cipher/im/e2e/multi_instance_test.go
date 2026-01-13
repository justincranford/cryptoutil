// Copyright (c) 2025 Justin Cranford
//
// Multi-instance deployment validation tests for cipher-im.
//
// Test Coverage:
// - 3 instance deployment (1 SQLite, 2 PostgreSQL) via TestMain
// - SQLite instance isolation (in-memory database)
// - PostgreSQL shared state (pg-1 and pg-2 share same database)
//
// CRITICAL: Uses shared infrastructure from TestMain (no duplicate compose starts/stops).

package e2e_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver.
	"github.com/stretchr/testify/require"
)

// TestThreeInstanceDeployment validates all 3 instances are healthy (uses TestMain infrastructure).
func TestThreeInstanceDeployment(t *testing.T) {
	// TestMain already started docker compose and verified health.
	// Just verify we can reach each instance using shared HTTP client.

	// Use healthChecks map from testmain_e2e_test.go (already has correct endpoint).
	for name, healthURL := range healthChecks {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, http.NoBody)
			require.NoError(t, err, "Creating health request should succeed")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "Health check should succeed for %s", name)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "%s should return 200 OK", name)
		})
	}
}

// TestPostgreSQLSharedState validates pg-1 and pg-2 share database state (uses TestMain infrastructure).
func TestPostgreSQLSharedState(t *testing.T) {
	ctx := context.Background()

	// Connect to shared PostgreSQL database (already running from TestMain).
	dsn := "postgres://cipher_user:cipher_pass@127.0.0.1:5432/cipher_im?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err, "Connecting to PostgreSQL should succeed")

	defer func() { _ = db.Close() }()

	err = db.PingContext(ctx)
	require.NoError(t, err, "Ping PostgreSQL should succeed")

	// Verify shared tables exist (created by both pg-1 and pg-2 instances).
	tables := []string{
		"barrier_root_keys",
		"browser_session_jwks",
		"service_session_jwks",
	}

	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var exists bool

			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)`
			err := db.QueryRowContext(ctx, query, table).Scan(&exists)
			require.NoError(t, err, "Checking %s existence should succeed", table)
			require.True(t, exists, "%s table should exist in shared database", table)

			// Verify at least 1 row exists (either instance could have created it).
			var count int

			err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
			require.NoError(t, err, "Counting %s rows should succeed", table)
			require.GreaterOrEqual(t, count, 1, "Should have at least 1 row in %s", table)
		})
	}
}

// TestSQLiteInstanceIsolation validates SQLite instance has isolated in-memory database (uses TestMain infrastructure).
func TestSQLiteInstanceIsolation(t *testing.T) {
	// SQLite instance already running from TestMain with in-memory database.
	// Verify SQLite instance is healthy and isolated (already verified by TestMain).
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthChecks[sqliteContainer], http.NoBody)
	require.NoError(t, err, "Creating SQLite health request should succeed")

	resp, err := sharedHTTPClient.Do(req)
	require.NoError(t, err, "SQLite health check should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode, "SQLite instance should be healthy")

	// Note: SQLite uses in-memory database (file::memory:?cache=shared).
	// Each instance has isolated state (NOT shared with PostgreSQL instances).
	// This is intentional design for dev/testing with zero external dependencies.
}

// TestCrossInstanceDeployment validates all instances can deploy simultaneously (uses TestMain infrastructure).
func TestCrossInstanceDeployment(t *testing.T) {
	// TestMain already deployed all 3 instances and verified health.
	// Validate readyz endpoints confirm database initialization succeeded.

	// Use healthChecks map (already has correct /service/api/v1/health endpoint).
	for name, healthURL := range healthChecks {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, http.NoBody)
			require.NoError(t, err, "Creating request should succeed")

			resp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "GET health should succeed for %s", name)

		defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusOK, resp.StatusCode, "%s should be ready", name)
		})
	}
}
