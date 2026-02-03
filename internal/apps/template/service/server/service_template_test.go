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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// initTestDB creates a test database for ServiceTemplate tests.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Create unique in-memory database per test.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := "file:" + dbID.String() + "?mode=memory&cache=private"

	sqlDB, err := sql.Open("sqlite", dsn)
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

// defaultTestConfig creates minimal valid ServiceTemplateServerSettings for tests.
func defaultTestConfig() *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings {
	return cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
}

// TestNewServiceTemplate_HappyPath tests successful ServiceTemplate creation.
func TestNewServiceTemplate_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, st)

	// Verify accessors.
	require.Equal(t, cfg, st.Config())
	require.Equal(t, db, st.DB())
	require.Equal(t, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite, st.DBType())
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())

	// Verify SQLDB accessor.
	sqlDB, err := st.SQLDB()
	require.NoError(t, err)
	require.NotNil(t, sqlDB)
}

// TestNewServiceTemplate_NilContext tests constructor with nil context.
func TestNewServiceTemplate_NilContext(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	cfg := defaultTestConfig()

	//nolint:staticcheck // Testing nil context validation.
	_, err := NewServiceTemplate(nil, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewServiceTemplate_NilConfig tests constructor with nil config.
func TestNewServiceTemplate_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)

	_, err := NewServiceTemplate(ctx, nil, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNewServiceTemplate_NilDatabase tests constructor with nil database.
func TestNewServiceTemplate_NilDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := defaultTestConfig()

	_, err := NewServiceTemplate(ctx, cfg, nil, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database cannot be nil")
}

// TestNewServiceTemplate_InvalidDatabaseType tests constructor with invalid database type.
func TestNewServiceTemplate_InvalidDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	_, err := NewServiceTemplate(ctx, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseType("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid database type")
}

// TestNewServiceTemplate_PostgreSQLDatabaseType tests constructor with PostgreSQL database type.
func TestNewServiceTemplate_PostgreSQLDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL)
	require.NoError(t, err)
	require.NotNil(t, st)
	require.Equal(t, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypePostgreSQL, st.DBType())
}

// TestServiceTemplate_Shutdown tests graceful shutdown of all components.
func TestServiceTemplate_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
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

// TestServiceTemplate_Shutdown_NilComponents tests shutdown with nil components.
func TestServiceTemplate_Shutdown_NilComponents(t *testing.T) {
	t.Parallel()

	// Create ServiceTemplate with nil components (edge case).
	st := &ServiceTemplate{
		telemetry: nil,
		jwkGen:    nil,
	}

	// Shutdown should not panic with nil components.
	st.Shutdown()
}

// TestServiceTemplate_SQLDB_Error tests SQLDB accessor error handling.
func TestServiceTemplate_SQLDB_Error(t *testing.T) {
	t.Parallel()

	// Create ServiceTemplate with mock GORM DB that will fail.
	// This is hard to test without mocking since GORM.DB() rarely fails in practice.
	// For now, we test the happy path (covered in TestNewServiceTemplate_HappyPath).
	// If we need error coverage, we'd need to use a mock framework.
	t.Skip("SQLDB error path requires mocking framework - happy path covered in HappyPath test")
}

// TestStartApplicationCore_PassThrough tests the wrapper function.
func TestStartApplicationCore_PassThrough(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Use minimal test config.
	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

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

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

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

// TestServiceTemplate_AccessorMethods tests all accessor methods for coverage.
func TestServiceTemplate_AccessorMethods(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite)
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
	require.Equal(t, cryptoutilAppsTemplateServiceServerRepository.DatabaseTypeSQLite, dbType)

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
