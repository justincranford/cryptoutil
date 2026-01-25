// Copyright (c) 2025 Justin Cranford
//
//

package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver
)

// Package-level shared database for all tests (TestMain pattern).
var sharedTestDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Generate unique database identifier
	dbID, err := googleUuid.NewV7()
	if err != nil {
		panic(fmt.Sprintf("failed to generate database UUID: %v", err))
	}

	// Open SQL database with modernc driver (CGO-free)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		panic(fmt.Sprintf("failed to open SQLite: %v", err))
	}

	// Configure SQLite for concurrent operations
	if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic(fmt.Sprintf("failed to enable WAL mode: %v", err))
	}

	if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic(fmt.Sprintf("failed to set busy timeout: %v", err))
	}

	// Create GORM connection
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to initialize GORM: %v", err))
	}

	// Configure connection pool
	gormDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("failed to get database instance: %v", err))
	}

	gormDB.SetMaxOpenConns(5)
	gormDB.SetMaxIdleConns(5)
	gormDB.SetConnMaxLifetime(0) // In-memory: never close connections
	gormDB.SetConnMaxIdleTime(0)

	// Auto-migrate schema
	err = db.AutoMigrate(
		// Template models needed for tenant/realm foreign key support
		&cryptoutilAppsTemplateServiceServerRepository.Tenant{},
		&cryptoutilAppsTemplateServiceServerRepository.TenantRealm{},
		// JOSE domain models
		&cryptoutilJoseDomain.ElasticJWK{},
		&cryptoutilJoseDomain.MaterialJWK{},
		&cryptoutilJoseDomain.AuditConfig{},
		&cryptoutilJoseDomain.AuditLogEntry{},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	// Assign to package-level variable
	sharedTestDB = db

	// Run all tests
	exitCode := m.Run()

	// Cleanup
	_ = sqlDB.Close()

	os.Exit(exitCode)
}

// setupTestDB is DEPRECATED - use shared sharedTestDB from TestMain instead.
// This function is kept for backward compatibility during incremental refactoring.
// New tests should use sharedTestDB directly without calling setupTestDB.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	return sharedTestDB
}

func createTestTenantAndRealm(t *testing.T, db *gorm.DB) (tenantID, realmID googleUuid.UUID) {
	t.Helper()

	tenantID = googleUuid.New()
	realmID = googleUuid.New()

	// Create tenant using GORM model.
	tenant := &cryptoutilAppsTemplateServiceServerRepository.Tenant{
		ID:          tenantID,
		Name:        "Test Tenant " + tenantID.String(),
		Description: "Test tenant for repository tests",
		Active:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, db.Create(tenant).Error)

	// Create tenant_realm using GORM model.
	tenantRealm := &cryptoutilAppsTemplateServiceServerRepository.TenantRealm{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	var result cryptoutilJoseDomain.ElasticJWK

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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create test data
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create JWK for tenant1
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create multiple JWKs
	for i := 0; i < 5; i++ {
		elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create JWKs for tenant1
	for i := 0; i < 3; i++ {
		elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
		elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create test data
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Try to increment non-existent ID (should succeed but affect 0 rows).
	err := repo.IncrementMaterialCount(ctx, googleUuid.New())
	require.NoError(t, err) // GORM doesn't error on 0 rows affected.
}

func TestElasticJWKGormRepository_Create_DuplicateKID(t *testing.T) {
	db := setupTestDB(t)
	tenantID, realmID := createTestTenantAndRealm(t, db)
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	elasticJWK1 := &cryptoutilJoseDomain.ElasticJWK{
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
	elasticJWK2 := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create a transaction and add it to context.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	// Create within transaction.
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
	repo := cryptoutilJoseRepository.NewElasticJWKRepository(db)

	ctx := context.Background()

	// Create a transaction and add it to context.
	tx := db.Begin()
	require.NoError(t, tx.Error)

	txCtx := cryptoutilAppsTemplateServiceServerRepository.WithTransaction(ctx, tx)

	// Create within transaction.
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
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
