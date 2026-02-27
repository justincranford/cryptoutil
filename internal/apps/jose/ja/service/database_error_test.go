// Copyright (c) 2025 Justin Cranford
//

package service

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

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// createClosedServiceDependencies creates a new in-memory SQLite database,
// applies migrations, creates all repositories and services, then closes
// the database connection to force database errors in service operations.
func createClosedServiceDependencies() (*gorm.DB, cryptoutilAppsJoseJaRepository.ElasticJWKRepository, cryptoutilAppsJoseJaRepository.MaterialJWKRepository, cryptoutilAppsJoseJaRepository.AuditLogRepository, cryptoutilAppsJoseJaRepository.AuditConfigRepository, error) {
	ctx := context.Background()
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to open SQLite: %w", err)
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

		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	// Apply migrations.
	if err := cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(sqlDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite); err != nil {
		_ = sqlDB.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Create repositories BEFORE closing.
	elasticRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(gormDB)
	materialRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(gormDB)
	auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(gormDB)
	auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(gormDB)

	// Close the underlying connection to force errors.
	if err := sqlDB.Close(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to close database: %w", err)
	}

	return gormDB, elasticRepo, materialRepo, auditLogRepo, auditConfigRepo, nil
}

// ====================
// ElasticJWK Service Database Error Tests
// ====================

func TestElasticJWKService_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Valid parameters, but database is closed.
	_, _, err = svc.CreateElasticJWK(ctx, googleUuid.New(), cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKService_GetDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKService_ListDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, _, err = svc.ListElasticJWKs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list elastic JWKs"))
}

func TestElasticJWKService_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err = svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on any database operation.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// AuditLog Service Database Error Tests
// ====================

func TestAuditLogService_LogOperationDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	elasticJWKID := googleUuid.New()
	err = svc.LogOperation(ctx, googleUuid.New(), &elasticJWKID, "test-operation", "test-request-id", true, nil)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create audit log"))
}

func TestAuditLogService_ListAuditLogsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_ListAuditLogsByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogsByElasticJWK(ctx, googleUuid.New(), googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on get elastic JWK or list logs.
	require.True(t,
		strings.Contains(err.Error(), "failed to get elastic JWK") ||
			strings.Contains(err.Error(), "failed to list audit logs"),
		"Expected get or list error, got: %v", err)
}

func TestAuditLogService_ListAuditLogsByOperationDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogsByOperation(ctx, googleUuid.New(), "test-operation", 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_GetAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err = svc.GetAuditConfig(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditLogService_UpdateAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	tenantID := googleUuid.New()
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    "test-operation",
		Enabled:      true,
		SamplingRate: cryptoutilSharedMagic.TestProbAlways,
	}

	err = svc.UpdateAuditConfig(ctx, tenantID, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update audit config"))
}

func TestAuditLogService_CleanupOldLogsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err = svc.CleanupOldLogs(ctx, googleUuid.New(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit logs"))
}

// ====================
// MaterialRotation Service Database Error Tests
// ====================

func TestMaterialRotationService_ListMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.ListMaterials(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or list materials.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetActiveMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or get active material.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetMaterialByKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetMaterialByKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid")
	require.Error(t, err)
	// Could fail on get elastic JWK or get material by KID.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.RotateMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or later operations.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err = svc.RetireMaterial(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK, get material by KID, or update.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWE Service Database Error Tests
// ====================
