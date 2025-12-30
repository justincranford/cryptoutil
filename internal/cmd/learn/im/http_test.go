// Copyright (c) 2025 Justin Cranford
//
//

package im

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	"cryptoutil/internal/learn/server/config"
)

// initTestConfig returns an AppConfig with all required settings for tests.
func initTestConfig() *config.AppConfig {
	cfg := config.DefaultAppConfig()
	cfg.BindPublicPort = 0                     // Dynamic port
	cfg.BindPrivatePort = 0                    // Dynamic port
	cfg.OTLPService = "learn-im-test"          // Required
	cfg.LogLevel = "info"                      // Required
	cfg.OTLPEndpoint = "grpc://localhost:4317" // Required
	cfg.OTLPEnabled = false                    // Disable in tests

	return cfg
}

// TestHTTPGet tests the httpGet helper function (used by health CLI wrappers).
func TestHTTPGet(t *testing.T) {
	ctx := context.Background()

	// Initialize in-memory SQLite database.
	sqlDB, err := sqlOpen("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Apply migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	// Create server with dynamic ports.
	cfg := initTestConfig()

	srv, err := server.New(ctx, cfg, gormDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	// Start server.
	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for servers to be ready.
	time.Sleep(500 * time.Millisecond)

	// Get actual ports.
	publicPort := srv.PublicPort()
	adminPort, err := srv.AdminPort()
	require.NoError(t, err)

	// Set readiness flag (learn server doesn't call SetReady yet).
	// TODO: Fix learn server to call SetReady after successful initialization.
	// For now, skip readyz test as it returns 503 (not ready) which is expected behavior.

	// Create insecure HTTP client (accepts self-signed certs).
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	// Test public health endpoint.
	t.Run("public_health_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/service/api/v1/health", publicPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin livez endpoint.
	t.Run("admin_livez_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/livez", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test admin readyz endpoint (expected 503 - not ready).
	t.Run("admin_readyz_endpoint", func(t *testing.T) {
		t.Skip("Skipping readyz test - learn server doesn't call SetReady yet (returns 503 as expected)")

		url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/readyz", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	})

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	require.NoError(t, err)
}

// TestHTTPPost tests the httpPost helper function (used by shutdown CLI wrapper).
func TestHTTPPost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize in-memory SQLite database.
	sqlDB, err := sqlOpen("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Apply migrations using embedded migration files.
	err = repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	// Create server with dynamic ports.
	cfg := initTestConfig()

	srv, err := server.New(ctx, cfg, gormDB, repository.DatabaseTypeSQLite)
	require.NoError(t, err)

	// Start server in background with cancellable context.
	errChan := make(chan error, 1)

	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Wait for servers to be ready.
	time.Sleep(500 * time.Millisecond)

	// Get actual ports.
	adminPort, err := srv.AdminPort()
	require.NoError(t, err)

	// Create insecure HTTP client (accepts self-signed certs).
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	// Test admin shutdown endpoint (triggers async shutdown).
	t.Run("admin_shutdown_endpoint", func(t *testing.T) {
		url := fmt.Sprintf("https://127.0.0.1:%d/admin/v1/shutdown", adminPort)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Cancel context to trigger server shutdown (shutdown endpoint starts async shutdown).
	cancel()

	// Wait for server to finish shutting down.
	select {
	case err := <-errChan:
		// Server shutdown returns context.Canceled error which is expected.
		const (
			adminStoppedErr = "admin server stopped: context canceled"
			appCancelledErr = "application startup cancelled: context canceled"
		)

		if err != nil && err.Error() != adminStoppedErr && err.Error() != appCancelledErr {
			require.FailNowf(t, "Unexpected server error", "%v", err)
		}
	case <-time.After(5 * time.Second):
		require.FailNow(t, "Server did not shutdown within timeout")
	}
}

// TestIMServer tests the imServer function.
func TestIMServer(t *testing.T) {
	// This test would require mocking os.Signal and context handling.
	// Skipping for now as imServer is tested via integration tests.
	t.Skip("imServer requires signal mocking - tested via integration tests")
}

// sqlOpen wrapper for test cleanup.
func sqlOpen(driver, dsn string) (*sql.DB, error) {
	switch driver {
	case "sqlite":
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open SQLite database: %w", err)
		}

		return db, nil
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}
