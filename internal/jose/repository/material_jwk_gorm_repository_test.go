// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite"
)

// setupMaterialJWKTestDB creates an in-memory SQLite database for testing.
func setupMaterialJWKTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()
	dsn := "file::memory:?cache=shared"

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// Auto-migrate models (template models first, then JOSE domain models).
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
		&cryptoutilJoseDomain.ElasticJWK{},
		&cryptoutilJoseDomain.MaterialJWK{},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// createMaterialJWKTestTenantAndRealm creates a test tenant and realm, returning their IDs.
func createMaterialJWKTestTenantAndRealm(t *testing.T, db *gorm.DB) (googleUuid.UUID, googleUuid.UUID) {
	t.Helper()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	tenant := cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          tenantID,
		Name:        "test-tenant-" + tenantID.String(),
		Description: "Test tenant for MaterialJWK tests",
		Active:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := db.Create(&tenant).Error
	require.NoError(t, err)

	realm := cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
		ID:        realmID,
		TenantID:  tenantID,
		RealmID:   googleUuid.New(),
		Type:      "username_password",
		Active:    true,
		Source:    "db",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = db.Create(&realm).Error
	require.NoError(t, err)

	return tenantID, realmID
}

// createTestElasticJWK creates a test Elastic JWK and returns it.
func createTestElasticJWK(t *testing.T, db *gorm.DB, tenantID, realmID googleUuid.UUID) *cryptoutilJoseDomain.ElasticJWK {
	t.Helper()

	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
		ID:                   googleUuid.New(),
		TenantID:             tenantID,
		RealmID:              realmID,
		KID:                  "elastic-kid-" + googleUuid.New().String(),
		KTY:                  "EC",
		ALG:                  "ES256",
		USE:                  "sig",
		MaxMaterials:         1000,
		CurrentMaterialCount: 0,
		CreatedAt:            time.Now().UnixMilli(),
	}
	err := db.Create(elasticJWK).Error
	require.NoError(t, err)

	return elasticJWK
}

func TestMaterialJWKGormRepository_Create(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	materialJWK := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "material-kid-" + googleUuid.New().String(),
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         true,
		BarrierVersion: 1,
	}

	err := repo.Create(ctx, materialJWK)
	require.NoError(t, err)

	// Verify it was created.
	var found cryptoutilJoseDomain.MaterialJWK

	err = db.First(&found, "id = ?", materialJWK.ID).Error
	require.NoError(t, err)
	require.Equal(t, materialJWK.MaterialKID, found.MaterialKID)
	require.Equal(t, materialJWK.Active, found.Active)
}

func TestMaterialJWKGormRepository_GetByMaterialKID(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	materialKID := "material-kid-" + googleUuid.New().String()
	materialJWK := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         true,
		BarrierVersion: 1,
	}
	err := db.Create(materialJWK).Error
	require.NoError(t, err)

	// Get by material KID.
	found, err := repo.GetByMaterialKID(ctx, elasticJWK.ID, materialKID)
	require.NoError(t, err)
	require.Equal(t, materialJWK.ID, found.ID)
	require.Equal(t, materialKID, found.MaterialKID)
}

func TestMaterialJWKGormRepository_GetByMaterialKID_NotFound(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	_, err := repo.GetByMaterialKID(ctx, elasticJWK.ID, "non-existent-material-kid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "material JWK not found")
}

func TestMaterialJWKGormRepository_ListByElasticJWK(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Create 3 material JWKs.
	for i := 0; i < 3; i++ {
		materialJWK := &cryptoutilJoseDomain.MaterialJWK{
			ID:             googleUuid.New(),
			ElasticJWKID:   elasticJWK.ID,
			MaterialKID:    "material-kid-" + googleUuid.New().String(),
			PrivateJWKJWE:  "encrypted-private-key",
			PublicJWKJWE:   "encrypted-public-key",
			Active:         i == 0, // Only first is active.
			BarrierVersion: 1,
		}
		err := db.Create(materialJWK).Error
		require.NoError(t, err)
	}

	// List all.
	materials, err := repo.ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)
	require.NoError(t, err)
	require.Len(t, materials, 3)

	// Test pagination.
	materials, err = repo.ListByElasticJWK(ctx, elasticJWK.ID, 0, 2)
	require.NoError(t, err)
	require.Len(t, materials, 2)

	materials, err = repo.ListByElasticJWK(ctx, elasticJWK.ID, 2, 2)
	require.NoError(t, err)
	require.Len(t, materials, 1)
}

