// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite" // Use modernc CGO-free SQLite.

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Test UUID generated once per test run for consistency.
var tenantTestUUID = googleUuid.Must(googleUuid.NewV7()).String()

func setupTenantTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique database name per test to avoid conflicts.
	dbName := fmt.Sprintf("file:tenant_test_%d?mode=memory&cache=private", time.Now().UTC().UnixNano())

	// Use database/sql with modernc.org/sqlite driver.
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dbName)
	require.NoError(t, err)

	// Configure connection pool per instructions.
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)

	// Wrap with GORM using sqlite Dialector.
	db, err := gorm.Open(gormsqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db
}

func TestNewTenantManager(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		db      *gorm.DB
		config  *TenantManagerConfig
		wantErr bool
	}{
		{
			name:    "nil database",
			db:      nil,
			config:  nil,
			wantErr: true,
		},
		{
			name:    "valid database with nil config",
			db:      setupTenantTestDB(t),
			config:  nil,
			wantErr: false,
		},
		{
			name: "valid database with config",
			db:   setupTenantTestDB(t),
			config: &TenantManagerConfig{
				IsolationMode:  TenantIsolationRow,
				DefaultRealmID: "default-realm",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			manager, err := NewTenantManager(tc.db, tc.config)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, manager)
			} else {
				require.NoError(t, err)
				require.NotNil(t, manager)
			}
		})
	}
}

func TestTenantManager_RegisterTenant(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode:  TenantIsolationRow,
		DefaultRealmID: "default-realm",
	})
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		tenant  *TenantConfig
		wantErr bool
	}{
		{
			name:    "nil tenant",
			tenant:  nil,
			wantErr: true,
		},
		{
			name:    "empty tenant ID",
			tenant:  &TenantConfig{Name: "test"},
			wantErr: true,
		},
		{
			name:    "empty tenant name",
			tenant:  &TenantConfig{ID: "tenant-1"},
			wantErr: true,
		},
		{
			name: "valid tenant",
			tenant: &TenantConfig{
				ID:      "tenant-1",
				Name:    "Test Tenant",
				Enabled: true,
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.RegisterTenant(ctx, tc.tenant)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify defaults applied.
				tenant, ok := manager.GetTenant(tc.tenant.ID)
				require.True(t, ok)
				require.Equal(t, TenantIsolationRow, tenant.IsolationMode)
				require.Equal(t, "default-realm", tenant.RealmID)
			}
		})
	}
}

func TestTenantManager_RegisterTenant_Duplicate(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	ctx := context.Background()

	tenant := &TenantConfig{
		ID:      "dup-tenant",
		Name:    "Duplicate Tenant",
		Enabled: true,
	}

	// First registration should succeed.
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Second registration should fail.
	err = manager.RegisterTenant(ctx, tenant)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestTenantManager_GetTenant(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Register tenant.
	tenant := &TenantConfig{
		ID:      "get-tenant",
		Name:    "Get Tenant",
		Enabled: true,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Found.
	found, ok := manager.GetTenant("get-tenant")
	require.True(t, ok)
	require.Equal(t, "Get Tenant", found.Name)

	// Not found.
	notFound, ok := manager.GetTenant("nonexistent")
	require.False(t, ok)
	require.Nil(t, notFound)
}

func TestTenantManager_ListTenants(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Register multiple tenants.
	for i := 0; i < 3; i++ {
		tenant := &TenantConfig{
			ID:      fmt.Sprintf("list-tenant-%d", i),
			Name:    fmt.Sprintf("List Tenant %d", i),
			Enabled: true,
		}
		err = manager.RegisterTenant(ctx, tenant)
		require.NoError(t, err)
	}

	// List tenants.
	tenants := manager.ListTenants()
	require.Len(t, tenants, 3)
}

func TestTenantManager_DeleteTenant(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Register tenant.
	tenant := &TenantConfig{
		ID:      "delete-tenant",
		Name:    "Delete Tenant",
		Enabled: true,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Delete tenant.
	err = manager.DeleteTenant(ctx, "delete-tenant")
	require.NoError(t, err)

	// Verify deleted.
	_, ok := manager.GetTenant("delete-tenant")
	require.False(t, ok)

	// Delete nonexistent tenant.
	err = manager.DeleteTenant(ctx, "nonexistent")
	require.Error(t, err)
}

func TestTenantManager_WithTenant_Row(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationRow,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Register tenant.
	tenant := &TenantConfig{
		ID:      "with-tenant",
		Name:    "With Tenant",
		Enabled: true,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Get scoped DB.
	scopedDB, err := manager.WithTenant(ctx, "with-tenant")
	require.NoError(t, err)
	require.NotNil(t, scopedDB)

	// Nonexistent tenant.
	_, err = manager.WithTenant(ctx, "nonexistent")
	require.Error(t, err)
}

func TestTenantManager_WithTenant_Disabled(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Register disabled tenant.
	tenant := &TenantConfig{
		ID:      "disabled-tenant",
		Name:    "Disabled Tenant",
		Enabled: false,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Should fail for disabled tenant.
	_, err = manager.WithTenant(ctx, "disabled-tenant")
	require.Error(t, err)
	require.Contains(t, err.Error(), cryptoutilSharedMagic.DefaultDatabaseContainerDisabled)
}

func TestTenantManager_WithTenant_Schema(t *testing.T) {
	t.Parallel()

	// Schema isolation requires PostgreSQL (CREATE SCHEMA not supported in SQLite).
	t.Skip("Schema isolation requires PostgreSQL - not supported in SQLite test DB")
}

func TestTenantManager_WithTenant_DatabaseIsolation(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationDatabase,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Register tenant.
	tenant := &TenantConfig{
		ID:      "db-tenant",
		Name:    "Database Tenant",
		Enabled: true,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Database isolation not implemented - should return error.
	_, err = manager.WithTenant(ctx, "db-tenant")
	require.Error(t, err)
	require.Contains(t, err.Error(), "not implemented")
}

func TestTenantContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// No tenant in context.
	tenant, ok := TenantFromContext(ctx)
	require.False(t, ok)
	require.Nil(t, tenant)

	// Add tenant to context.
	tenantCtx := &TenantContext{
		TenantID:   "ctx-tenant",
		TenantName: "Context Tenant",
		RealmID:    "realm-1",
	}
	ctx = ContextWithTenant(ctx, tenantCtx)

	// Retrieve tenant from context.
	found, ok := TenantFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "ctx-tenant", found.TenantID)
	require.Equal(t, "Context Tenant", found.TenantName)
	require.Equal(t, "realm-1", found.RealmID)
}

func TestValidateTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tenantID string
		wantErr  bool
	}{
		{
			name:     "empty",
			tenantID: "",
			wantErr:  true,
		},
		{
			name:     "too short",
			tenantID: "short",
			wantErr:  true,
		},
		{
			name:     "valid UUID",
			tenantID: tenantTestUUID,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateTenantID(tc.tenantID)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSanitizeSchemaName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "simple name",
			input:  "tenant_abc",
			expect: "tenant_abc",
		},
		{
			name:   "with special characters",
			input:  "tenant-abc.def",
			expect: "tenant_abc_def",
		},
		{
			name:   "starts with number",
			input:  "123tenant",
			expect: "t_123tenant",
		},
		{
			name:   "uppercase",
			input:  "Tenant_ABC",
			expect: "tenant_abc",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := sanitizeSchemaName(tc.input)
			require.Equal(t, tc.expect, result)
		})
	}
}
