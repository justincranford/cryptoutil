// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// createClosedDatabase creates a new in-memory SQLite database, applies migrations,
// closes the connection, and returns a GORM DB with the closed connection.
// This enables testing database error paths.
func createClosedDatabase() (*gorm.DB, error) {
	ctx := context.Background()
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite.
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")

	// Create GORM DB.
	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()

		return nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	// Apply migrations.
	if err := ApplyJoseJAMigrations(sqlDB, DatabaseTypeSQLite); err != nil {
		_ = sqlDB.Close()

		return nil, err
	}

	// Close the underlying connection to force errors.
	if err := sqlDB.Close(); err != nil {
		return nil, fmt.Errorf("failed to close database: %w", err)
	}

	return gormDB, nil
}

// ====================
// ElasticJWK Repository Database Error Tests
// ====================

func TestElasticJWKRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-create-error",
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
	}

	err = repo.Create(ctx, jwk)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKRepository_GetDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, err = repo.Get(ctx, googleUuid.New(), "test-kid")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKRepository_GetByIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, err = repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK by ID"))
}

func TestElasticJWKRepository_ListDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, _, err = repo.List(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count elastic JWKs") ||
			strings.Contains(err.Error(), "failed to list elastic JWKs"),
		"Expected count or list error, got: %v", err)
}

func TestElasticJWKRepository_UpdateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-update-error",
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
	}

	err = repo.Update(ctx, jwk)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update elastic JWK"))
}

func TestElasticJWKRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err = repo.Delete(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete elastic JWK"))
}

func TestElasticJWKRepository_IncrementMaterialCountDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err = repo.IncrementMaterialCount(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to increment material count"))
}

func TestElasticJWKRepository_DecrementMaterialCountDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err = repo.DecrementMaterialCount(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to decrement material count"))
}

// ====================
// MaterialJWK Repository Database Error Tests
// ====================

func TestMaterialJWKRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	material := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:           *id,
		ElasticJWKID: *elasticJWKID,
		MaterialKID:  "test-material-error",
		Active:       true,
	}

	err = repo.Create(ctx, material)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create material JWK"))
}

func TestMaterialJWKRepository_GetByMaterialKIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err = repo.GetByMaterialKID(ctx, "test-kid")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by KID"))
}

func TestMaterialJWKRepository_GetByIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err = repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by ID"))
}

func TestMaterialJWKRepository_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err = repo.GetActiveMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get active material JWK"))
}

func TestMaterialJWKRepository_ListByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, _, err = repo.ListByElasticJWK(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count material JWKs") ||
			strings.Contains(err.Error(), "failed to list material JWKs"),
		"Expected count or list error, got: %v", err)
}

func TestMaterialJWKRepository_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	newMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:           *id,
		ElasticJWKID: *elasticJWKID,
		MaterialKID:  "new-material",
		Active:       true,
	}

	err = repo.RotateMaterial(ctx, *elasticJWKID, newMaterial)
	require.Error(t, err)
	// Transaction or any step could fail.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "sql: database is closed"),
		"Expected database error, got: %v", err)
}

// TestMaterialJWKRepository_RotateMaterialCreateError tests the "failed to create new material" error path
// inside RotateMaterial by using a duplicate MaterialKID to cause a constraint violation.
func TestMaterialJWKRepository_RotateMaterialCreateError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := NewMaterialJWKRepository(testDB)
	elasticRepo := NewElasticJWKRepository(testDB)

	// Create unique test data.
	elasticJWKID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	materialKID := googleUuid.NewString() // Use UUID for uniqueness.

	// First create an ElasticJWK to satisfy foreign key constraint.
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *elasticJWKID,
		TenantID:     *tenantID,
		KID:          googleUuid.NewString(),
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    "RS256",
		Use:          "sig",
		MaxMaterials: 10,
	}
	err := elasticRepo.Create(ctx, elasticJWK)
	require.NoError(t, err)

	// Create first material with a specific MaterialKID.
	firstMaterialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	firstMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:            *firstMaterialID,
		ElasticJWKID:  *elasticJWKID,
		MaterialKID:   materialKID, // This KID will be duplicated.
		PrivateJWKJWE: "encrypted-private-1",
		PublicJWKJWE:  "encrypted-public-1",
		Active:        false,
	}
	err = repo.Create(ctx, firstMaterial)
	require.NoError(t, err)

	// Now try to rotate with a NEW material that uses the SAME MaterialKID (duplicate).
	secondMaterialID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	duplicateMaterial := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:            *secondMaterialID,
		ElasticJWKID:  *elasticJWKID,
		MaterialKID:   materialKID, // DUPLICATE - should cause UNIQUE constraint violation.
		PrivateJWKJWE: "encrypted-private-2",
		PublicJWKJWE:  "encrypted-public-2",
		Active:        true,
	}

	// This should fail on the "Create" inside the transaction due to duplicate MaterialKID.
	err = repo.RotateMaterial(ctx, *elasticJWKID, duplicateMaterial)
	require.Error(t, err)
	// Should hit the "failed to create new material" error path.
	require.True(t,
		strings.Contains(err.Error(), "failed to create new material") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed"),
		"Expected create material error, got: %v", err)
}

