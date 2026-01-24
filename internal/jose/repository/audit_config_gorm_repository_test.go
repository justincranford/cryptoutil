// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

// Test operation constants.
const (
	testOpEncrypt = "encrypt"
	testOpKeygen  = "keygen"
	testOpRotate  = "rotate"
)

func setupAuditConfigTestDB(t *testing.T) *gorm.DB {
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
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
		&cryptoutilJoseDomain.AuditConfig{},
	)
	require.NoError(t, err)

	return db
}

func TestAuditConfigGormRepository_Get(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := testOpEncrypt

	// Create config.
	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.5,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Get config.
	result, err := repo.Get(ctx, tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, tenantID, result.TenantID)
	require.Equal(t, operation, result.Operation)
	require.True(t, result.Enabled)
	require.Equal(t, 0.5, result.SamplingRate)
}

func TestAuditConfigGormRepository_Get_NotFound(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()

	result, err := repo.Get(ctx, tenantID, "nonexistent")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "audit config not found")
}

func TestAuditConfigGormRepository_GetAll(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()

	// Create multiple configs.
	operations := []string{"encrypt", "decrypt", "sign", "verify"}
	for _, op := range operations {
		config := &cryptoutilJoseDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    op,
			Enabled:      true,
			SamplingRate: 0.1,
		}
		err := repo.Upsert(ctx, config)
		require.NoError(t, err)
	}

	// Get all configs.
	results, err := repo.GetAll(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, results, 4)
}

func TestAuditConfigGormRepository_GetAll_Empty(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()

	results, err := repo.GetAll(ctx, tenantID)
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestAuditConfigGormRepository_Upsert_Create(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := testOpKeygen

	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      false,
		SamplingRate: 0.0,
	}

	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Verify created.
	result, err := repo.Get(ctx, tenantID, operation)
	require.NoError(t, err)
	require.False(t, result.Enabled)
	require.Equal(t, 0.0, result.SamplingRate)
}

func TestAuditConfigGormRepository_Upsert_Update(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := testOpRotate

	// Create config.
	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.1,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Update config.
	config.Enabled = false
	config.SamplingRate = 0.5
	err = repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Verify updated.
	result, err := repo.Get(ctx, tenantID, operation)
	require.NoError(t, err)
	require.False(t, result.Enabled)
	require.Equal(t, 0.5, result.SamplingRate)
}

func TestAuditConfigGormRepository_Delete(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := "decrypt"

	// Create config.
	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.2,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	// Delete config.
	err = repo.Delete(ctx, tenantID, operation)
	require.NoError(t, err)

	// Verify deleted.
	result, err := repo.Get(ctx, tenantID, operation)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestAuditConfigGormRepository_Delete_NotFound(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()

	err := repo.Delete(ctx, tenantID, "nonexistent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "audit config not found")
}

func TestAuditConfigGormRepository_IsEnabled_Enabled(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := "sign"

	// Create enabled config.
	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.75,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	enabled, samplingRate, err := repo.IsEnabled(ctx, tenantID, operation)
	require.NoError(t, err)
	require.True(t, enabled)
	require.Equal(t, 0.75, samplingRate)
}

func TestAuditConfigGormRepository_IsEnabled_Disabled(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := "verify"

	// Create disabled config.
	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      false,
		SamplingRate: 0.0,
	}
	err := repo.Upsert(ctx, config)
	require.NoError(t, err)

	enabled, samplingRate, err := repo.IsEnabled(ctx, tenantID, operation)
	require.NoError(t, err)
	require.False(t, enabled)
	require.Equal(t, 0.0, samplingRate)
}

func TestAuditConfigGormRepository_IsEnabled_NoConfig(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()

	// When no config exists, should return disabled (conservative default).
	enabled, samplingRate, err := repo.IsEnabled(ctx, tenantID, "nonexistent")
	require.NoError(t, err)
	require.False(t, enabled)
	require.Equal(t, 0.0, samplingRate)
}

func TestAuditConfigGormRepository_TenantIsolation(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenant1 := googleUuid.New()
	tenant2 := googleUuid.New()
	operation := testOpEncrypt

	// Create config for tenant1.
	config1 := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenant1,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.5,
	}
	err := repo.Upsert(ctx, config1)
	require.NoError(t, err)

	// Create config for tenant2 with different settings.
	config2 := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenant2,
		Operation:    operation,
		Enabled:      false,
		SamplingRate: 0.1,
	}
	err = repo.Upsert(ctx, config2)
	require.NoError(t, err)

	// Verify tenant1 config.
	result1, err := repo.Get(ctx, tenant1, operation)
	require.NoError(t, err)
	require.True(t, result1.Enabled)
	require.Equal(t, 0.5, result1.SamplingRate)

	// Verify tenant2 config.
	result2, err := repo.Get(ctx, tenant2, operation)
	require.NoError(t, err)
	require.False(t, result2.Enabled)
	require.Equal(t, 0.1, result2.SamplingRate)

	// Verify GetAll isolation.
	results1, err := repo.GetAll(ctx, tenant1)
	require.NoError(t, err)
	require.Len(t, results1, 1)

	results2, err := repo.GetAll(ctx, tenant2)
	require.NoError(t, err)
	require.Len(t, results2, 1)
}

func TestAuditConfigGormRepository_WithTransaction(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := testOpKeygen

	// Create config within transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.25,
	}
	err := repo.Upsert(txCtx, config)
	require.NoError(t, err)

	// Commit transaction.
	err = tx.Commit().Error
	require.NoError(t, err)

	// Verify outside transaction.
	result, err := repo.Get(ctx, tenantID, operation)
	require.NoError(t, err)
	require.Equal(t, tenantID, result.TenantID)
}

func TestAuditConfigGormRepository_WithTransaction_Rollback(t *testing.T) {
	t.Parallel()
	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	ctx := context.Background()

	tenantID := googleUuid.New()
	operation := testOpRotate

	// Create config within transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	config := &cryptoutilJoseDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    operation,
		Enabled:      true,
		SamplingRate: 0.5,
	}
	err := repo.Upsert(txCtx, config)
	require.NoError(t, err)

	// Rollback transaction.
	err = tx.Rollback().Error
	require.NoError(t, err)

	// Verify config was not persisted.
	result, err := repo.Get(ctx, tenantID, operation)
	require.Error(t, err)
	require.Nil(t, result)
}
