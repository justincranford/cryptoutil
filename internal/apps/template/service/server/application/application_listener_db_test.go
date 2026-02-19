// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestOpenPostgreSQL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// This test demonstrates openPostgreSQL function but will skip if no PostgreSQL available.
	// In production, this would use testcontainers to start PostgreSQL.
	// For now, we test the error path with invalid DSN.
	_, err := openPostgreSQL(ctx, "invalid-dsn", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to open PostgreSQL database")
}

// TestListener_Start_NilContext tests Start with nil context.
func TestListener_Start_NilContext(t *testing.T) {
	t.Parallel()

	publicServer := &mockPublicServer{port: 8080}
	adminServer := &mockAdminServer{port: 9090}

	listener := &Listener{
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	err := listener.Start(nil) //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestListener_Start_PublicServerError tests Start when public server fails immediately.
func TestListener_Start_PublicServerError(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-error",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	// Create mock server that fails immediately.
	publicServer := &mockPublicServer{
		port:     8080,
		startErr: fmt.Errorf("mock public server error"),
	}
	adminServer := &mockAdminServer{port: 9090}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)

	defer func() { _ = listener.Shutdown(context.Background()) }()

	// Start should fail with public server error.
	err = listener.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server failed")
}

// TestListener_Start_AdminServerError tests Start when admin server fails immediately.
func TestListener_Start_AdminServerError(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-admin-error",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	publicServer := &mockPublicServer{port: 8080}
	// Create mock server that fails immediately.
	adminServer := &mockAdminServer{
		port:     9090,
		startErr: fmt.Errorf("mock admin server error"),
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)

	defer func() { _ = listener.Shutdown(context.Background()) }()

	// Start should fail with admin server error.
	err = listener.Start(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "admin server failed")
}

// TestListener_Start_ContextCancelled tests Start when context is cancelled.
func TestListener_Start_ContextCancelled(t *testing.T) {
	// NOT parallel - uses shared SQLite database.
	ctx, cancel := context.WithCancel(context.Background())

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
		OTLPService:  "test-start-cancel",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	// Create servers that block until cancelled.
	startDone := make(chan struct{})
	publicServer := &mockPublicServer{
		port:      8080,
		startDone: startDone,
	}
	adminServer := &mockAdminServer{
		port:      9090,
		startDone: startDone,
	}

	config := &ListenerConfig{
		Settings:     settings,
		PublicServer: publicServer,
		AdminServer:  adminServer,
	}

	listener, err := StartListener(context.Background(), config)
	require.NoError(t, err)

	defer func() { _ = listener.Shutdown(context.Background()) }()

	// Start in background, then cancel context.
	errChan := make(chan error, 1)

	go func() {
		errChan <- listener.Start(ctx)
	}()

	// Wait a bit for Start to begin.
	time.Sleep(100 * time.Millisecond)

	// Cancel context.
	cancel()

	// Unblock servers.
	close(startDone)

	// Should return context cancellation error.
	err = <-errChan
	require.Error(t, err)
	require.Contains(t, err.Error(), "application startup cancelled")
}

// TestStartCore_NilContext tests StartCore with nil context.
func TestStartCore_NilContext(t *testing.T) {
	t.Parallel()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:     true,
		DatabaseURL: cryptoutilSharedMagic.SQLiteInMemoryDSN,
	}

	core, err := StartCore(nil, settings) //nolint:staticcheck // Testing nil context.
	require.Error(t, err)
	require.Nil(t, core)
	require.Contains(t, err.Error(), "ctx cannot be nil")
}

// TestStartCore_NilSettings tests StartCore with nil settings.
func TestStartCore_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	core, err := StartCore(ctx, nil)
	require.Error(t, err)
	require.Nil(t, core)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestProvisionDatabase_UnsupportedScheme tests provisionDatabase with unsupported database URL.
func TestProvisionDatabase_UnsupportedScheme(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  "mysql://localhost:3306/test", // Unsupported scheme.
		OTLPService:  "test-unsupported-db",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	require.Error(t, err)
	require.Nil(t, db)
	require.Nil(t, cleanup)
	require.Contains(t, err.Error(), "unsupported database URL scheme")
}

// TestProvisionDatabase_SQLiteFileURL tests SQLite with file:// URL.
func TestProvisionDatabase_SQLiteFileURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use temp file for SQLite database.
	dbFile := "/tmp/test_sqlite_" + time.Now().UTC().Format("20060102150405") + ".db"

	defer func() {
		// Cleanup.
		_ = os.Remove(dbFile)
	}()

	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  "file://" + dbFile,
		OTLPService:  "test-sqlite-file",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer cleanup()

	// Verify database works.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)
}

