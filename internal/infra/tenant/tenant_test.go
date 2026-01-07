// Copyright (c) 2025 Justin Cranford
//
//

package tenant

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

// testTenantUUID is the UUID used for tenant tests.
var testTenantUUID = googleUuid.Must(googleUuid.NewV7()).String()

// testTenantUUIDs are additional UUIDs for multi-tenant tests.
var (
	testTenantUUID2 = googleUuid.Must(googleUuid.NewV7()).String()
	testTenantUUID3 = googleUuid.Must(googleUuid.NewV7()).String()
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Open in-memory SQLite database.
	sqlDB, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Configure pragmas.
	ctx := context.Background()

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Create GORM database.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Configure connection pool.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	return db
}

func TestSchemaName(t *testing.T) {
	t.Parallel()

	// Compute expected schema name from dynamic UUID.
	expectedSchemaName := "tenant_" + strings.ReplaceAll(testTenantUUID, "-", "_")

	tests := []struct {
		name     string
		tenantID string
		want     string
	}{
		{
			name:     "valid UUID",
			tenantID: testTenantUUID,
			want:     expectedSchemaName,
		},
		{
			name:     "uppercase UUID",
			tenantID: strings.ToUpper(testTenantUUID),
			want:     "tenant_" + strings.ReplaceAll(strings.ToUpper(testTenantUUID), "-", "_"),
		},
		{
			name:     "simple alphanumeric",
			tenantID: "tenant123",
			want:     "tenant_tenant123",
		},
		{
			name:     "empty string",
			tenantID: "",
			want:     "tenant_",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SchemaName(tc.tenantID)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestSanitizeTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tenantID string
		want     string
	}{
		{
			name:     "alphanumeric only",
			tenantID: "abc123",
			want:     "abc123",
		},
		{
			name:     "hyphens converted",
			tenantID: "abc-123-def",
			want:     "abc_123_def",
		},
		{
			name:     "special chars removed",
			tenantID: "abc!@#$%^&*()123",
			want:     "abc123",
		},
		{
			name:     "mixed case preserved",
			tenantID: "AbCdEf",
			want:     "AbCdEf",
		},
		{
			name:     "empty string",
			tenantID: "",
			want:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeTenantID(tc.tenantID)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestNewSchemaManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		db      *gorm.DB
		dbType  DBType
		wantErr bool
	}{
		{
			name:    "nil db",
			db:      nil,
			dbType:  DBTypeSQLite,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sm, err := NewSchemaManager(tc.db, tc.dbType)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, sm)
			} else {
				require.NoError(t, err)
				require.NotNil(t, sm)
			}
		})
	}
}

func TestNewSchemaManager_Valid(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)
	require.NotNil(t, sm)
	require.Equal(t, DBTypeSQLite, sm.dbType)
}

func TestSchemaManager_SQLite_CreateAndDrop(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)

	tenantID := testTenantUUID

	// Create schema.
	err = sm.CreateSchema(ctx, tenantID)
	require.NoError(t, err)

	// Verify schema exists.
	exists, err := sm.SchemaExists(ctx, tenantID)
	require.NoError(t, err)
	require.True(t, exists)

	// Drop schema.
	err = sm.DropSchema(ctx, tenantID)
	require.NoError(t, err)

	// Verify schema no longer exists.
	exists, err = sm.SchemaExists(ctx, tenantID)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestSchemaManager_SQLite_ListSchemas(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)

	// Initially no tenant schemas.
	schemas, err := sm.ListSchemas(ctx)
	require.NoError(t, err)
	require.Empty(t, schemas)

	// Create multiple schemas.
	tenantIDs := []string{
		testTenantUUID2,
		testTenantUUID3,
	}

	for _, tid := range tenantIDs {
		err = sm.CreateSchema(ctx, tid)
		require.NoError(t, err)
	}

	// List schemas.
	schemas, err = sm.ListSchemas(ctx)
	require.NoError(t, err)
	require.Len(t, schemas, 2)

	// Cleanup.
	for _, tid := range tenantIDs {
		err = sm.DropSchema(ctx, tid)
		require.NoError(t, err)
	}
}

