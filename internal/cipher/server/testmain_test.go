// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testDB    *gorm.DB
	testSQLDB *sql.DB

	testCipherIMServer *server.CipherIMServer
	baseURL            string
	adminURL           string

	testJWKGenService    *cryptoutilJose.JWKGenService
	testTelemetryService *cryptoutilTelemetry.TelemetryService

	testTLSCfg *cryptoutilTLSGenerator.TLSGeneratedSettings
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup: Create shared heavyweight resources ONCE.
	dbID, _ := googleUuid.NewV7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	// CRITICAL: Store sql.DB reference in package variable.
	// In-memory SQLite databases are destroyed when all connections close.
	// Storing reference prevents GC from closing connection during parallel test execution.
	var err error

	testSQLDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	// Configure SQLite for concurrent operations.
	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Run migrations.
	if err := repository.ApplyMigrations(testSQLDB, repository.DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to run migrations: " + err.Error())
	}

	// Initialize telemetry.
	testTelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown() // LIFO: cleanup last service created.

	// Initialize JWK Generation Service.
	testJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}
	defer testJWKGenService.Shutdown() // LIFO: cleanup JWK service.

	// Generate TLS config for HTTP client (server creates its own TLS config).
	testTLSCfg, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		panic("TestMain: failed to generate TLS config: " + err.Error())
	}

	// Create shared CipherIMServer (includes barrier service, repositories, both public and admin servers).
	testCipherIMServer, baseURL, adminURL, err = createTestCipherIMServer(testDB)
	if err != nil {
		panic("TestMain: failed to create test cipher-im server: " + err.Error())
	}

	// Defer database close and server shutdown (LIFO: executes AFTER m.Run() completes).
	defer func() {
		_ = testSQLDB.Close()
	}()
	defer testCipherIMServer.Shutdown(context.Background())

	// Record start time for benchmark.
	startTime := time.Now()

	// Run all tests (defer statements will execute cleanup AFTER m.Run() completes).
	exitCode := m.Run()

	elapsed := time.Since(startTime)

	// Log timing for comparison (visible in test output).
	// IMPORTANT: This timing includes TestMain setup overhead, which is amortized across all tests.
	// Individual test functions no longer pay setup cost - they reuse shared resources.
	println("TestMain: All tests completed in", elapsed.String())

	os.Exit(exitCode)
}

// Helper functions now use shared test resources from TestMain.

// initTestDB is deprecated - tests should use testDB directly or call cleanTestDB for isolation.
// Kept for backward compatibility during migration.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	// Clean existing data for test isolation while reusing schema and connection.
	cleanTestDB(t)

	return testDB
}

// cleanTestDB truncates all tables for test isolation while preserving schema.
func cleanTestDB(t *testing.T) {
	t.Helper()

	// Truncate tables to ensure test isolation (schema already exists from TestMain).
	tables := []string{"messages", "users", "messages_recipient_jwks"}
	for _, table := range tables {
		require.NoError(t, testDB.Exec("DELETE FROM "+table).Error)
	}
}

// createTestPublicServer creates PublicServer with its own dependencies for test isolation.
// Note: This old helper creates PublicServer without barrier service.
// For full server testing with both public and admin servers, use shared testCipherIMServer from TestMain.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	// Clean database for test isolation.
	cleanTestDB(t)

	// Create repositories for this server instance.
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	// Create messageRecipientJWKRepo with nil barrier service (repo doesn't actually need barrier service for basic operations).
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db, nil)

	// Generate JWT secret for this server instance.
	jwtSecretID, err := googleUuid.NewV7()
	require.NoError(t, err)

	jwtSecret := jwtSecretID.String()

	// Use port 0 for dynamic allocation.
	const testPort = 0

	// Create PublicServer without barrier service (passed as nil).
	// Barrier service requires unseal keys and adds significant setup complexity.
	// Tests needing barrier service should use testCipherIMServer from TestMain.
	publicServer, err := server.NewPublicServer(
		ctx,
		testPort,
		userRepo,
		messageRepo,
		messageRecipientJWKRepo, // messageRecipientJWKRepo created with nil barrier service
		testJWKGenService,
		nil, // barrierService - nil for lightweight testing
		jwtSecret,
		testTLSCfg,
	)
	require.NoError(t, err)

	// Start server in background.
	errChan := make(chan error, 1)

	go func() {
		if startErr := publicServer.Start(ctx); startErr != nil {
			errChan <- startErr
		}
	}()

	// Wait for server to bind to port.
	const (
		maxWaitAttempts = 50
		waitInterval    = 100 * time.Millisecond
	)

	actualPort := 0
	for i := 0; i < maxWaitAttempts; i++ {
		actualPort = publicServer.ActualPort()
		if actualPort > 0 {
			break
		}

		select {
		case err := <-errChan:
			require.NoError(t, err)
		case <-time.After(waitInterval):
		}
	}

	if actualPort == 0 {
		t.Fatal("createTestPublicServerShared: server did not bind to port")
	}

	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilMagic.IPv4Loopback, actualPort)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := publicServer.Shutdown(ctx); err != nil {
			t.Logf("createTestPublicServerShared cleanup: failed to shutdown server: %v", err)
		}
	})

	return publicServer, baseURL
}
