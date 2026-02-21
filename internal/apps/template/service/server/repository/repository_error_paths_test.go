// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	gormSQLite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
)

// TestRoleRepository_ListByTenant_DBError verifies ListByTenant returns an error
// when the underlying DB connection is closed.
func TestRoleRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewRoleRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestUserRoleRepository_ListRolesByUser_DBError verifies ListRolesByUser returns
// an error when the DB connection is closed.
func TestUserRoleRepository_ListRolesByUser_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUserRoleRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListRolesByUser(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestUserRoleRepository_ListUsersByRole_DBError verifies ListUsersByRole returns
// an error when the DB connection is closed.
func TestUserRoleRepository_ListUsersByRole_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUserRoleRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListUsersByRole(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestClientRoleRepository_ListRolesByClient_DBError verifies ListRolesByClient
// returns an error when the DB connection is closed.
func TestClientRoleRepository_ListRolesByClient_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewClientRoleRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListRolesByClient(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestClientRoleRepository_ListClientsByRole_DBError verifies ListClientsByRole
// returns an error when the DB connection is closed.
func TestClientRoleRepository_ListClientsByRole_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewClientRoleRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListClientsByRole(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestTenantRealmRepository_ListByTenant_DBError verifies ListByTenant returns
// an error when the DB connection is closed.
func TestTenantRealmRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewTenantRealmRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)
require.Error(t, err)
}

// TestTenantJoinRequestRepository_Update_DBError verifies Update returns an
// error when the DB connection is closed.
func TestTenantJoinRequestRepository_Update_DBError(t *testing.T) {
t.Parallel()

db := setupJoinRequestTestDB(t)
repo := NewTenantJoinRequestRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

request := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
ID: googleUuid.Must(googleUuid.NewV7()),
}

err = repo.Update(ctx, request)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to update join request")
}

// TestTenantJoinRequestRepository_GetByID_DBError verifies GetByID returns a
// non-record-not-found error when the DB connection is closed.
func TestTenantJoinRequestRepository_GetByID_DBError(t *testing.T) {
t.Parallel()

db := setupJoinRequestTestDB(t)
repo := NewTenantJoinRequestRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.GetByID(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
require.Contains(t, err.Error(), "failed to get join request")
}

// TestTenantJoinRequestRepository_ListByTenant_DBError verifies ListByTenant
// returns an error when the DB connection is closed.
func TestTenantJoinRequestRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupJoinRequestTestDB(t)
repo := NewTenantJoinRequestRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestTenantJoinRequestRepository_ListByStatus_DBError verifies ListByStatus
// returns an error when the DB connection is closed.
func TestTenantJoinRequestRepository_ListByStatus_DBError(t *testing.T) {
t.Parallel()

db := setupJoinRequestTestDB(t)
repo := NewTenantJoinRequestRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByStatus(ctx, "pending")
require.Error(t, err)
}

// TestTenantJoinRequestRepository_ListByTenantAndStatus_DBError verifies
// ListByTenantAndStatus returns an error when the DB connection is closed.
func TestTenantJoinRequestRepository_ListByTenantAndStatus_DBError(t *testing.T) {
t.Parallel()

db := setupJoinRequestTestDB(t)
repo := NewTenantJoinRequestRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenantAndStatus(ctx, googleUuid.Must(googleUuid.NewV7()), "pending")
require.Error(t, err)
}

// TestTenantRepository_List_DBError verifies List returns an error when the DB
// connection is closed.
func TestTenantRepository_List_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewTenantRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.List(ctx, true)
require.Error(t, err)
}

// TestTenantRepository_Update_DBError verifies Update returns an error when the
// DB connection is closed.
func TestTenantRepository_Update_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewTenantRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

tenant := &Tenant{
ID: googleUuid.Must(googleUuid.NewV7()),
}

err = repo.Update(ctx, tenant)
require.Error(t, err)
}

// TestTenantRepository_CountUsersAndClients_DBError verifies CountUsersAndClients
// returns an error when the DB connection is closed (user count fails first).
func TestTenantRepository_CountUsersAndClients_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewTenantRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, _, err = repo.CountUsersAndClients(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestTenantRepository_Delete_CountError verifies Delete returns an error when
// CountUsersAndClients fails due to the DB connection being closed.
func TestTenantRepository_Delete_CountError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewTenantRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

err = repo.Delete(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestUserRepository_ListByTenant_DBError verifies ListByTenant returns an error
// when the DB connection is closed.
func TestUserRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUserRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)
require.Error(t, err)
}

// TestClientRepository_ListByTenant_DBError verifies ListByTenant returns an
// error when the DB connection is closed.
func TestClientRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewClientRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()), true)
require.Error(t, err)
}

// TestUnverifiedUserRepository_ListByTenant_DBError verifies ListByTenant
// returns an error when the DB connection is closed.
func TestUnverifiedUserRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUnverifiedUserRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestUnverifiedUserRepository_DeleteExpired_DBError verifies DeleteExpired
// returns an error when the DB connection is closed.
func TestUnverifiedUserRepository_DeleteExpired_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUnverifiedUserRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.DeleteExpired(ctx)
require.Error(t, err)
}

// TestUnverifiedClientRepository_ListByTenant_DBError verifies ListByTenant
// returns an error when the DB connection is closed.
func TestUnverifiedClientRepository_ListByTenant_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUnverifiedClientRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.ListByTenant(ctx, googleUuid.Must(googleUuid.NewV7()))
require.Error(t, err)
}

// TestUnverifiedClientRepository_DeleteExpired_DBError verifies DeleteExpired
// returns an error when the DB connection is closed.
func TestUnverifiedClientRepository_DeleteExpired_DBError(t *testing.T) {
t.Parallel()

db := setupTestDB(t)
repo := NewUnverifiedClientRepository(db)
ctx := context.Background()

sqlDB, err := db.DB()
require.NoError(t, err)
require.NoError(t, sqlDB.Close())

_, err = repo.DeleteExpired(ctx)
require.Error(t, err)
}

// setupPartialTestDB creates an isolated in-memory SQLite DB with only the
// specified models migrated. Used to trigger errors in specific code paths.
func setupPartialTestDB(t *testing.T, models ...interface{}) *gorm.DB {
t.Helper()

dsn := fmt.Sprintf("file:partial_%s?mode=memory&cache=shared", googleUuid.Must(googleUuid.NewV7()).String())

sqlDB, err := sql.Open("sqlite", dsn)
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

sqlDB2.SetMaxOpenConns(5)
sqlDB2.SetMaxIdleConns(5)
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
