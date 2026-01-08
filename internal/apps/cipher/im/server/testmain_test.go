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

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"cryptoutil/internal/apps/cipher/im/repository"
	"cryptoutil/internal/apps/cipher/im/server"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
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

	// Setup: Create shared heavyweight resources ONCE using helper.
	resources, err := cipherTesting.SetupTestServer(ctx, false)
	if err != nil {
		panic("TestMain: failed to setup test server: " + err.Error())
	}
	defer resources.Shutdown(context.Background())

	// Assign to package-level variables for backward compatibility with existing tests.
	testDB = resources.DB
	testSQLDB = resources.SQLDB
	testCipherIMServer = resources.CipherIMServer
	baseURL = resources.BaseURL
	adminURL = resources.AdminURL
	testJWKGenService = resources.JWKGenService
	testTelemetryService = resources.TelemetryService
	testTLSCfg = resources.TLSCfg

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
	jwtSecretID, err := cryptoutilRandom.GenerateUUIDv7()
	require.NoError(t, err)

	jwtSecret := jwtSecretID.String()

	// Use port 0 for dynamic allocation.
	const testPort = 0

	// Get session manager from testCipherIMServer (created in TestMain).
	// Session manager is required for PublicServer creation.
	sessionManager := testCipherIMServer.SessionManager()
	require.NotNil(t, sessionManager, "testCipherIMServer.SessionManager() should not be nil")

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
		nil,            // barrierService - nil for lightweight testing
		sessionManager, // sessionManagerService - from testCipherIMServer
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
