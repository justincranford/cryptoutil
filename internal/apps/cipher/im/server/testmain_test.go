// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"cryptoutil/internal/apps/cipher/im/server"
	cipherTesting "cryptoutil/internal/apps/cipher/im/testing"
	cryptoutilTLSGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
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
