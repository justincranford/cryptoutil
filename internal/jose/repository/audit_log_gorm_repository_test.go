// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"cryptoutil/internal/jose/domain"
	"cryptoutil/internal/jose/repository"

	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func setupAuditLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Open SQL database first with modernc driver.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Configure SQLite for concurrent operations.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	// Wrap with GORM using Dialector pattern (uses already-opened connection).
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate required tables.
	err = db.AutoMigrate(
		&cryptoutilTemplateRepository.TenantRealm{},
		&domain.AuditLogEntry{},
	)
	require.NoError(t, err)

	return db
}

func createTestAuditLogEntry(tenantID, realmID googleUuid.UUID, operation string) *domain.AuditLogEntry {
	return &domain.AuditLogEntry{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		UserID:       nil,
		Operation:    operation,
		ResourceType: "elastic_jwk",
		ResourceID:   googleUuid.New().String(),
		Success:      true,
		ErrorMessage: nil,
		Metadata:     nil,
	}
}

func TestAuditLogGormRepository_Create(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")

	err := repo.Create(ctx, entry)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, entry.ID)

	// Verify created.
	result, err := repo.GetByID(ctx, entry.ID)
	require.NoError(t, err)
	require.Equal(t, tenantID, result.TenantID)
	require.Equal(t, realmID, result.RealmID)
	require.Equal(t, "encrypt", result.Operation)
}

func TestAuditLogGormRepository_Create_AutoGenerateID(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	entry := &domain.AuditLogEntry{
		ID:           googleUuid.Nil, // Should auto-generate.
		TenantID:     tenantID,
		RealmID:      realmID,
		Operation:    "decrypt",
		ResourceType: "material_jwk",
		ResourceID:   googleUuid.New().String(),
		Success:      true,
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)
	require.NotEqual(t, googleUuid.Nil, entry.ID)
}

func TestAuditLogGormRepository_GetByID(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	entry := createTestAuditLogEntry(tenantID, realmID, "sign")
	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, entry.ID)
	require.NoError(t, err)
	require.Equal(t, entry.ID, result.ID)
	require.Equal(t, "sign", result.Operation)
}

func TestAuditLogGormRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	result, err := repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "audit log entry not found")
}

