// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"database/sql"
	"io/fs"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

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

	_, _, err = repo.List(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

	_, _, err = repo.ListByElasticJWK(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

	_, _, err = repo.ListByOperation(ctx, googleUuid.New(), cryptoutilAppsJoseJaDomain.OperationSign, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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

	_, err = repo.DeleteOlderThan(ctx, googleUuid.New(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
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

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	// Close immediately without applying migrations.
	err = sqlDB.Close()
	require.NoError(t, err)

	// Now try to apply migrations to the closed DB.
	err = ApplyJoseJAMigrations(sqlDB, DatabaseTypeSQLite)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to apply jose-ja migrations"))
}
