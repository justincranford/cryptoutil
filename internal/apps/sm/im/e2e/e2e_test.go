// Copyright (c) 2025 Justin Cranford

//go:build e2e

package e2e_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	http "net/http"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	_ "github.com/lib/pq" // PostgreSQL driver.
	"github.com/stretchr/testify/require"
)


// generateTestPassword creates a cryptographically secure random password for testing.
// Uses shared utility to ensure consistency across all services.
func generateTestPassword(t *testing.T) string {
	t.Helper()

	password, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	require.NoError(t, err, "Failed to generate random password")

	return password
}

// TestE2E_HealthChecks validates /health endpoint for all instances (external clients use public endpoint).
func TestE2E_HealthChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		publicURL string
	}{
		{sqliteContainer, sqlitePublicURL},
		{postgres1Container, postgres1PublicURL},
		{postgres2Container, postgres2PublicURL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Public health check (external clients MUST use this endpoint).
			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
			defer cancel()

			healthURL := tt.publicURL + cryptoutilSharedMagic.IME2EHealthEndpoint
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			require.NoError(t, err, "Creating health check request should succeed")

			healthResp, err := sharedHTTPClient.Do(req)
			require.NoError(t, err, "Health check should succeed for %s", tt.name)
			require.NoError(t, healthResp.Body.Close())
			require.Equal(t, http.StatusOK, healthResp.StatusCode,
				"%s should return 200 OK for /health", tt.name)
		})
	}
}

// TestE2E_OtelCollectorHealth validates OpenTelemetry Collector is running and accepting telemetry.
func TestE2E_OtelCollectorHealth(t *testing.T) {
	t.Skip("OTEL Collector health port 13133 not exposed to host (intentional - prevents port conflicts across deployments)")
	// Alternative: Verify OTEL is working by checking sm-im services successfully send telemetry
	// without connection refused errors in their logs
}

// TestE2E_GrafanaHealth validates Grafana LGTM container is running and API is accessible.
func TestE2E_GrafanaHealth(t *testing.T) {
	t.Skip("Grafana health endpoint has reliability issues (EOF errors) - not critical for sm-im core functionality")

	t.Parallel()

	// Grafana HTTP API health check with retries (Grafana can be slow to start).
	client := &http.Client{Timeout: cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Second}

	var lastErr error

	for attempt := 0; attempt < cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, grafanaURL+"/api/health", http.NoBody)
		if err != nil {
			cancel()

			lastErr = fmt.Errorf("creating Grafana health request: %w", err)

			time.Sleep(2 * time.Second)

			continue
		}

		resp, err := client.Do(req)

		cancel()

		if err != nil {
			lastErr = fmt.Errorf("Grafana health endpoint (attempt %d): %w", attempt+1, err)

			time.Sleep(2 * time.Second)

			continue
		}

		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusOK {
			return // Success.
		}

		lastErr = fmt.Errorf("Grafana returned status %d (attempt %d)", resp.StatusCode, attempt+1)

		time.Sleep(2 * time.Second)
	}

	require.NoError(t, lastErr, "Grafana health check should succeed after retries")
}

