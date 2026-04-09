// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	gormSQLite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceServerDomain "cryptoutil/internal/apps/framework/service/server/domain"
)

// TestRepository_ClosedDB_ReturnsError verifies that every repository method
// returns an error when the underlying DB connection is closed.
func TestRepository_ClosedDB_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupDB func(t *testing.T) *gorm.DB
		callFn  func(ctx context.Context, db *gorm.DB) error
	}{
		{
			name:    "RoleRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewRoleRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "UserRoleRepository ListRolesByUser",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUserRoleRepository(db).ListRolesByUser(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "UserRoleRepository ListUsersByRole",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUserRoleRepository(db).ListUsersByRole(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "ClientRoleRepository ListRolesByClient",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewClientRoleRepository(db).ListRolesByClient(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "ClientRoleRepository ListClientsByRole",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewClientRoleRepository(db).ListClientsByRole(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "TenantRealmRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantRealmRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)

				return err
			},
		},
		{
			name:    "TenantJoinRequestRepository Update",
			setupDB: setupJoinRequestTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				return NewTenantJoinRequestRepository(db).Update(ctx, &cryptoutilAppsFrameworkServiceServerDomain.TenantJoinRequest{
					ID: googleUuid.Must(googleUuid.NewV7()),
				})
			},
		},
		{
			name:    "TenantJoinRequestRepository GetByID",
			setupDB: setupJoinRequestTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantJoinRequestRepository(db).GetByID(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "TenantJoinRequestRepository ListByTenant",
			setupDB: setupJoinRequestTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantJoinRequestRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "TenantJoinRequestRepository ListByStatus",
			setupDB: setupJoinRequestTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantJoinRequestRepository(db).ListByStatus(ctx, "pending")

				return err
			},
		},
		{
			name:    "TenantJoinRequestRepository ListByTenantAndStatus",
			setupDB: setupJoinRequestTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantJoinRequestRepository(db).ListByTenantAndStatus(ctx, googleUuid.Must(googleUuid.NewV7()), "pending")

				return err
			},
		},
		{
			name:    "TenantRepository List",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewTenantRepository(db).List(ctx, true)

				return err
			},
		},
		{
			name:    "TenantRepository Update",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				return NewTenantRepository(db).Update(ctx, &Tenant{ID: googleUuid.Must(googleUuid.NewV7())})
			},
		},
		{
			name:    "TenantRepository CountUsersAndClients",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, _, err := NewTenantRepository(db).CountUsersAndClients(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "TenantRepository Delete via CountError",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				return NewTenantRepository(db).Delete(ctx, googleUuid.Must(googleUuid.NewV7()))
			},
		},
		{
			name:    "UserRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUserRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)

				return err
			},
		},
		{
			name:    "ClientRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewClientRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)

				return err
			},
		},
		{
			name:    "UnverifiedUserRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUnverifiedUserRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "UnverifiedUserRepository DeleteExpired",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUnverifiedUserRepository(db).DeleteExpired(ctx)

				return err
			},
		},
		{
			name:    "UnverifiedClientRepository ListByTenant",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUnverifiedClientRepository(db).ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))

				return err
			},
		},
		{
			name:    "UnverifiedClientRepository DeleteExpired",
			setupDB: setupTestDB,
			callFn: func(ctx context.Context, db *gorm.DB) error {
				_, err := NewUnverifiedClientRepository(db).DeleteExpired(ctx)

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := tc.setupDB(t)
			ctx := context.Background()

			sqlDB, err := db.DB()
			require.NoError(t, err)
			require.NoError(t, sqlDB.Close())

			err = tc.callFn(ctx, db)
			require.Error(t, err)
		})
	}
}

// setupPartialTestDB creates an isolated in-memory SQLite DB with only the
// specified models migrated. Used to trigger errors in specific code paths.
func setupPartialTestDB(t *testing.T, models ...any) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:partial_%s?mode=memory&cache=shared", googleUuid.Must(googleUuid.NewV7()).String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	dialector := gormSQLite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	sqlDB2, err := db.DB()
	require.NoError(t, err)

	sqlDB2.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB2.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB2.SetConnMaxLifetime(0)

	err = db.AutoMigrate(models...)
	require.NoError(t, err)

	return db
}

// TestTenantRepository_CountUsersAndClients_ClientCountError verifies that when
// the Client table is missing, CountUsersAndClients returns an error on client
// count after the user count succeeds.
func TestTenantRepository_CountUsersAndClients_ClientCountError(t *testing.T) {
	t.Parallel()

	// Migrate only Tenant and User so client count fails (no clients table).
	db := setupPartialTestDB(t, &Tenant{}, &User{})
	repo := NewTenantRepository(db)
	ctx := context.Background()

	_, _, err := repo.CountUsersAndClients(ctx, googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
}

// TestTenantRepository_Delete_DBError verifies Delete returns an error when the
// underlying db.Delete call fails (tenants table dropped after migration).
func TestTenantRepository_Delete_DBError(t *testing.T) {
	t.Parallel()

	// Migrate all tables, then drop the tenants table so the delete operation
	// itself fails while CountUsersAndClients still succeeds (0 users, 0 clients).
	db := setupPartialTestDB(t, &User{}, &Client{})
	ctx := context.Background()

	// SQLite does not enforce FK constraints by default, so dropping tenants
	// table is safe here and makes db.Delete(&Tenant{}) fail.
	result := db.Exec("DROP TABLE IF EXISTS tenants")
	require.NoError(t, result.Error)

	repo := NewTenantRepository(db)

	err := repo.Delete(ctx, googleUuid.Must(googleUuid.NewV7()))
	require.Error(t, err)
}