func TestMaterialJWKGormRepository_ListByElasticJWK_Empty(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// List with no materials.
	materials, err := repo.ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)
	require.NoError(t, err)
	require.Empty(t, materials)
}

func TestMaterialJWKGormRepository_GetActiveMaterial(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Create inactive material.
	inactiveMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "inactive-material-kid",
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         false,
		BarrierVersion: 1,
	}
	err := db.Create(inactiveMaterial).Error
	require.NoError(t, err)

	// Create active material.
	activeMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "active-material-kid",
		PrivateJWKJWE:  "encrypted-private-key-active",
		PublicJWKJWE:   "encrypted-public-key-active",
		Active:         true,
		BarrierVersion: 1,
	}
	err = db.Create(activeMaterial).Error
	require.NoError(t, err)

	// Get active material.
	found, err := repo.GetActiveMaterial(ctx, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, activeMaterial.ID, found.ID)
	require.True(t, found.Active)
}

func TestMaterialJWKGormRepository_GetActiveMaterial_NotFound(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// No materials at all - should return not found.
	_, err := repo.GetActiveMaterial(ctx, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no active material JWK found")
}

func TestMaterialJWKGormRepository_GetActiveMaterial_NoActive(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Create only inactive material.
	inactiveMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "inactive-material-kid",
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         false,
		BarrierVersion: 1,
	}
	err := db.Create(inactiveMaterial).Error
	require.NoError(t, err)

	// Should return not found since no active material exists.
	_, err = repo.GetActiveMaterial(ctx, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no active material JWK found")
}

func TestMaterialJWKGormRepository_RotateMaterial(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Create initial active material.
	oldMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "old-material-kid",
		PrivateJWKJWE:  "encrypted-private-key-old",
		PublicJWKJWE:   "encrypted-public-key-old",
		Active:         true,
		BarrierVersion: 1,
	}
	err := db.Create(oldMaterial).Error
	require.NoError(t, err)

	// Rotate to new material.
	newMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		MaterialKID:    "new-material-kid",
		PrivateJWKJWE:  "encrypted-private-key-new",
		PublicJWKJWE:   "encrypted-public-key-new",
		BarrierVersion: 2,
	}
	err = repo.RotateMaterial(ctx, elasticJWK.ID, newMaterial)
	require.NoError(t, err)

	// Verify old material is now inactive with retired_at set.
	var oldFound cryptoutilJoseDomain.MaterialJWK

	err = db.First(&oldFound, "id = ?", oldMaterial.ID).Error
	require.NoError(t, err)
	require.False(t, oldFound.Active)
	require.NotNil(t, oldFound.RetiredAt)

	// Verify new material is active.
	var newFound cryptoutilJoseDomain.MaterialJWK

	err = db.First(&newFound, "id = ?", newMaterial.ID).Error
	require.NoError(t, err)
	require.True(t, newFound.Active)
	require.Equal(t, elasticJWK.ID, newFound.ElasticJWKID)
}

func TestMaterialJWKGormRepository_RotateMaterial_NoExistingActive(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Rotate without any existing material (first rotation).
	newMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		MaterialKID:    "first-material-kid",
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		BarrierVersion: 1,
	}
	err := repo.RotateMaterial(ctx, elasticJWK.ID, newMaterial)
	require.NoError(t, err)

	// Verify new material is active.
	var found cryptoutilJoseDomain.MaterialJWK

	err = db.First(&found, "id = ?", newMaterial.ID).Error
	require.NoError(t, err)
	require.True(t, found.Active)
}

