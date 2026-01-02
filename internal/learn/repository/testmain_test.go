// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilTemplateBarrier "cryptoutil/internal/template/server/barrier"
)

var (
	testDB             *gorm.DB
	testSQLDB          *sql.DB // CRITICAL: Keep reference to prevent GC - in-memory SQLite requires open connection
	testBarrierService *cryptoutilTemplateBarrier.BarrierService
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
	if err := ApplyMigrations(testSQLDB, DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to run migrations: " + err.Error())
	}

	// Initialize telemetry.
	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-repository-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://" + cryptoutilMagic.HostnameLocalhost + ":4317",
	}

	testTelemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}

	// Initialize JWK Generation Service.
	testJWKGenService, err := cryptoutilJose.NewJWKGenService(ctx, testTelemetryService, false)
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

	// Run all tests.
	exitCode := m.Run()

	// Cleanup happens via defer.
	os.Exit(exitCode)
}
