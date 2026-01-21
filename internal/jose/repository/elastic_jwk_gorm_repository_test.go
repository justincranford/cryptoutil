// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	cryptoutilTemplateRepository "cryptoutil/internal/apps/template/service/server/repository"
	"cryptoutil/internal/jose/domain"
	"cryptoutil/internal/jose/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

func setupTestDB(t *testing.T) *gorm.DB {
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

	// Use GORM AutoMigrate instead of golang-migrate to avoid CGO dependency.
	// AutoMigrate creates tables based on GORM model struct tags.
	err = db.AutoMigrate(
		// Template models needed for tenant/realm foreign key support.
		&cryptoutilTemplateRepository.Tenant{},
		&cryptoutilTemplateRepository.TenantRealm{},
		// JOSE domain models.
		&domain.ElasticJWK{},
		&domain.MaterialJWK{},
		&domain.AuditConfig{},
		&domain.AuditLogEntry{},
	)
	require.NoError(t, err)

	return db
}

func createTestTenantAndRealm(t *testing.T, db *gorm.DB) (tenantID, realmID googleUuid.UUID) {
	t.Helper()

	tenantID = googleUuid.New()
	realmID = googleUuid.New()

	// Create tenant using GORM model.
	tenant := &cryptoutilTemplateRepository.Tenant{
		ID:          tenantID,
		Name:        "Test Tenant " + tenantID.String(),
		Description: "Test tenant for repository tests",
		Active:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create tenant_realm using GORM model.
	tenantRealm := &cryptoutilTemplateRepository.TenantRealm{
		ID:        googleUuid.New(),
		TenantID:  tenantID,
		RealmID:   realmID,
		Type:      "test",
		Active:    true,
		Source:    "db",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(tenantRealm).Error)

	return tenantID, realmID
}

func TestElasticJWKGormRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	elasticJWK := &domain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  "test-kid-001",
		KTY:                  "RSA",
		ALG:                  "RS256",
		USE:                  "sig",
		MaxMaterials:         1000,
		CurrentMaterialCount: 0,
	}

	err := repo.Create(ctx, elasticJWK)
	require.NoError(t, err)

	// Verify in database
	var result domain.ElasticJWK

	err = db.Where("id = ?", elasticJWK.ID).First(&result).Error
	require.NoError(t, err)
	require.Equal(t, elasticJWK.TenantID, result.TenantID)
	require.Equal(t, elasticJWK.RealmID, result.RealmID)
	require.Equal(t, elasticJWK.KID, result.KID)
	require.Equal(t, elasticJWK.KTY, result.KTY)
	require.Equal(t, elasticJWK.ALG, result.ALG)
	require.Equal(t, elasticJWK.USE, result.USE)
}

func TestElasticJWKGormRepository_Get(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create test data
	elasticJWK := &domain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  "test-kid-002",
		KTY:                  "EC",
		ALG:                  "ES256",
		USE:                  "sig",
		MaxMaterials:         500,
		CurrentMaterialCount: 0,
	}
	require.NoError(t, repo.Create(ctx, elasticJWK))

	// Test Get
	result, err := repo.Get(ctx, tenantID, realmID, "test-kid-002")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, elasticJWK.ID, result.ID)
	require.Equal(t, elasticJWK.KID, result.KID)
	require.Equal(t, elasticJWK.KTY, result.KTY)
}

func TestElasticJWKGormRepository_Get_NotFound(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Test Get non-existent KID
	result, err := repo.Get(ctx, tenantID, realmID, "non-existent-kid")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "elastic JWK not found")
}

func TestElasticJWKGormRepository_Get_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	tenantID1, realmID1 := createTestTenantAndRealm(t, db)
	tenantID2, realmID2 := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create JWK for tenant1
	elasticJWK := &domain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID1,
		RealmID:              realmID1,
		KID:                  "tenant1-kid",
		KTY:                  "RSA",
		ALG:                  "RS256",
		USE:                  "sig",
		MaxMaterials:         1000,
		CurrentMaterialCount: 0,
	}
	require.NoError(t, repo.Create(ctx, elasticJWK))

	// Try to get with tenant2 credentials (should fail)
	result, err := repo.Get(ctx, tenantID2, realmID2, "tenant1-kid")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "elastic JWK not found")
}