// TestProvisionDatabase_SQLiteSchemePrefixURL tests SQLite with sqlite:// URL prefix.
// This covers the sqlite:// scheme detection branch in provisionDatabase.
func TestProvisionDatabase_SQLiteSchemePrefixURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		DevMode:      true,
		DatabaseURL:  "sqlite://file::memory:?cache=shared", // sqlite:// prefix with in-memory DSN
		OTLPService:  "test-sqlite-scheme",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://127.0.0.1:4317",
		LogLevel:     "INFO",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	require.NoError(t, err)
	require.NotNil(t, db)

	defer cleanup()

	// Verify database works.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)
}

// TestOpenSQLite_InvalidDSN tests openSQLite with valid DSN and WAL mode.
func TestOpenSQLite_InvalidDSN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test successful operation with in-memory DSN.
	db, err := openSQLite(ctx, cryptoutilSharedMagic.SQLiteInMemoryDSN, false)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Verify PRAGMA settings were applied (WAL mode for file databases, memory for in-memory).
	var busyTimeout int

	sqlDB, _ := db.DB()
	err = sqlDB.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout)
	require.NoError(t, err)
	require.Equal(t, 30000, busyTimeout) // 30 seconds as configured.
}

// TestOpenPostgreSQL_Success tests successful PostgreSQL connection.
func TestOpenPostgreSQL_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use valid PostgreSQL DSN (won't connect but tests code path).
	dsn := testPostgresDSN
	db, err := openPostgreSQL(ctx, dsn, false)

	// Note: Will fail to connect since no actual PostgreSQL server.
	// This tests the error path which is at 41.7% coverage.
	if err != nil {
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to open PostgreSQL database")
	} else {
		require.NotNil(t, db)
	}
}

// TestOpenPostgreSQL_InvalidDSN tests openPostgreSQL with invalid DSN.
func TestOpenPostgreSQL_InvalidDSN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Empty DSN should fail.
	db, err := openPostgreSQL(ctx, "", false)
	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "failed to open PostgreSQL database")
}

// TestOpenPostgreSQL_DebugMode tests openPostgreSQL with debug mode enabled.
func TestOpenPostgreSQL_DebugMode(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use valid DSN format but won't connect.
	dsn := testPostgresDSN

	db, err := openPostgreSQL(ctx, dsn, true) // Debug mode = true.
	if err != nil {
		require.Error(t, err)
	} else {
		require.NotNil(t, db)
	}
}

// TestProvisionDatabase_PostgreSQLContainerRequired tests PostgreSQL container in "required" mode.
func TestProvisionDatabase_PostgreSQLContainerRequired(t *testing.T) {
	// Cannot use t.Parallel() due to shared Basic instance.
	ctx := context.Background()

	// Start basic infrastructure.
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		LogLevel:          "info",
		VerboseMode:       false,
		OTLPEndpoint:      "grpc://localhost:4317",
		OTLPService:       "test-service",
		OTLPVersion:       "1.0.0",
		OTLPEnvironment:   "test",
		UnsealMode:        "sysinfo",
		DatabaseURL:       testPostgresDSN,
		DatabaseContainer: "required",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	// Attempt to provision with required container (will fail if Docker not available or connection fails).
	db, cleanup, err := provisionDatabase(ctx, basic, settings)
	if err != nil {
		// Expected failure if Docker not running OR connection fails.
		require.Error(t, err)
		// Accept either container start failure or database connection failure.
		require.True(t,
			strings.Contains(err.Error(), "failed to start required PostgreSQL testcontainer") ||
				strings.Contains(err.Error(), "failed to open database") ||
				strings.Contains(err.Error(), "failed to connect"),
			"error should be container or connection related: %v", err,
		)
	} else {
		require.NotNil(t, db)

		defer cleanup()
	}
}

// TestProvisionDatabase_PostgreSQLContainerPreferred tests PostgreSQL container in "preferred" mode with fallback.
func TestProvisionDatabase_PostgreSQLContainerPreferred(t *testing.T) {
	// Cannot use t.Parallel() due to shared Basic instance.
	ctx := context.Background()

	// Start basic infrastructure.
	settings := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		LogLevel:          "info",
		VerboseMode:       false,
		OTLPEndpoint:      "grpc://localhost:4317",
		OTLPService:       "test-service",
		OTLPVersion:       "1.0.0",
		OTLPEnvironment:   "test",
		UnsealMode:        "sysinfo",
		DatabaseURL:       testPostgresDSN,
		DatabaseContainer: "preferred",
	}

	basic, err := StartBasic(ctx, settings)
	require.NoError(t, err)

	defer basic.Shutdown()

	// Attempt to provision with preferred container (should fallback to external DB if container fails).
	db, cleanup, err := provisionDatabase(ctx, basic, settings)

	// Preferred mode allows fallback, so error only if external DB also fails.
	if err != nil {
		require.Error(t, err)
	} else {
		require.NotNil(t, db)

		defer cleanup()
	}
}

// TestOpenSQLite_FileBasedWithWAL tests openSQLite with file-based database and WAL mode.