func TestAuditLogGormRepository_ListByTenantRealm(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create multiple entries.
	for i := 0; i < 5; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
		// Small sleep to ensure different timestamps.
		time.Sleep(time.Millisecond)
	}

	// List all.
	results, err := repo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 5)

	// List with pagination.
	results, err = repo.ListByTenantRealm(ctx, tenantID, realmID, 0, 3)
	require.NoError(t, err)
	require.Len(t, results, 3)

	results, err = repo.ListByTenantRealm(ctx, tenantID, realmID, 3, 3)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestAuditLogGormRepository_ListByTenantRealm_Empty(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	results, err := repo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestAuditLogGormRepository_ListByOperation(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create entries with different operations.
	operations := []string{"encrypt", "encrypt", "decrypt", "sign", "verify"}
	for _, op := range operations {
		entry := createTestAuditLogEntry(tenantID, realmID, op)
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// List by operation.
	results, err := repo.ListByOperation(ctx, tenantID, "encrypt", 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 2)

	results, err = repo.ListByOperation(ctx, tenantID, "decrypt", 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 1)
}

func TestAuditLogGormRepository_ListByResource(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	resourceID := googleUuid.New().String()

	// Create entries for same resource.
	for i := 0; i < 3; i++ {
		entry := &domain.AuditLogEntry{
			ID:           googleUuid.New(),
			TenantID:     tenantID,
			RealmID:      realmID,
			Operation:    "encrypt",
			ResourceType: "elastic_jwk",
			ResourceID:   resourceID,
			Success:      true,
		}
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Create entry for different resource.
	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	// List by resource.
	results, err := repo.ListByResource(ctx, "elastic_jwk", resourceID, 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 3)
}

func TestAuditLogGormRepository_ListByTimeRange(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Record start time.
	startTime := time.Now()

	// Create entries.
	for i := 0; i < 5; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	// Record end time.
	endTime := time.Now()

	// List by time range.
	results, err := repo.ListByTimeRange(ctx, tenantID, startTime.Add(-time.Second), endTime.Add(time.Second), 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 5)

	// List with time range that excludes all entries.
	futureStart := endTime.Add(time.Hour)
	futureEnd := futureStart.Add(time.Hour)
	results, err = repo.ListByTimeRange(ctx, tenantID, futureStart, futureEnd, 0, 10)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestAuditLogGormRepository_Count(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create entries.
	for i := 0; i < 7; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(7), count)
}

func TestAuditLogGormRepository_Count_Isolation(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenant1 := googleUuid.New()
	tenant2 := googleUuid.New()
	realmID := googleUuid.New()

	// Create entries for tenant1.
	for i := 0; i < 3; i++ {
		entry := createTestAuditLogEntry(tenant1, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Create entries for tenant2.
	for i := 0; i < 5; i++ {
		entry := createTestAuditLogEntry(tenant2, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Verify counts are isolated.
	count1, err := repo.Count(ctx, tenant1)
	require.NoError(t, err)
	require.Equal(t, int64(3), count1)

	count2, err := repo.Count(ctx, tenant2)
	require.NoError(t, err)
	require.Equal(t, int64(5), count2)
}

func TestAuditLogGormRepository_DeleteOlderThan(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create entries.
	for i := 0; i < 5; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	// Record time after creating entries.
	cutoffTime := time.Now().Add(time.Second)

	// Delete entries older than cutoff (should delete all).
	deleted, err := repo.DeleteOlderThan(ctx, tenantID, cutoffTime)
	require.NoError(t, err)
	require.Equal(t, int64(5), deleted)

	// Verify all deleted.
	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func TestAuditLogGormRepository_DeleteOlderThan_Partial(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create old entries.
	for i := 0; i < 3; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Record cutoff time.
	time.Sleep(10 * time.Millisecond)

	cutoffTime := time.Now()

	// Create new entries after cutoff.
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 2; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "decrypt")
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	// Delete entries older than cutoff.
	deleted, err := repo.DeleteOlderThan(ctx, tenantID, cutoffTime)
	require.NoError(t, err)
	require.Equal(t, int64(3), deleted)

	// Verify remaining entries.
	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
}

func TestAuditLogGormRepository_CreateWithSampling_Always(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// With 1.0 sampling rate, should always create.
	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
	created, err := repo.CreateWithSampling(ctx, entry, 1.0)
	require.NoError(t, err)
	require.True(t, created)

	// Verify created.
	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestAuditLogGormRepository_CreateWithSampling_Never(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// With 0.0 sampling rate, should never create.
	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
	created, err := repo.CreateWithSampling(ctx, entry, 0.0)
	require.NoError(t, err)
	require.False(t, created)

	// Verify not created.
	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)
}

func TestAuditLogGormRepository_CreateWithSampling_Probabilistic(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// With 0.5 sampling rate, should create roughly half.
	createdCount := 0

	for i := 0; i < 100; i++ {
		entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
		created, err := repo.CreateWithSampling(ctx, entry, 0.5)
		require.NoError(t, err)

		if created {
			createdCount++
		}
	}

	// Should be roughly 50, allow variance (25-75).
	require.Greater(t, createdCount, 20)
	require.Less(t, createdCount, 80)

	// Verify count matches.
	count, err := repo.Count(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, int64(createdCount), count)
}

func TestAuditLogGormRepository_WithOptionalFields(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	userID := googleUuid.New()
	errorMsg := "operation failed"
	metadata := `{"key": "value"}`

	entry := &domain.AuditLogEntry{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		UserID:       &userID,
		Operation:    "encrypt",
		ResourceType: "elastic_jwk",
		ResourceID:   googleUuid.New().String(),
		Success:      false,
		ErrorMessage: &errorMsg,
		Metadata:     &metadata,
	}

	err := repo.Create(ctx, entry)
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, entry.ID)
	require.NoError(t, err)
	require.NotNil(t, result.UserID)
	require.Equal(t, userID, *result.UserID)
	require.NotNil(t, result.ErrorMessage)
	require.Equal(t, errorMsg, *result.ErrorMessage)
	require.NotNil(t, result.Metadata)
	require.Equal(t, metadata, *result.Metadata)
}

func TestAuditLogGormRepository_WithTransaction(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create entry within transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilTemplateRepository.WithTransaction(ctx, tx)

	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
	err := repo.Create(txCtx, entry)
	require.NoError(t, err)

	// Commit transaction.
	err = tx.Commit().Error
	require.NoError(t, err)

	// Verify outside transaction.
	result, err := repo.GetByID(ctx, entry.ID)
	require.NoError(t, err)
	require.Equal(t, entry.ID, result.ID)
}

func TestAuditLogGormRepository_WithTransaction_Rollback(t *testing.T) {
	t.Parallel()
	db := setupAuditLogTestDB(t)
	repo := repository.NewAuditLogGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	// Create entry within transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilTemplateRepository.WithTransaction(ctx, tx)

	entry := createTestAuditLogEntry(tenantID, realmID, "encrypt")
	err := repo.Create(txCtx, entry)
	require.NoError(t, err)

	// Rollback transaction.
	err = tx.Rollback().Error
	require.NoError(t, err)

	// Verify entry was not persisted.
	result, err := repo.GetByID(ctx, entry.ID)
	require.Error(t, err)
	require.Nil(t, result)
}