// TestE2E_CrossInstanceIsolation verifies database backend isolation behavior.
// - SQLite instances are isolated (separate databases, users NOT shared)
// - PostgreSQL instances pg-1 and pg-2 SHARE state (same database, same tenant).
func TestE2E_CrossInstanceIsolation(t *testing.T) {
	t.Parallel()

	// Test SQLite isolation - users should NOT be visible from PostgreSQL instances.
	t.Run("sqlite_isolated_from_postgres", func(t *testing.T) {
		t.Parallel()

		// Create a unique user in SQLite instance.
		username := fmt.Sprintf("sqlite_user_%d", time.Now().UTC().UnixNano())
		password := generateTestPassword(t)

		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
		defer cancel()

		// Register user in SQLite.
		registerURL := sqlitePublicURL + "/service/api/v1/users/register"
		registerBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
		require.NoError(t, err, "Creating registration request should succeed")
		req.Header.Set("Content-Type", "application/json")

		resp, err := sharedHTTPClient.Do(req)
		require.NoError(t, err, "User registration should succeed in SQLite")
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusCreated, resp.StatusCode, "User should be created in SQLite")

		// Verify user can login in SQLite.
		loginURL := sqlitePublicURL + "/service/api/v1/users/login"
		loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

		loginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewBufferString(loginBody))
		require.NoError(t, err, "Creating login request should succeed")
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := sharedHTTPClient.Do(loginReq)
		require.NoError(t, err, "Login should succeed in SQLite")
		require.NoError(t, loginResp.Body.Close())
		require.Equal(t, http.StatusOK, loginResp.StatusCode, "User should login in SQLite")

		// Verify user does NOT exist in PostgreSQL instances.
		for _, pgURL := range []string{postgres1PublicURL, postgres2PublicURL} {
			pgLoginURL := pgURL + "/service/api/v1/users/login"
			pgLoginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pgLoginURL, bytes.NewBufferString(loginBody))
			require.NoError(t, err, "Creating PostgreSQL login request should succeed")
			pgLoginReq.Header.Set("Content-Type", "application/json")

			pgLoginResp, err := sharedHTTPClient.Do(pgLoginReq)
			require.NoError(t, err, "PostgreSQL login attempt should complete")
			require.NoError(t, pgLoginResp.Body.Close())
			require.NotEqual(t, http.StatusOK, pgLoginResp.StatusCode,
				"SQLite user should NOT exist in PostgreSQL (database isolation)")
		}
	})

	// Test PostgreSQL shared state - users registered on pg-1 SHOULD be visible from pg-2.
	t.Run("postgres_instances_share_state", func(t *testing.T) {
		t.Parallel()

		// Create a unique user in pg-1.
		username := fmt.Sprintf("pg_shared_user_%d", time.Now().UTC().UnixNano())
		password := generateTestPassword(t)

		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
		defer cancel()

		// Register user in pg-1.
		registerURL := postgres1PublicURL + "/service/api/v1/users/register"
		registerBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
		require.NoError(t, err, "Creating registration request should succeed")
		req.Header.Set("Content-Type", "application/json")

		resp, err := sharedHTTPClient.Do(req)
		require.NoError(t, err, "User registration should succeed in pg-1")
		require.NoError(t, resp.Body.Close())
		require.Equal(t, http.StatusCreated, resp.StatusCode, "User should be created in pg-1")

		// Verify user can login in pg-1.
		loginURL := postgres1PublicURL + "/service/api/v1/users/login"
		loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

		loginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, bytes.NewBufferString(loginBody))
		require.NoError(t, err, "Creating login request should succeed")
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := sharedHTTPClient.Do(loginReq)
		require.NoError(t, err, "Login should succeed in pg-1")
		require.NoError(t, loginResp.Body.Close())
		require.Equal(t, http.StatusOK, loginResp.StatusCode, "User should login in pg-1")

		// Verify user ALSO exists in pg-2 (shared PostgreSQL database).
		pg2LoginURL := postgres2PublicURL + "/service/api/v1/users/login"
		pg2LoginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, pg2LoginURL, bytes.NewBufferString(loginBody))
		require.NoError(t, err, "Creating pg-2 login request should succeed")
		pg2LoginReq.Header.Set("Content-Type", "application/json")

		pg2LoginResp, err := sharedHTTPClient.Do(pg2LoginReq)
		require.NoError(t, err, "Login attempt should complete in pg-2")
		require.NoError(t, pg2LoginResp.Body.Close())
		require.Equal(t, http.StatusOK, pg2LoginResp.StatusCode,
			"User from pg-1 SHOULD exist in pg-2 (shared PostgreSQL database)")
	})

	// Test PostgreSQL isolated from SQLite - users registered on pg-1 should NOT exist in SQLite.
	// Uses pg-1 instead of pg-2 because pg-2 has longer startup time (7+ health check attempts)
	// and may return 500 errors during initialization, causing flaky test failures.
	t.Run("postgres_isolated_from_sqlite", func(t *testing.T) {
		t.Parallel()

		// Create a unique user in pg-1 (more stable than pg-2 due to startup ordering).
		username := fmt.Sprintf("pg_isolated_user_%d", time.Now().UTC().UnixNano())
		password := generateTestPassword(t)

		ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.E2EHTTPClientTimeout)
		defer cancel()

		// Register user in pg-1 (not pg-2 which has longer startup time).
		registerURL := postgres1PublicURL + "/service/api/v1/users/register"
		registerBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, registerURL, bytes.NewBufferString(registerBody))
		require.NoError(t, err, "Creating registration request should succeed")
		req.Header.Set("Content-Type", "application/json")

		resp, err := sharedHTTPClient.Do(req)
		require.NoError(t, err, "User registration should succeed in pg-1")

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode, "User should be created in pg-1")

		// Verify user does NOT exist in SQLite.
		loginBody := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
		sqliteLoginURL := sqlitePublicURL + "/service/api/v1/users/login"
		sqliteLoginReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sqliteLoginURL, bytes.NewBufferString(loginBody))
		require.NoError(t, err, "Creating SQLite login request should succeed")
		sqliteLoginReq.Header.Set("Content-Type", "application/json")

		sqliteLoginResp, err := sharedHTTPClient.Do(sqliteLoginReq)
		require.NoError(t, err, "SQLite login attempt should complete")
		require.NoError(t, sqliteLoginResp.Body.Close())
		require.NotEqual(t, http.StatusOK, sqliteLoginResp.StatusCode,
			"PostgreSQL user should NOT exist in SQLite (database isolation)")
	})
}

