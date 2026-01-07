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

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTemplateServerRepository "cryptoutil/internal/apps/template/service/server/repository"
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

	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0) // In-memory: keep connections alive.

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db
}

// defaultTestConfig creates minimal valid ServerSettings for tests.
func defaultTestConfig() *cryptoutilConfig.ServerSettings {
	return cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true)
}

// TestNewServiceTemplate_HappyPath tests successful ServiceTemplate creation.
func TestNewServiceTemplate_HappyPath(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilTemplateServerRepository.DatabaseTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, st)

	// Verify accessors.
	require.Equal(t, cfg, st.Config())
	require.Equal(t, db, st.DB())
	require.Equal(t, cryptoutilTemplateServerRepository.DatabaseTypeSQLite, st.DBType())
	require.NotNil(t, st.Telemetry())
	require.NotNil(t, st.JWKGen())
	require.Nil(t, st.Barrier()) // No barrier option provided.

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
	_, err := NewServiceTemplate(nil, cfg, db, cryptoutilTemplateServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "context cannot be nil")
}

// TestNewServiceTemplate_NilConfig tests constructor with nil config.
func TestNewServiceTemplate_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)

	_, err := NewServiceTemplate(ctx, nil, db, cryptoutilTemplateServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "config cannot be nil")
}

// TestNewServiceTemplate_NilDatabase tests constructor with nil database.
func TestNewServiceTemplate_NilDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := defaultTestConfig()

	_, err := NewServiceTemplate(ctx, cfg, nil, cryptoutilTemplateServerRepository.DatabaseTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database cannot be nil")
}

// TestNewServiceTemplate_InvalidDatabaseType tests constructor with invalid database type.
func TestNewServiceTemplate_InvalidDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	_, err := NewServiceTemplate(ctx, cfg, db, cryptoutilTemplateServerRepository.DatabaseType("invalid"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid database type")
}

// TestNewServiceTemplate_PostgreSQLDatabaseType tests constructor with PostgreSQL database type.
func TestNewServiceTemplate_PostgreSQLDatabaseType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilTemplateServerRepository.DatabaseTypePostgreSQL)
	require.NoError(t, err)
	require.NotNil(t, st)
	require.Equal(t, cryptoutilTemplateServerRepository.DatabaseTypePostgreSQL, st.DBType())
}

// TestNewServiceTemplate_WithBarrierOption tests WithBarrier functional option.
func TestNewServiceTemplate_WithBarrierOption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := initTestDB(t)
	cfg := defaultTestConfig()

	// NOTE: BarrierService requires unseal key initialization.
	// For this test, we verify the option mechanism works (barrier will be nil until Phase 5b).
	// Phase 5b will add actual barrier service initialization.

	st, err := NewServiceTemplate(ctx, cfg, db, cryptoutilTemplateServerRepository.DatabaseTypeSQLite, WithBarrier(nil))
	require.NoError(t, err)
	require.NotNil(t, st)

	// Barrier is nil because we passed nil (Phase 5b will initialize properly).
	require.Nil(t, st.Barrier())
}