func TestElasticJWKGormRepository_List(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create multiple JWKs
	for i := 0; i < 5; i++ {
		elasticJWK := &domain.ElasticJWK{
			ID:                   googleUuid.New(),
			TenantID:             tenantID,
			RealmID:              realmID,
			KID:                  fmt.Sprintf("kid-%03d", i),
			KTY:                  "RSA",
			ALG:                  "RS256",
			USE:                  "sig",
			MaxMaterials:         1000,
			CurrentMaterialCount: 0,
		}
		require.NoError(t, repo.Create(ctx, elasticJWK))
	}

	// Test List - all
	results, err := repo.List(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 5)

	// Test List - pagination
	results, err = repo.List(ctx, tenantID, realmID, 0, 2)
	require.NoError(t, err)
	require.Len(t, results, 2)

	results, err = repo.List(ctx, tenantID, realmID, 2, 2)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestElasticJWKGormRepository_List_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	tenantID1, realmID1 := createTestTenantAndRealm(t, db)
	tenantID2, realmID2 := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create JWKs for tenant1
	for i := 0; i < 3; i++ {
		elasticJWK := &domain.ElasticJWK{
			ID:                   googleUuid.New(),
			TenantID:             tenantID1,
			RealmID:              realmID1,
			KID:                  fmt.Sprintf("tenant1-kid-%d", i),
			KTY:                  "RSA",
			ALG:                  "RS256",
			USE:                  "sig",
			MaxMaterials:         1000,
			CurrentMaterialCount: 0,
		}
		require.NoError(t, repo.Create(ctx, elasticJWK))
	}

	// Create JWKs for tenant2
	for i := 0; i < 2; i++ {
		elasticJWK := &domain.ElasticJWK{
			ID:                   googleUuid.New(),
			TenantID:             tenantID2,
			RealmID:              realmID2,
			KID:                  fmt.Sprintf("tenant2-kid-%d", i),
			KTY:                  "EC",
			ALG:                  "ES256",
			USE:                  "sig",
			MaxMaterials:         500,
			CurrentMaterialCount: 0,
		}
		require.NoError(t, repo.Create(ctx, elasticJWK))
	}

	// List tenant1 - should get 3
	results, err := repo.List(ctx, tenantID1, realmID1, 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 3)

	// List tenant2 - should get 2
	results, err = repo.List(ctx, tenantID2, realmID2, 0, 10)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestElasticJWKGormRepository_IncrementMaterialCount(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create test data
	elasticJWK := &domain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  "test-kid-inc",
		KTY:                  "RSA",
		ALG:                  "RS256",
		USE:                  "sig",
		MaxMaterials:         1000,
		CurrentMaterialCount: 0,
	}
	require.NoError(t, repo.Create(ctx, elasticJWK))

	// Increment count
	err := repo.IncrementMaterialCount(ctx, elasticJWK.ID)
	require.NoError(t, err)

	// Verify count increased
	result, err := repo.Get(ctx, tenantID, realmID, "test-kid-inc")
	require.NoError(t, err)
	require.Equal(t, 1, result.CurrentMaterialCount)

	// Increment again
	err = repo.IncrementMaterialCount(ctx, elasticJWK.ID)
	require.NoError(t, err)

	result, err = repo.Get(ctx, tenantID, realmID, "test-kid-inc")
	require.NoError(t, err)
	require.Equal(t, 2, result.CurrentMaterialCount)
}

func TestElasticJWKGormRepository_IncrementMaterialCount_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Try to increment non-existent ID (should succeed but affect 0 rows).
	err := repo.IncrementMaterialCount(ctx, googleUuid.New())
	require.NoError(t, err) // GORM doesn't error on 0 rows affected.
}

func TestElasticJWKGormRepository_Create_DuplicateKID(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	elasticJWK1 := &domain.ElasticJWK{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		KID:          "duplicate-kid",
		KTY:          "RSA",
		ALG:          "RS256",
		USE:          "sig",
		MaxMaterials: 1000,
	}
	require.NoError(t, repo.Create(ctx, elasticJWK1))

	// Try to create another with same KID (should fail due to unique constraint).
	elasticJWK2 := &domain.ElasticJWK{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		KID:          "duplicate-kid", // Same KID.
		KTY:          "EC",
		ALG:          "ES256",
		USE:          "sig",
		MaxMaterials: 500,
	}
	err := repo.Create(ctx, elasticJWK2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create elastic JWK")
}

func TestElasticJWKGormRepository_WithTransaction(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create a transaction and add it to context.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilTemplateRepository.WithTransaction(ctx, tx)

	// Create within transaction.
	elasticJWK := &domain.ElasticJWK{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		KID:          "tx-test-kid",
		KTY:          "RSA",
		ALG:          "RS256",
		USE:          "sig",
		MaxMaterials: 1000,
	}
	err := repo.Create(txCtx, elasticJWK)
	require.NoError(t, err)

	// Commit transaction.
	require.NoError(t, tx.Commit().Error)

	// Verify it was persisted.
	result, err := repo.Get(ctx, tenantID, realmID, "tx-test-kid")
	require.NoError(t, err)
	require.Equal(t, elasticJWK.ID, result.ID)
}

func TestElasticJWKGormRepository_WithTransaction_Rollback(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := repository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create a transaction and add it to context.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilTemplateRepository.WithTransaction(ctx, tx)

	// Create within transaction.
	elasticJWK := &domain.ElasticJWK{
		ID:           googleUuid.New(),
		TenantID:     tenantID,
		RealmID:      realmID,
		KID:          "rollback-test-kid",
		KTY:          "RSA",
		ALG:          "RS256",
		USE:          "sig",
		MaxMaterials: 1000,
	}
	err := repo.Create(txCtx, elasticJWK)
	require.NoError(t, err)

	// Rollback transaction.
	require.NoError(t, tx.Rollback().Error)

	// Verify it was NOT persisted.
	_, err = repo.Get(ctx, tenantID, realmID, "rollback-test-kid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "elastic JWK not found")
}