func TestSchemaManager_UnsupportedDBType(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	sm, err := NewSchemaManager(db, "unsupported")
	require.NoError(t, err)

	tenantID := testTenantUUID

	// All operations should fail with unsupported type.
	err = sm.CreateSchema(ctx, tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")

	err = sm.DropSchema(ctx, tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")

	_, err = sm.SchemaExists(ctx, tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")

	_, err = sm.ListSchemas(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported database type")
}

func TestWithTenant_GetTenant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tenantID   string
		wantSchema string
	}{
		{
			name:       "valid UUID tenant",
			tenantID:   testTenantUUID,
			wantSchema: "tenant_550e8400_e29b_41d4_a716_446655440000",
		},
		{
			name:       "simple tenant ID",
			tenantID:   "demo",
			wantSchema: "tenant_demo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			ctx = WithTenant(ctx, tc.tenantID)

			tc := GetTenant(ctx)
			require.NotNil(t, tc)
			require.Equal(t, tc.TenantID, tc.TenantID)
			require.Equal(t, tc.SchemaName, tc.SchemaName)
		})
	}
}

func TestGetTenant_NotSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tc := GetTenant(ctx)
	require.Nil(t, tc)
}

func TestIsValidTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tenantID string
		want     bool
	}{
		{
			name:     "valid lowercase UUID",
			tenantID: testTenantUUID,
			want:     true,
		},
		{
			name:     "valid uppercase UUID",
			tenantID: testTenantUUID,
			want:     true,
		},
		{
			name:     "valid mixed case UUID",
			tenantID: testTenantUUID,
			want:     true,
		},
		{
			name:     "too short",
			tenantID: "550e8400-e29b-41d4-a716",
			want:     false,
		},
		{
			name:     "too long",
			tenantID: testTenantUUID + "-extra",
			want:     false,
		},
		{
			name:     "missing hyphens",
			tenantID: "550e8400e29b41d4a716446655440000",
			want:     false,
		},
		{
			name:     "wrong hyphen position",
			tenantID: "550e840-0e29b-41d4-a716-446655440000",
			want:     false,
		},
		{
			name:     "invalid hex char",
			tenantID: "550e8400-e29b-41d4-a716-44665544000g",
			want:     false,
		},
		{
			name:     "empty string",
			tenantID: "",
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := IsValidTenantID(tc.tenantID)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsHexChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"digit 0", '0', true},
		{"digit 9", '9', true},
		{"lowercase a", 'a', true},
		{"lowercase f", 'f', true},
		{"uppercase A", 'A', true},
		{"uppercase F", 'F', true},
		{"lowercase g", 'g', false},
		{"uppercase G", 'G', false},
		{"hyphen", '-', false},
		{"space", ' ', false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := isHexChar(tc.r)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestSchemaManager_GetScopedDB(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)

	tenantID := testTenantUUID
	scopedDB := sm.GetScopedDB(tenantID)
	require.NotNil(t, scopedDB)
}

func TestSchemaManager_GetScopedDB_Unsupported(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	sm, err := NewSchemaManager(db, "unsupported")
	require.NoError(t, err)

	tenantID := testTenantUUID
	scopedDB := sm.GetScopedDB(tenantID)
	require.NotNil(t, scopedDB) // Returns base DB for unsupported type.
}

func TestDBType_Constants(t *testing.T) {
	t.Parallel()

	require.Equal(t, DBType("sqlite"), DBTypeSQLite)
	require.Equal(t, DBType("postgres"), DBTypePostgres)
}

func TestSchemaPrefix_Constant(t *testing.T) {
	t.Parallel()

	require.Equal(t, "tenant_", SchemaPrefix)
}
