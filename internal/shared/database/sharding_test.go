// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgresModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // SQLite driver.
)

var testPostgresDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Try to start PostgreSQL testcontainer for schema-level tests.
	// Gracefully fallback to nil if Docker is unavailable.
	func() {
		defer func() {
			if r := recover(); r != nil {
				// testcontainers panics when Docker Desktop not running.
				testPostgresDB = nil
			}
		}()

		dbName := fmt.Sprintf("testdb_%s", googleUuid.Must(googleUuid.NewV7()))
		userName := fmt.Sprintf("user_%s", googleUuid.Must(googleUuid.NewV7()))

		container, err := postgresModule.Run(ctx,
			"postgres:18-alpine",
			postgresModule.WithDatabase(dbName),
			postgresModule.WithUsername(userName),
			postgresModule.WithPassword("testpassword"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds*time.Second),
			),
		)
		if err != nil {
			testPostgresDB = nil

			return
		}

		connStr, connErr := container.ConnectionString(ctx, "sslmode=disable")
		if connErr != nil {
			testPostgresDB = nil

			return
		}

		pgDB, openErr := gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if openErr != nil {
			testPostgresDB = nil

			return
		}

		testPostgresDB = pgDB
	}()

	os.Exit(m.Run())
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
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
	require.Equal(t, cryptoutilSharedMagic.SubjectTypePublic, cfg.DefaultSchema)
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
	cfg.Strategy = ShardStrategy(99) // Unknown strategy.
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

	if testPostgresDB == nil {
		t.Skip("Schema-level sharding requires PostgreSQL - Docker unavailable")
	}

	cfg := &ShardConfig{
		Strategy:        StrategySchemaLevel,
		SchemaPrefix:    "tenant_",
		DefaultSchema:   cryptoutilSharedMagic.SubjectTypePublic,
		EnableMigration: true,
	}
	m := NewShardManager(testPostgresDB, cfg)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	t.Run("creates schema and returns DB", func(t *testing.T) {
		tenantDB, err := m.GetDB(ctx)
		require.NoError(t, err)
		require.NotNil(t, tenantDB)
	})

	t.Run("returns cached DB on second call", func(t *testing.T) {
		tenantDB1, err := m.GetDB(ctx)
		require.NoError(t, err)

		tenantDB2, err := m.GetDB(ctx)
		require.NoError(t, err)

		// Both should be non-nil DB instances.
		require.NotNil(t, tenantDB1)
		require.NotNil(t, tenantDB2)
	})

	t.Run("no migration when disabled", func(t *testing.T) {
		cfg2 := &ShardConfig{
			Strategy:        StrategySchemaLevel,
			SchemaPrefix:    "nomig_",
			DefaultSchema:   cryptoutilSharedMagic.SubjectTypePublic,
			EnableMigration: false,
		}
		m2 := NewShardManager(testPostgresDB, cfg2)
		tenantID2 := googleUuid.Must(googleUuid.NewV7())
		ctx2 := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID2})

		// Without migration, schema doesn't exist - but SET search_path still works.
		tenantDB, err := m2.GetDB(ctx2)
		require.NoError(t, err)
		require.NotNil(t, tenantDB)
	})
}

func TestShardManager_DropTenantSchema(t *testing.T) {
	t.Parallel()

	if testPostgresDB == nil {
		t.Skip("Schema operations require PostgreSQL - Docker unavailable")
	}

	cfg := &ShardConfig{
		Strategy:        StrategySchemaLevel,
		SchemaPrefix:    "droptenant_",
		DefaultSchema:   cryptoutilSharedMagic.SubjectTypePublic,
		EnableMigration: true,
	}
	m := NewShardManager(testPostgresDB, cfg)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	// First create the schema.
	_, err := m.GetDB(ctx)
	require.NoError(t, err)

	// Now drop it.
	err = m.DropTenantSchema(tenantID)
	require.NoError(t, err)

	// Dropping again (CASCADE, IF EXISTS) should not error.
	err = m.DropTenantSchema(tenantID)
	require.NoError(t, err)
}

func TestShardManager_GetDB_SchemaLevel_InvalidCache(t *testing.T) {
	t.Parallel()

	// Use SQLite — the corrupt cache test never queries the DB, it only checks
	// the type assertion after a direct schemaCache.Store of a non-*gorm.DB value.
	db := setupTestDB(t)

	cfg := &ShardConfig{
		Strategy:        StrategySchemaLevel,
		SchemaPrefix:    "invalid_cache_",
		DefaultSchema:   cryptoutilSharedMagic.SubjectTypePublic,
		EnableMigration: false, // No migration needed — cache is pre-populated.
	}
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	schemaName := cfg.SchemaPrefix + tenantID.String()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	// Inject a non-*gorm.DB value directly into the cache to trigger the type assertion failure.
	m.schemaCache.Store(schemaName, "not-a-gorm-db")

	_, err := m.GetDB(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid cached DB type")
}

func TestShardManager_GetDB_SchemaLevel_EnsureSchemaError(t *testing.T) {
	t.Parallel()

	// Use SQLite which does not support CREATE SCHEMA — triggers ensureSchema failure.
	db := setupTestDB(t)

	cfg := &ShardConfig{
		Strategy:        StrategySchemaLevel,
		SchemaPrefix:    "schema_err_",
		DefaultSchema:   cryptoutilSharedMagic.SubjectTypePublic,
		EnableMigration: true, // Forces ensureSchema call — fails on SQLite.
	}
	m := NewShardManager(db, cfg)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	_, err := m.GetDB(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create schema")
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

func TestValidateSchemaName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid alphanumeric", input: "tenant_abc123", wantErr: false},
		{name: "valid with hyphens", input: fmt.Sprintf("tenant-%s", googleUuid.Must(googleUuid.NewV7())), wantErr: false},
		{name: "valid underscores", input: "schema_prefix_test", wantErr: false},
		{name: "invalid dollar sign", input: "tenant$bad", wantErr: true},
		{name: "invalid space", input: "tenant bad", wantErr: true},
		{name: "invalid semicolon", input: "tenant;drop", wantErr: true},
		{name: "invalid quotes", input: `tenant"inject`, wantErr: true},
		{name: "empty string", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateSchemaName(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid schema name")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShardManager_GetDB_InvalidSchemaPrefix(t *testing.T) {
	t.Parallel()

	sm := &ShardManager{
		config: &ShardConfig{
			Strategy:     StrategySchemaLevel,
			SchemaPrefix: "invalid$prefix_",
		},
	}

	tenantID := googleUuid.New()
	ctx := WithTenantContext(context.Background(), &TenantContext{TenantID: tenantID})

	_, err := sm.GetDB(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid schema name")
}

func TestShardManager_DropTenantSchema_InvalidSchemaPrefix(t *testing.T) {
	t.Parallel()

	sm := &ShardManager{
		config: &ShardConfig{
			Strategy:     StrategySchemaLevel,
			SchemaPrefix: "invalid$prefix_",
		},
	}

	tenantID := googleUuid.New()
	err := sm.DropTenantSchema(tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid schema name")
}
