// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

var (
	testDB             *gorm.DB
	testSQLDB          *sql.DB // CRITICAL: Keep reference to prevent GC - in-memory SQLite requires open connection
	testJWKGenService  *cryptoutilSharedCryptoJose.JWKGenService
	testBarrierService *cryptoutilAppsTemplateServiceServerBarrier.Service
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup: Create shared heavyweight resources ONCE.
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	// CRITICAL: Store sql.DB reference in package variable.
	// In-memory SQLite databases are destroyed when all connections close.
	// Storing reference prevents GC from closing connection during parallel test execution.
	var err error

	testSQLDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	defer func() {
		if err := testSQLDB.Close(); err != nil {
			panic("TestMain: failed to close SQLite: " + err.Error())
		}
	}()

	// Configure SQLite for concurrent operations.
	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Run migrations.
	if err := ApplyCipherIMMigrations(testSQLDB, DatabaseTypeSQLite); err != nil {
		panic("TestMain: failed to run migrations: " + err.Error())
	}

	// Initialize telemetry.
	telemetrySettings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	testTelemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings)
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK Generation Service.
	testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK service: " + err.Error())
	}
	defer testJWKGenService.Shutdown()

	// Initialize Barrier Service.
	// Generate a simple test unseal key using JWE with A256GCM encryption and A256KW key wrapping.
	_, testUnsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate test unseal JWK: " + err.Error())
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal keys service: " + err.Error())
	}
	defer unsealKeysService.Shutdown()

	barrierRepo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	testBarrierService, err = cryptoutilAppsTemplateServiceServerBarrier.NewService(ctx, testTelemetryService, testJWKGenService, barrierRepo, unsealKeysService)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}
	defer testBarrierService.Shutdown()

	// Run all tests.
	exitCode := m.Run()

	// Cleanup happens via defer.
	os.Exit(exitCode)
}
