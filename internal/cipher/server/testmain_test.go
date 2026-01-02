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
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/cipher/repository"
	"cryptoutil/internal/cipher/server"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilTemplateBarrier "cryptoutil/internal/template/server/barrier"
)

var (
	testDB                  *gorm.DB
	testSQLDB               *sql.DB // CRITICAL: Keep reference to prevent GC - in-memory SQLite requires open connection
	testUserRepo            *repository.UserRepository
	testMessageRepo         *repository.MessageRepository
	testMessageRecipientJWK *repository.MessageRecipientJWKRepository
	testBarrierService      *cryptoutilTemplateBarrier.BarrierService
	testJWKGenService       *cryptoutilJose.JWKGenService
	testTelemetryService    *cryptoutilTelemetry.TelemetryService
	testJWTSecret           string
	testTLSCfg              *cryptoutilTLSGenerator.TLSGeneratedSettings
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

	// Initialize JWK Generation Service.
	testJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}

	// Initialize Barrier Service.
	// Generate a simple test unseal key using JWE with A256GCM encryption and A256KW key wrapping.
	_, testUnsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate test unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}

	testBarrierService, err = cryptoutilTemplateBarrier.NewBarrierService(ctx, testTelemetryService, testJWKGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}

	// Initialize repositories.
	testUserRepo = repository.NewUserRepository(testDB)
	testMessageRepo = repository.NewMessageRepository(testDB)
	testMessageRecipientJWK = repository.NewMessageRecipientJWKRepository(testDB, testBarrierService)

	// Generate TLS config.
	testTLSCfg, err = cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilMagic.HostnameLocalhost},
		[]string{cryptoutilMagic.IPv4Loopback},
		cryptoutilMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		panic("TestMain: failed to generate TLS config: " + err.Error())
	}

	// Generate JWT secret.
	jwtSecretID, err := googleUuid.NewV7()
	if err != nil {
		panic("TestMain: failed to generate JWT secret: " + err.Error())
	}

	testJWTSecret = jwtSecretID.String()

	// Record start time for benchmark.
	startTime := time.Now()

	// Run all tests.
	exitCode := m.Run()

	// CRITICAL: DO NOT close database immediately.
	// Parallel tests may still be running after m.Run() returns.
	// The Go test runner will wait for all goroutines to complete before process exit.
	// Closing resources here would cause "database is closed" errors in slow parallel tests.

	// Cleanup telemetry (safe to shutdown immediately - no active operations).
	if testTelemetryService != nil {
		testTelemetryService.Shutdown()
	}

	// DEFER database close to OS process exit (let Go runtime handle cleanup).
	// Database will auto-close when process exits.
	// This prevents "database is closed" errors in long-running parallel tests.

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

// createTestPublicServer now uses shared resources.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	// Clean database for test isolation.
	cleanTestDB(t)

	// Use port 0 for dynamic allocation.
	const testPort = 0

	publicServer, err := server.NewPublicServer(
		ctx,
		testPort,
		testUserRepo,
		testMessageRepo,
		testMessageRecipientJWK,
		testJWKGenService,
		testBarrierService,
		testJWTSecret,
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