func TestMaterialJWKRepository_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err = repo.RetireMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to retire material JWK"))
}

func TestMaterialJWKRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	err = repo.Delete(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete material JWK"))
}

func TestMaterialJWKRepository_CountMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err = repo.CountMaterials(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to count material JWKs"))
}

// ====================
// AuditConfig Repository Database Error Tests
// ====================

func TestAuditConfigRepository_GetDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.Get(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditConfigRepository_GetAllForTenantDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.GetAllForTenant(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit configs for tenant"))
}

func TestAuditConfigRepository_UpsertDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     *tenantID,
		Operation:    cryptoutilAppsJoseJaDomain.OperationSign,
		Enabled:      true,
		SamplingRate: 0.5,
	}

	err = repo.Upsert(ctx, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to upsert audit config"))
}

func TestAuditConfigRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	err = repo.Delete(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete audit config"))
}

func TestAuditConfigRepository_ShouldAuditDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditConfigRepository(closedDB)

	_, err = repo.ShouldAudit(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to") ||
		strings.Contains(err.Error(), "sql: database is closed"))
}

// ====================
// AuditLog Repository Database Error Tests
// ====================

func TestAuditLogRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	entry := &cryptoutilAppsJoseJaDomain.AuditLogEntry{
		ID:        *id,
		TenantID:  *tenantID,
		Operation: cryptoutilAppsJoseJaDomain.OperationSign,
		Success:   true,
		RequestID: googleUuid.NewString(),
	}

	err = repo.Create(ctx, entry)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create audit log entry"))
}

func TestAuditLogRepository_ListDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err = repo.List(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}

func TestAuditLogRepository_ListByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err = repo.ListByElasticJWK(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}

func TestAuditLogRepository_ListByOperationDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, _, err = repo.ListByOperation(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign, 0, 10)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count audit log entries") ||
			strings.Contains(err.Error(), "failed to list audit log entries"),
		"Expected count or list error, got: %v", err)
}

func TestAuditLogRepository_GetByRequestIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, err = repo.GetByRequestID(ctx, googleUuid.NewString())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit log entry by request ID"))
}

func TestAuditLogRepository_DeleteOlderThanDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB, err := createClosedDatabase()
	require.NoError(t, err)

	ctx := context.Background()
	repo := NewAuditLogRepository(closedDB)

	_, err = repo.DeleteOlderThan(ctx, googleUuid.New(), 30)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit log entries"))
}

// ====================
// MergedFS Error Tests
// ====================

func TestMergedFS_ReadDirEmpty(t *testing.T) {
	t.Parallel()

	// Get the merged FS.
	mergedFS := GetMergedMigrationsFS()

	// Try to read a directory that doesn't exist.
	readDirFS, ok := mergedFS.(interface {
		ReadDir(name string) ([]fs.DirEntry, error)
	})
	require.True(t, ok, "mergedFS does not implement ReadDir")

	_, err := readDirFS.ReadDir("nonexistent_directory")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "directory not found"))
}

func TestMergedFS_OpenError(t *testing.T) {
	t.Parallel()

	// Get the merged FS.
	mergedFS := GetMergedMigrationsFS()

	// Try to open a file that doesn't exist.
	_, err := mergedFS.Open("nonexistent_file.sql")

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to open from template"))
}

func TestMergedFS_ReadFileError(t *testing.T) {
	t.Parallel()

	// Get the merged FS.
	mergedFS := GetMergedMigrationsFS()

	// Try to read a file that doesn't exist.
	readFileFS, ok := mergedFS.(interface {
		ReadFile(name string) ([]byte, error)
	})
	require.True(t, ok, "mergedFS does not implement ReadFile")

	_, err := readFileFS.ReadFile("nonexistent_file.sql")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to read from template"))
}

func TestMergedFS_StatError(t *testing.T) {
	t.Parallel()

	// Get the merged FS.
	mergedFS := GetMergedMigrationsFS()

	// Try to stat a file that doesn't exist.
	statFS, ok := mergedFS.(interface {
		Stat(name string) (fs.FileInfo, error)
	})
	require.True(t, ok, "mergedFS does not implement Stat")

	_, err := statFS.Stat("nonexistent_file.sql")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to stat from template"))
}

func TestApplyJoseJAMigrations_Error(t *testing.T) {
	t.Parallel()

	// Create a closed database to force migration error.
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Close immediately without applying migrations.
	err = sqlDB.Close()
	require.NoError(t, err)

	// Now try to apply migrations to the closed DB.
	err = ApplyJoseJAMigrations(sqlDB, DatabaseTypeSQLite)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to apply jose-ja migrations"))
}
