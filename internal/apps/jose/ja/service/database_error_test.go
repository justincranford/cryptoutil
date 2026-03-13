// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilTestdb "cryptoutil/internal/apps/template/service/testing/testdb"
)

// newClosedServiceDeps creates a closed SQLite database with migrations applied,
// then creates all repositories. The closed DB forces database errors in service operations.
func newClosedServiceDeps(t *testing.T) (cryptoutilAppsJoseJaRepository.ElasticJWKRepository, cryptoutilAppsJoseJaRepository.MaterialJWKRepository, cryptoutilAppsJoseJaRepository.AuditLogRepository, cryptoutilAppsJoseJaRepository.AuditConfigRepository) {
	t.Helper()

	closedDB := cryptoutilTestdb.NewClosedSQLiteDB(t, func(sqlDB *sql.DB) error {
		return cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(sqlDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite)
	})

	elasticRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(closedDB)
	materialRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(closedDB)
	auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(closedDB)
	auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(closedDB)

	return elasticRepo, materialRepo, auditLogRepo, auditConfigRepo
}

// ====================
// ElasticJWK Service Database Error Tests
// ====================

func TestElasticJWKService_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Valid parameters, but database is closed.
	_, _, err := svc.CreateElasticJWK(ctx, googleUuid.New(), cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, cryptoutilAppsJoseJaDomain.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKService_GetDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKService_ListDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, _, err := svc.ListElasticJWKs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list elastic JWKs"))
}

func TestElasticJWKService_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err := svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
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

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	elasticJWKID := googleUuid.New()
	err := svc.LogOperation(ctx, googleUuid.New(), &elasticJWKID, "test-operation", "test-request-id", true, nil)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create audit log"))
}

func TestAuditLogService_ListAuditLogsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err := svc.ListAuditLogs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_ListAuditLogsByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err := svc.ListAuditLogsByElasticJWK(ctx, googleUuid.New(), googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	// Could fail on get elastic JWK or list logs.
	require.True(t,
		strings.Contains(err.Error(), "failed to get elastic JWK") ||
			strings.Contains(err.Error(), "failed to list audit logs"),
		"Expected get or list error, got: %v", err)
}

func TestAuditLogService_ListAuditLogsByOperationDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err := svc.ListAuditLogsByOperation(ctx, googleUuid.New(), "test-operation", 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_GetAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err := svc.GetAuditConfig(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditLogService_UpdateAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	tenantID := googleUuid.New()
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    "test-operation",
		Enabled:      true,
		SamplingRate: cryptoutilSharedMagic.TestProbAlways,
	}

	err := svc.UpdateAuditConfig(ctx, tenantID, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update audit config"))
}

func TestAuditLogService_CleanupOldLogsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err := svc.CleanupOldLogs(ctx, googleUuid.New(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit logs"))
}

// ====================
// MaterialRotation Service Database Error Tests
// ====================

func TestMaterialRotationService_ListMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.ListMaterials(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or list materials.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetActiveMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or get active material.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetMaterialByKIDDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.GetMaterialByKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid")
	require.Error(t, err)
	// Could fail on get elastic JWK or get material by KID.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err := svc.RotateMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or later operations.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, materialRepo, _, _ := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err := svc.RetireMaterial(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New())
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