func TestMaterialJWKGormRepository_CountMaterials(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Count with no materials.
	count, err := repo.CountMaterials(ctx, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(0), count)

	// Create 3 materials.
	for i := 0; i < 3; i++ {
		materialJWK := &cryptoutilJoseDomain.MaterialJWK{
			ID:             googleUuid.New(),
			ElasticJWKID:   elasticJWK.ID,
			MaterialKID:    "material-kid-" + googleUuid.New().String(),
			PrivateJWKJWE:  "encrypted-private-key",
			PublicJWKJWE:   "encrypted-public-key",
			Active:         i == 0,
			BarrierVersion: 1,
		}
		err := db.Create(materialJWK).Error
		require.NoError(t, err)
	}

	// Count should be 3.
	count, err = repo.CountMaterials(ctx, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)
}

func TestMaterialJWKGormRepository_CountMaterials_Isolation(t *testing.T) {
	db := setupMaterialJWKTestDB(t)

	// Create two elastic JWKs with different tenant/realm.
	tenantID1, realmID1 := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK1 := createTestElasticJWK(t, db, tenantID1, realmID1)

	tenantID2, realmID2 := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK2 := createTestElasticJWK(t, db, tenantID2, realmID2)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Create 2 materials for elastic JWK 1.
	for i := 0; i < 2; i++ {
		materialJWK := &cryptoutilJoseDomain.MaterialJWK{
			ID:             googleUuid.New(),
			ElasticJWKID:   elasticJWK1.ID,
			MaterialKID:    "material-kid-e1-" + googleUuid.New().String(),
			PrivateJWKJWE:  "encrypted-private-key",
			PublicJWKJWE:   "encrypted-public-key",
			Active:         i == 0,
			BarrierVersion: 1,
		}
		err := db.Create(materialJWK).Error
		require.NoError(t, err)
	}

	// Create 5 materials for elastic JWK 2.
	for i := 0; i < 5; i++ {
		materialJWK := &cryptoutilJoseDomain.MaterialJWK{
			ID:             googleUuid.New(),
			ElasticJWKID:   elasticJWK2.ID,
			MaterialKID:    "material-kid-e2-" + googleUuid.New().String(),
			PrivateJWKJWE:  "encrypted-private-key",
			PublicJWKJWE:   "encrypted-public-key",
			Active:         i == 0,
			BarrierVersion: 1,
		}
		err := db.Create(materialJWK).Error
		require.NoError(t, err)
	}

	// Count for elastic JWK 1 should be 2.
	count1, err := repo.CountMaterials(ctx, elasticJWK1.ID)
	require.NoError(t, err)
	require.Equal(t, int64(2), count1)

	// Count for elastic JWK 2 should be 5.
	count2, err := repo.CountMaterials(ctx, elasticJWK2.ID)
	require.NoError(t, err)
	require.Equal(t, int64(5), count2)
}

func TestMaterialJWKGormRepository_WithTransaction(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Start transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	materialJWK := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "tx-material-kid",
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         true,
		BarrierVersion: 1,
	}

	err := repo.Create(txCtx, materialJWK)
	require.NoError(t, err)

	// Commit transaction.
	err = tx.Commit().Error
	require.NoError(t, err)

	// Verify it was persisted.
	found, err := repo.GetByMaterialKID(ctx, elasticJWK.ID, "tx-material-kid")
	require.NoError(t, err)
	require.Equal(t, materialJWK.ID, found.ID)
}

func TestMaterialJWKGormRepository_WithTransaction_Rollback(t *testing.T) {
	db := setupMaterialJWKTestDB(t)
	tenantID, realmID := createMaterialJWKTestTenantAndRealm(t, db)
	elasticJWK := createTestElasticJWK(t, db, tenantID, realmID)

	repo := cryptoutilJoseRepository.NewMaterialJWKRepository(db)
	ctx := context.Background()

	// Start transaction.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	materialJWK := &cryptoutilJoseDomain.MaterialJWK{
		ID:             googleUuid.New(),
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    "rollback-material-kid",
		PrivateJWKJWE:  "encrypted-private-key",
		PublicJWKJWE:   "encrypted-public-key",
		Active:         true,
		BarrierVersion: 1,
	}

	err := repo.Create(txCtx, materialJWK)
	require.NoError(t, err)

	// Rollback transaction.
	err = tx.Rollback().Error
	require.NoError(t, err)

	// Verify it was NOT persisted.
	_, err = repo.GetByMaterialKID(ctx, elasticJWK.ID, "rollback-material-kid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "material JWK not found")
}
