// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps/framework/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// initTestDB creates a test database for ServiceFramework tests.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Create unique in-memory database per test.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: keep connections alive.

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db
}

// defaultTestConfig creates minimal valid ServiceFrameworkServerSettings for tests.
func defaultTestConfig() *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings {
	return cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}

// TestNewServiceFramework_HappyPath tests successful ServiceFramework creation.
func TestNewServiceFramework_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceFramework(ctx, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, st)

	// Verify accessors.
	require.Equal(t, cfg, st.Config())
	require.Equal(t, db, st.DB())
	require.Equal(t, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite, st.DBType())
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())

	// Verify SQLDB accessor.
	sqlDB, err := st.SQLDB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)
}

// TestNewServiceFramework_NilContext tests constructor with nil context.
func TestNewServiceFramework_NilContext(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := defaultTestConfig()

	//nolint:staticcheck // Testing nil context validation.
	_, err := NewServiceFramework(nil, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewServiceFramework_NilConfig tests constructor with nil config.
func TestNewServiceFramework_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)

	_, err := NewServiceFramework(ctx, nil, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNewServiceFramework_NilDatabase tests constructor with nil database.
func TestNewServiceFramework_NilDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := defaultTestConfig()

	_, err := NewServiceFramework(ctx, cfg, nil, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database cannot be nil")
}

// TestNewServiceFramework_InvalidDatabaseType tests constructor with invalid database type.
func TestNewServiceFramework_InvalidDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	_, err := NewServiceFramework(ctx, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseType("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid database type")
}

// TestNewServiceFramework_PostgreSQLDatabaseType tests constructor with PostgreSQL database type.
func TestNewServiceFramework_PostgreSQLDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceFramework(ctx, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL)
	require.NoError(t, err)
	require.NotNil(t, st)
	require.Equal(t, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypePostgreSQL, st.DBType())
}

// TestServiceFramework_Shutdown tests graceful shutdown of all components.
func TestServiceFramework_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceFramework(ctx, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, st)

	// Verify components are initialized.
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())

	// Shutdown should not panic and should release resources.
	st.Shutdown()

	// After shutdown, components should still be accessible (not set to nil).
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())
}

// TestServiceFramework_Shutdown_NilComponents tests shutdown with nil components.
func TestServiceFramework_Shutdown_NilComponents(t *testing.T) {
	t.Parallel()

	// Create ServiceFramework with nil components (edge case).
	st := &ServiceFramework{
		telemetry: nil,
		jwkGen:    nil,
	}

	// Shutdown should not panic with nil components.
	st.Shutdown()
}

// TestServiceFramework_SQLDB_Error tests SQLDB accessor error handling.
func TestServiceFramework_SQLDB_Error(t *testing.T) {
	t.Parallel()

	// Create ServiceFramework with mock GORM DB that will fail.
	// This is hard to test without mocking since GORM.DB() rarely fails in practice.
	// For now, we test the happy path (covered in TestNewServiceFramework_HappyPath).
	// If we need error coverage, we'd need to use a mock framework.
	t.Skip("SQLDB error path requires mocking framework - happy path covered in HappyPath test")
}

// TestStartApplicationCore_PassThrough tests the wrapper function.
func TestStartApplicationCore_PassThrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use minimal test config.
	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	// StartApplicationCore should pass through to application.StartApplicationCore.
	core, err := StartApplicationCore(ctx, settings)
	require.NoError(t, err)
	require.NotNil(t, core)

	// Cleanup.
	defer core.Shutdown()

	// Verify core components initialized.
	require.NotNil(t, core.DB)
	require.NotNil(t, core.Basic)
}

// TestStartApplicationCore_NilContext tests wrapper with nil context.
func TestStartApplicationCore_NilContext(t *testing.T) {
	t.Parallel()

	settings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	//nolint:staticcheck // Testing nil context validation.
	_, err := StartApplicationCore(nil, settings)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ctx cannot be nil")
}

// TestStartApplicationCore_NilSettings tests wrapper with nil settings.
func TestStartApplicationCore_NilSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, err := StartApplicationCore(ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "settings cannot be nil")
}

// TestServiceFramework_AccessorMethods tests all accessor methods for coverage.
func TestServiceFramework_AccessorMethods(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceFramework(ctx, cfg, db, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, st)

	// Test Config() accessor.
	config := st.Config()
	require.NotNil(t, config)
	require.Equal(t, cfg, config)

	// Test DB() accessor.
	dbFromAccessor := st.DB()
	require.NotNil(t, dbFromAccessor)
	require.Equal(t, db, dbFromAccessor)

	// Test SQLDB() accessor.
	sqlDB, err := st.SQLDB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)

	// Test DBType() accessor.
	dbType := st.DBType()
	require.Equal(t, cryptoutilAppsFrameworkServiceServerRepository.DatabaseTypeSQLite, dbType)

	// Test Telemetry() accessor.
	telemetry := st.Telemetry()
	require.NotNil(t, telemetry)

	// Test JWKGen() accessor.
	jwkGen := st.JWKGen()
	require.NotNil(t, jwkGen)

	// Test Shutdown to cover nil checks for services (telemetry, jwkGen).
	st.Shutdown()
	// After shutdown, verify services are still accessible (shutdown is graceful, not destructive).
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())
}
