// Copyright (c) 2025 Justin Cranford
//
//

package tenant

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

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

	require.Equal(t, DBType(cryptoutilSharedMagic.TestDatabaseSQLite), DBTypeSQLite)
	require.Equal(t, DBType(cryptoutilSharedMagic.DockerServicePostgres), DBTypePostgres)
}

func TestSchemaPrefix_Constant(t *testing.T) {
	t.Parallel()

	require.Equal(t, "tenant_", SchemaPrefix)
}

func TestNewSchemaManager_DBError(t *testing.T) {
	t.Parallel()

	// Create a GORM DB with a PreparedStmt connector wrapping nil sql.DB,
	// which causes db.DB() to return ErrInvalidDB.
	bareDB := &gorm.DB{
		Config: &gorm.Config{
			ConnPool: &gorm.PreparedStmtDB{},
		},
	}

	_, err := NewSchemaManager(bareDB, DBTypeSQLite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get sql.DB")
}

func TestListSQLiteSchemas_RowsErrError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)

	// Create some tenant schemas so rows iteration has data.
	err = sm.CreateSchema(context.Background(), SchemaName(testTenantUUID2))
	require.NoError(t, err)

	// Cancel context before listing to trigger rows.Err().
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = sm.ListSchemas(ctx)
	require.Error(t, err)
}

func TestListSQLiteSchemas_ClosedDBError(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	sm, err := NewSchemaManager(db, DBTypeSQLite)
	require.NoError(t, err)

	// Close the underlying SQL connection to force query failure.
	sqlDB, err := db.DB()
	require.NoError(t, err)

	err = sqlDB.Close()
	require.NoError(t, err)

	_, err = sm.ListSchemas(context.Background())
	require.Error(t, err)
}
