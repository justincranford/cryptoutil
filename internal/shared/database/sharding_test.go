// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package database

import (
	"context"
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // SQLite driver
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return db
}

func TestDefaultShardConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultShardConfig()
	require.Equal(t, StrategyRowLevel, cfg.Strategy)
	require.Equal(t, "tenant_", cfg.SchemaPrefix)
	require.Equal(t, "public", cfg.DefaultSchema)
	require.True(t, cfg.EnableMigration)
}

func TestNewShardManager(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	t.Run("with config", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultShardConfig()
		m := NewShardManager(db, cfg)
		require.NotNil(t, m)
		require.Equal(t, cfg, m.config)
	})

	t.Run("nil config uses default", func(t *testing.T) {
		t.Parallel()

		m := NewShardManager(db, nil)
		require.NotNil(t, m)
		require.NotNil(t, m.config)
		require.Equal(t, StrategyRowLevel, m.config.Strategy)
	})
}

func TestShardManager_GetDBNContext(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	cfg := DefaultShardConfig()
	m := NewShardManager(db, cfg)

	t.Run("no tenant context", func(t *testing.T) {
		t.Parallel()

		_, err := m.GetDB(context.Background())
		require.ErrorIs(t, err, ErrNoTenantContext)
	})

	t.Run("nil tenant id", func(t *testing.T) {
		t.Parallel()

		ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: googleUuid.Nil})
		_, err := m.GetDB(ctx)
		require.ErrorIs(t, err, ErrInvalidTenantID)
	})
}

func TestShardManager_GetDB_RowLevel(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	cfg := DefaultShardConfig()
	cfg.Strategy = StrategyRowLevel
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.New()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	tenantDB, err := m.GetDB(ctx)
	require.NoError(t, err)
	require.NotNil(t, tenantDB)
}

func TestShardManager_GetDB_DatabaseLevel(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	cfg := DefaultShardConfig()
	cfg.Strategy = StrategyDatabaseLevel
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.New()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	_, err := m.GetDB(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "database-level sharding not yet implemented")
}

func TestShardManager_GetDB_UnknownStrategy(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	cfg := DefaultShardConfig()
	cfg.Strategy = ShardStrategy(99) // Unknown strategy
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.New()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	_, err := m.GetDB(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown shard strategy")
}

func TestShardManager_GetTenantSchemaName(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	cfg := DefaultShardConfig()
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	name := m.GetTenantSchemaName(tenantID)
	require.Equal(t, "tenant_"+tenantID.String(), name)
}

func TestShardManager_GetDB_SchemaLevel(t *testing.T) {
	t.Parallel()
	// Skip: SQLite does not support CREATE SCHEMA, this is PostgreSQL-specific
	t.Skip("Schema-level sharding requires PostgreSQL")
}

func TestShardManager_DropTenantSchema(t *testing.T) {
	t.Parallel()
	// Skip: SQLite does not support DROP SCHEMA, this is PostgreSQL-specific
	t.Skip("Schema operations require PostgreSQL")
}

func TestShardStrategyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		strategy ShardStrategy
		expected string
	}{
		{StrategyRowLevel, "row-level"},
		{StrategySchemaLevel, "schema-level"},
		{StrategyDatabaseLevel, "database-level"},
		{ShardStrategy(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.strategy.String())
		})
	}
}
