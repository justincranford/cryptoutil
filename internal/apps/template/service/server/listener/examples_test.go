// Copyright (c) 2025 Justin Cranford
//
//

// Package listener_test provides examples of ApplicationListener usage patterns.
//
// This file demonstrates the BEFORE and AFTER of migrating to ApplicationListener.

package listener_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ========================================
// BEFORE: Messy TestMain (150+ lines)
// ========================================

/*
var (
	testDB    *gorm.DB
	testSQLDB *sql.DB

	testSmIMServer *server.SmIMServer
	baseURL            string
	adminURL           string

	testJWKGenService    *cryptoutilJose.JWKGenService
	testTelemetryService *cryptoutilTelemetry.TelemetryService

	testTLSCfg *cryptoutilTLSGenerator.TLSGeneratedSettings
)

func TestMainBefore(m *testing.M) {
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

	// Create shared SmIMServer (includes barrier service, repositories, both public and admin servers).
	testSmIMServer, baseURL, adminURL, err = createTestSmIMServer(testDB)
	if err != nil {
		panic("TestMain: failed to create test sm-im server: " + err.Error())
	}

	// Defer database close and server shutdown (LIFO: executes AFTER m.Run() completes).
	defer func() {
		_ = testSQLDB.Close()
	}()
	defer testSmIMServer.Shutdown(context.Background())

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
*/

// ========================================
// AFTER: Clean TestMain (30 lines)
// ========================================

// Example functions below are for documentation purposes only.
// They demonstrate the AFTER pattern but are not executed in tests.
// Prefixed with underscore to indicate they are intentionally unused.

//nolint:unused // Example variables for documentation
var (
	// Single listener encapsulates ALL infrastructure (telemetry, JWK gen, DB, servers).
	_testListener *cryptoutilAppsTemplateServiceServerListener.ApplicationListener
	_baseURL      string
	_adminURL     string
)

//nolint:unused // Example function for documentation
func _testMainAfter(m *testing.M) {
	ctx := context.Background()

	// Create in-memory SQLite database (reusable helper).
	db, sqlDB, err := _createInMemoryDB(ctx)
	if err != nil {
		panic("failed to create database: " + err.Error())
	}

	defer func() { _ = sqlDB.Close() }()

	// Apply migrations (product-specific, but pattern is standard).
	// if err := repository.ApplyMigrations(sqlDB, repository.DatabaseTypeSQLite); err != nil {
	//     panic("failed to apply migrations: " + err.Error())
	// }

	// Configure application (product-specific settings injected here).
	cfg := &cryptoutilAppsTemplateServiceServerListener.ApplicationConfig{
		ServiceTemplateServerSettings: cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true),
		DB:                            db,
		DBType:                        cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite,
		// PublicHandlers: registerSmIMHandlers, // Inject product-specific routes
		// AdminHandlers:  registerBarrierRotation,  // Optional: barrier rotation endpoints
	}

	// Start application (ONE LINE - encapsulates 150+ lines of boilerplate!).
	_testListener, err = cryptoutilAppsTemplateServiceServerListener.StartApplicationListener(ctx, cfg)
	if err != nil {
		panic("failed to start application: " + err.Error())
	}
	defer _testListener.Shutdown()

	// Extract URLs for tests (automatic port allocation).
	_baseURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, _testListener.ActualPublicPort())
	_adminURL = fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.IPv4Loopback, _testListener.ActualPrivatePort())

	os.Exit(m.Run())
}

// ========================================
// Reusable Helper Functions
// ========================================

// _createInMemoryDB creates an in-memory SQLite database configured for concurrent operations.
// Returns GORM DB, sql.DB (for migrations), and error.
//
// This helper is REUSABLE across all services (sm-im, jose-ja, identity-*, pki-ca).
// Extract to internal/template/testing/database/ for shared usage.
//
//nolint:unused // Example function for documentation
func _createInMemoryDB(ctx context.Context) (*gorm.DB, *sql.DB, error) {
	dbID, _ := googleUuid.NewV7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite for concurrent operations (WAL mode, busy timeout).
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()

		return nil, nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	return db, sqlDB, nil
}

// ========================================
// Usage Example: Health Checks
// ========================================

//nolint:unused // Example function for documentation
func _exampleHealthChecks() {
	// Assuming _testListener is initialized in TestMain.
	// Liveness check (lightweight - process alive?).
	liveResult, err := cryptoutilAppsTemplateServiceServerListener.SendLivenessCheck(_testListener.Config())
	if err != nil {
		fmt.Printf("Liveness check failed: %v\n", err)
	} else {
		fmt.Printf("Liveness: %s\n", liveResult)
	}

	// Readiness check (heavyweight - dependencies healthy?).
	readyResult, err := cryptoutilAppsTemplateServiceServerListener.SendReadinessCheck(_testListener.Config())
	if err != nil {
		fmt.Printf("Readiness check failed: %v\n", err)
	} else {
		fmt.Printf("Readiness: %s\n", readyResult)
	}

	// Graceful shutdown via API.
	err = cryptoutilAppsTemplateServiceServerListener.SendShutdownRequest(_testListener.Config())
	if err != nil {
		fmt.Printf("Shutdown request failed: %v\n", err)
	} else {
		fmt.Println("Shutdown initiated")
	}
}

// ========================================
// Usage Example: Individual Tests
// ========================================

func TestSomething(t *testing.T) {
	t.Parallel()

	// Tests now have access to:
	// - baseURL: Full public server URL with actual port
	// - adminURL: Full admin server URL (always 127.0.0.1:9090 in production, dynamic in tests)
	// - testListener: For advanced operations (shutdown, port inspection)

	// Example: Make request to public API.
	// resp, err := http.Get(baseURL + "/api/v1/messages")

	// Example: Check liveness endpoint.
	// resp, err := http.Get(adminURL + "/admin/v1/livez")

	// Example: Graceful shutdown from test.
	// testListener.Shutdown()
}

// ========================================
// Key Differences Summary
// ========================================

/*
BEFORE (150+ lines):
- Manual SQLite configuration (20 lines)
- Manual GORM setup (10 lines)
- Manual migrations (5 lines)
- Manual telemetry initialization (10 lines)
- Manual JWK Gen Service (10 lines)
- Manual TLS config generation (10 lines)
- Manual server creation (30 lines)
- Manual defer cleanup (15 lines)
- Manual URL construction (5 lines)
- Manual timing/logging (5 lines)

AFTER (30 lines):
- createInMemoryDB helper (1 line call)
- ApplyMigrations (1 line)
- ApplicationConfig creation (7 lines)
- StartApplicationListener (1 line!)
- defer Shutdown (1 line)
- URL extraction (2 lines)
- os.Exit(m.Run()) (1 line)

BENEFITS:
1. Consistency: Same pattern across ALL services
2. Readability: Clear intent, minimal boilerplate
3. Maintainability: Infrastructure changes isolated to ApplicationListener
4. Testability: Easy to mock ApplicationListener for meta-testing
5. Production-Ready: Same code path for test and production
*/
