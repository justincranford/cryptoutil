// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

var (
	testCtx              context.Context
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testGormDB           *gorm.DB
	testSQLDB            *sql.DB // CRITICAL: Keep reference to prevent GC — in-memory SQLite requires open connection
	testOrmRepository    *OrmRepository
)

func TestMain(m *testing.M) {
	testCtx = context.Background()

	var err error

	testTelemetryService, err = cryptoutilSharedTelemetry.NewTelemetryService(testCtx, &cryptoutilSharedTelemetry.TelemetrySettings{
		LogLevel:        cryptoutilSharedMagic.DefaultLogLevelInfo,
		OTLPEnabled:     false,
		OTLPService:     "framework-orm-test",
		OTLPInstance:    "test",
		OTLPVersion:     "0.0.0-test",
		OTLPEnvironment: "test",
		OTLPHostname:    cryptoutilSharedMagic.DefaultOTLPHostnameDefault,
		OTLPEndpoint:    cryptoutilSharedMagic.DefaultOTLPEndpointDefault,
	})
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create TelemetryService")

	defer testTelemetryService.Shutdown()

	testJWKGenService, err = cryptoutilSharedCryptoJose.NewJWKGenService(testCtx, testTelemetryService, false)
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create JWKGenService")

	defer testJWKGenService.Shutdown()

	dbID, err := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to generate DB ID")

	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	testSQLDB, err = sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to open SQLite")

	defer func() { _ = testSQLDB.Close() }()

	_, _ = testSQLDB.Exec("PRAGMA journal_mode=WAL;")
	_, _ = testSQLDB.Exec("PRAGMA busy_timeout=30000;")
	testSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.DBMaxPingAttempts)
	testSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.DBMaxPingAttempts)

	testGormDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to open GORM")

	testOrmRepository, err = NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, false)
	cryptoutilSharedApperr.RequireNoError(err, "TestMain: failed to create OrmRepository")

	os.Exit(m.Run())
}
