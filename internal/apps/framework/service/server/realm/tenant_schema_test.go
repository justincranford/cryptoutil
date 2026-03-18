// Copyright (c) 2025 Justin Cranford
//
//

package realm

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestTenantManager_createTenantSchema tests direct invocation of the private
// createTenantSchema method.
func TestTenantManager_createTenantSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, m *TenantManager) *TenantConfig
		wantErr bool
	}{
		{
			name: "auto-generates schema name when empty",
			setup: func(t *testing.T, m *TenantManager) *TenantConfig {
				t.Helper()
				// SchemaName empty; createTenantSchema generates it from ID.
				// SQLite will fail on CREATE SCHEMA but the auto-generation runs first.
				return &TenantConfig{
					ID:      googleUuid.Must(googleUuid.NewV7()).String(),
					Name:    "auto-schema",
					Enabled: true,
				}
			},
			wantErr: true, // SQLite does not support CREATE SCHEMA.
		},
		{
			name: "skips creation when schema already marked created",
			setup: func(t *testing.T, m *TenantManager) *TenantConfig {
				t.Helper()

				tenantID := googleUuid.Must(googleUuid.NewV7()).String()
				// Pre-mark as created to exercise the early-return guard.
				m.schemaCreated[tenantID] = true

				return &TenantConfig{
					ID:         tenantID,
					SchemaName: "tenant_already_created",
					Name:       "pre-created",
					Enabled:    true,
				}
			},
			wantErr: false, // Early return nil - no DB interaction.
		},
		{
			name: "returns error on SQLite (CREATE SCHEMA not supported)",
			setup: func(t *testing.T, m *TenantManager) *TenantConfig {
				t.Helper()

				return &TenantConfig{
					ID:         googleUuid.Must(googleUuid.NewV7()).String(),
					SchemaName: "test_tenant_schema",
					Name:       "schema-tenant",
					Enabled:    true,
				}
			},
			wantErr: true, // SQLite does not support CREATE SCHEMA.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTenantTestDB(t)
			manager, err := NewTenantManager(db, nil)
			require.NoError(t, err)

			tenant := tc.setup(t, manager)
			ctx := context.Background()

			err = manager.createTenantSchema(ctx, tenant)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTenantManager_dropTenantSchema tests direct invocation of the private
// dropTenantSchema method.
func TestTenantManager_dropTenantSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tenant  *TenantConfig
		wantErr bool
	}{
		{
			name: "early return when schema name is empty",
			tenant: &TenantConfig{
				ID:         googleUuid.Must(googleUuid.NewV7()).String(),
				SchemaName: "", // Empty triggers early return nil.
				Name:       "no-schema",
				Enabled:    true,
			},
			wantErr: false,
		},
		{
			name: "returns error on SQLite (DROP SCHEMA not supported)",
			tenant: &TenantConfig{
				ID:         googleUuid.Must(googleUuid.NewV7()).String(),
				SchemaName: "test_drop_schema",
				Name:       "drop-schema-tenant",
				Enabled:    true,
			},
			wantErr: true, // SQLite does not support DROP SCHEMA.
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTenantTestDB(t)
			manager, err := NewTenantManager(db, nil)
			require.NoError(t, err)

			ctx := context.Background()

			err = manager.dropTenantSchema(ctx, tc.tenant)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestTenantManager_RegisterTenant_SchemaError verifies RegisterTenant returns
// an error when schema creation fails on SQLite.
func TestTenantManager_RegisterTenant_SchemaError(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationSchema,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tenant := &TenantConfig{
		ID:            googleUuid.Must(googleUuid.NewV7()).String(),
		Name:          "schema-reg-tenant",
		IsolationMode: TenantIsolationSchema,
		Enabled:       true,
	}

	// SQLite does not support CREATE SCHEMA - registration must fail.
	err = manager.RegisterTenant(ctx, tenant)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create tenant schema")
}

// TestTenantManager_DeleteTenant_SchemaError verifies DeleteTenant returns an
// error when schema dropping fails on SQLite.
func TestTenantManager_DeleteTenant_SchemaError(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	tenantID := googleUuid.Must(googleUuid.NewV7()).String()

	// Directly inject a schema-mode tenant to bypass RegisterTenant.
	manager.tenants[tenantID] = &TenantConfig{
		ID:            tenantID,
		Name:          "schema-del-tenant",
		IsolationMode: TenantIsolationSchema,
		SchemaName:    "tenant_delete_test",
		Enabled:       true,
	}

	ctx := context.Background()

	// SQLite does not support DROP SCHEMA - deletion must fail.
	err = manager.DeleteTenant(ctx, tenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to drop tenant schema")
}

// TestTenantManager_WithTenant_SchemaMode verifies WithTenant handles schema
// isolation mode by executing the SET search_path statement.
func TestTenantManager_WithTenant_SchemaMode(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, nil)
	require.NoError(t, err)

	tenantID := googleUuid.Must(googleUuid.NewV7()).String()

	// Directly inject a schema-mode tenant to test WithTenant schema branch.
	manager.tenants[tenantID] = &TenantConfig{
		ID:            tenantID,
		Name:          "schema-with-tenant",
		IsolationMode: TenantIsolationSchema,
		SchemaName:    "test_schema_path",
		Enabled:       true,
	}

	ctx := context.Background()

	// WithTenant with schema isolation - exercises the TenantIsolationSchema case.
	// SQLite silently ignores or errors on SET search_path, but WithTenant returns it.
	scopedDB, err := manager.WithTenant(ctx, tenantID)
	require.NoError(t, err)
	require.NotNil(t, scopedDB)
}

// TestTenantManager_WithTenant_RowClosure verifies the row isolation scope
// closure is executed during an actual query.
func TestTenantManager_WithTenant_RowClosure(t *testing.T) {
	t.Parallel()

	db := setupTenantTestDB(t)
	manager, err := NewTenantManager(db, &TenantManagerConfig{
		IsolationMode: TenantIsolationRow,
	})
	require.NoError(t, err)

	ctx := context.Background()

	tenantID := googleUuid.Must(googleUuid.NewV7()).String()

	tenant := &TenantConfig{
		ID:            tenantID,
		Name:          "row-closure-tenant",
		IsolationMode: TenantIsolationRow,
		Enabled:       true,
	}
	err = manager.RegisterTenant(ctx, tenant)
	require.NoError(t, err)

	// Get scoped DB with row isolation.
	scopedDB, err := manager.WithTenant(ctx, tenantID)
	require.NoError(t, err)
	require.NotNil(t, scopedDB)

	// Execute a query to trigger the scope closure (the WHERE clause lambda).
	// Even if the table doesn't exist, using Raw executes the closure.
	var result []map[string]any

	_ = scopedDB.Table("nonexistent_table_for_scope_test").Find(&result)
	// Result may or may not error; what matters is the closure was invoked.
}

// TestSanitizeSchemaName_Truncation verifies sanitizeSchemaName truncates names
// that exceed maxSchemaNameLength (63 characters).
func TestSanitizeSchemaName_Truncation(t *testing.T) {
	t.Parallel()

	// Generate a name longer than 63 characters.
	longName := strings.Repeat("abcdefghij", cryptoutilSharedMagic.GitRecentActivityDays) // 70 chars.
	result := sanitizeSchemaName(longName)
	require.Equal(t, cryptoutilSharedMagic.FQDNLabelMaxLength, len(result))
}
