// Copyright (c) 2025 Justin Cranford
//

package repository

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"fmt"
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

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
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
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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

	_, _, err = repo.List(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
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
		Algorithm:    cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		Use:          cryptoutilSharedMagic.JoseKeyUseSig,
		MaxMaterials: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
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

	_, _, err = repo.ListByElasticJWK(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on Count or Find - either error path is valid.
	require.True(t,
		strings.Contains(err.Error(), "failed to count material JWKs") ||
			strings.Contains(err.Error(), "failed to list material JWKs"),
		"Expected count or list error, got: %v", err)
}
