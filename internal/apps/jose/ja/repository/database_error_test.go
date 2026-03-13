// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilTestdb "cryptoutil/internal/apps/template/service/testing/testdb"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// newClosedDB creates a closed SQLite DB using the shared testdb helper.
func newClosedDB(t *testing.T) *gorm.DB {
	t.Helper()

	return cryptoutilTestdb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error {
		return ApplyJoseJAMigrations(sqlDB, DatabaseTypeSQLite)
	})
}

// ====================
// ElasticJWK Repository Database Error Tests
// ====================

func TestElasticJWKRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-create-error",
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	}

	err := repo.Create(ctx, jwk)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKRepository_GetDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, err := repo.Get(ctx, googleUuid.New(), "test-kid")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKRepository_GetByIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, err := repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK by ID"))
}

func TestElasticJWKRepository_ListDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	_, _, err := repo.List(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count elastic JWKs") ||
			strings.Contains(err.Error(), "failed to list elastic JWKs"),
		"Expected count or list error, got: %v", err)
}

func TestElasticJWKRepository_UpdateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	id, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	tenantID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()

	jwk := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:           *id,
		TenantID:     *tenantID,
		KID:          "test-update-error",
		KeyType:      cryptoutilAppsJoseJaDomain.KeyTypeRSA,
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
	}

	err := repo.Update(ctx, jwk)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update elastic JWK"))
}

func TestElasticJWKRepository_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err := repo.Delete(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete elastic JWK"))
}

func TestElasticJWKRepository_IncrementMaterialCountDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err := repo.IncrementMaterialCount(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to increment material count"))
}

func TestElasticJWKRepository_DecrementMaterialCountDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewElasticJWKRepository(closedDB)

	err := repo.DecrementMaterialCount(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to decrement material count"))
}

// ====================
// MaterialJWK Repository Database Error Tests
// ====================

func TestMaterialJWKRepository_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

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

	err := repo.Create(ctx, material)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create material JWK"))
}

func TestMaterialJWKRepository_GetByMaterialKIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetByMaterialKID(ctx, "test-kid")
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by KID"))
}

func TestMaterialJWKRepository_GetByIDDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetByID(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get material JWK by ID"))
}

func TestMaterialJWKRepository_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, err := repo.GetActiveMaterial(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get active material JWK"))
}

func TestMaterialJWKRepository_ListByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	closedDB := newClosedDB(t)

	ctx := context.Background()
	repo := NewMaterialJWKRepository(closedDB)

	_, _, err := repo.ListByElasticJWK(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count material JWKs") ||
			strings.Contains(err.Error(), "failed to list material JWKs"),
		"Expected count or list error, got: %v", err)
}
