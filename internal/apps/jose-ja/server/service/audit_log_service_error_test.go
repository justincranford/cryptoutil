package service

import (
	"context"
	"strings"
	"testing"

	cryptoutilAppsJoseJaModel "cryptoutil/internal/apps/jose-ja/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestAuditLogService_CleanupOldLogsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err := svc.CleanupOldLogs(ctx, googleUuid.New(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit logs"))
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

func TestAuditLogService_ListAuditLogsDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err := svc.ListAuditLogs(ctx, googleUuid.New(), 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

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

func TestAuditLogService_UpdateAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	elasticRepo, _, auditLogRepo, auditConfigRepo := newClosedServiceDeps(t)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	tenantID := googleUuid.New()
	config := &cryptoutilAppsJoseJaModel.AuditConfig{
		TenantID:     tenantID,
		Operation:    "test-operation",
		Enabled:      true,
		SamplingRate: cryptoutilSharedMagic.TestProbAlways,
	}

	err := svc.UpdateAuditConfig(ctx, tenantID, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update audit config"))
}

func TestAuditLogsByElasticJWK_TenantMismatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
	tenantID := googleUuid.New()

	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaModel.KeyUseSig, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.NoError(t, err)

	differentTenantID := googleUuid.New()

	_, _, err = auditSvc.ListAuditLogsByElasticJWK(ctx, differentTenantID, elasticJWK.ID, 0, cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	require.Error(t, err)
	require.Contains(t, err.Error(), "elastic JWK not found")
}