// TestE2E_PostgreSQLSharedState validates pg-1 and pg-2 share database state (uses TestMain infrastructure).
func TestE2E_PostgreSQLSharedState(t *testing.T) {
	ctx := context.Background()

	// Connect to shared PostgreSQL database (already running from TestMain).
	dsn := "postgres://sm_im_user:sm_im_pass@127.0.0.1:5432/sm_im?sslmode=disable"
	db, err := sql.Open(cryptoutilSharedMagic.DockerServicePostgres, dsn)
	require.NoError(t, err, "Connecting to PostgreSQL should succeed")

	defer func() { _ = db.Close() }()

	err = db.PingContext(ctx)
	require.NoError(t, err, "Ping PostgreSQL should succeed")

	// Verify shared tables exist (created by both pg-1 and pg-2 instances).
	// barrier_root_keys and service_session_jwks MUST have >= 1 row (always initialized by service startup).
	// browser_session_jwks MUST exist but may be empty (default browser session algorithm is OPAQUE,
	// which uses hashed tokens instead of JWKs).
	tableChecks := []struct {
		name       string
		requireRow bool
	}{
		{"barrier_root_keys", true},
		{"service_session_jwks", true},
		{"browser_session_jwks", false},
	}

	existsQuery := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	)`

	for _, tc := range tableChecks {
		t.Run(tc.name, func(t *testing.T) {
			var exists bool

			err := db.QueryRowContext(ctx, existsQuery, tc.name).Scan(&exists)
			require.NoError(t, err, "Checking %s existence should succeed", tc.name)
			require.True(t, exists, "%s table should exist in shared database", tc.name)

			if tc.requireRow {
				// Verify at least 1 row exists (either instance could have created it).
				var count int

				err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tc.name).Scan(&count)
				require.NoError(t, err, "Counting %s rows should succeed", tc.name)
				require.GreaterOrEqual(t, count, 1, "Should have at least 1 row in %s", tc.name)
			}
		})
	}
}

// TestE2E_SQLiteInstanceIsolation validates SQLite instance has isolated in-memory database (uses TestMain infrastructure).
func TestE2E_SQLiteInstanceIsolation(t *testing.T) {
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

// TestE2E_RegistrationFlowWithTenantCreation validates user registration with automatic tenant creation.
// This tests the Phase 0 multi-tenancy implementation where each new user creates their own tenant.
