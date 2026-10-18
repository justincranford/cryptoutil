// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"errors"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var errJWKServiceInjected = errors.New("injected repository failure")

func TestAuditLogService_ErrorPaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWKID := newTestUUID(t)
	elasticRepo := &elasticJWKRepositoryStub{
		getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
			return &cryptoutilAppsSmKmsModel.ElasticJWK{ID: elasticJWKID, TenantID: tenantID}, nil
		},
	}
	logRepo := &auditLogRepositoryStub{}
	configRepo := &auditConfigRepositoryStub{}
	service := NewAuditLogService(logRepo, configRepo, elasticRepo)

	configRepo.shouldAuditFn = func(context.Context, googleUuid.UUID, string) (bool, error) {
		return false, errJWKServiceInjected
	}
	logRepo.createFn = func(_ context.Context, entry *cryptoutilAppsSmKmsModel.AuditLogEntry) error {
		require.Equal(t, tenantID, entry.TenantID)
		require.Equal(t, cryptoutilAppsSmKmsModel.OperationSign, entry.Operation)

		return nil
	}

	require.NoError(t, service.LogOperation(testContext, tenantID, nil, cryptoutilAppsSmKmsModel.OperationSign, "fallback-audit", true, nil))

	configRepo.shouldAuditFn = func(context.Context, googleUuid.UUID, string) (bool, error) {
		return false, nil
	}
	logRepo.createFn = func(context.Context, *cryptoutilAppsSmKmsModel.AuditLogEntry) error {
		t.Fatal("disabled operation should not create an audit entry")

		return nil
	}

	require.NoError(t, service.LogOperation(testContext, tenantID, nil, cryptoutilAppsSmKmsModel.OperationSign, "disabled-audit", true, nil))

	configRepo.shouldAuditFn = func(context.Context, googleUuid.UUID, string) (bool, error) {
		return true, nil
	}
	logRepo.createFn = func(context.Context, *cryptoutilAppsSmKmsModel.AuditLogEntry) error {
		return errJWKServiceInjected
	}

	require.ErrorContains(t, service.LogOperation(testContext, tenantID, nil, cryptoutilAppsSmKmsModel.OperationSign, "create-error", true, nil), "failed to create audit log")

	logRepo.listFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, _, err := service.ListAuditLogs(testContext, tenantID, 0, 1)
	require.ErrorContains(t, err, "failed to list audit logs")

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return nil, errJWKServiceInjected
	}
	_, _, err = service.ListAuditLogsByElasticJWK(testContext, tenantID, elasticJWKID, 0, 1)
	require.ErrorContains(t, err, "failed to get elastic JWK")

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return &cryptoutilAppsSmKmsModel.ElasticJWK{ID: elasticJWKID, TenantID: tenantID}, nil
	}
	logRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, _, err = service.ListAuditLogsByElasticJWK(testContext, tenantID, elasticJWKID, 0, 1)
	require.ErrorContains(t, err, "failed to list audit logs")
	_, _, err = service.ListAuditLogsByElasticJWK(testContext, newTestUUID(t), elasticJWKID, 0, 1)
	require.ErrorContains(t, err, "not found")

	logRepo.listByOperationFn = func(context.Context, googleUuid.UUID, string, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, _, err = service.ListAuditLogsByOperation(testContext, tenantID, cryptoutilAppsSmKmsModel.OperationSign, 0, 1)
	require.ErrorContains(t, err, "failed to list audit logs by operation")

	configRepo.getAllFn = func(context.Context, googleUuid.UUID) ([]*cryptoutilAppsSmKmsModel.AuditConfig, error) {
		return nil, errJWKServiceInjected
	}
	_, err = service.GetAuditConfig(testContext, tenantID)
	require.ErrorContains(t, err, "failed to get audit configs")

	configRepo.getAllFn = func(context.Context, googleUuid.UUID) ([]*cryptoutilAppsSmKmsModel.AuditConfig, error) {
		return []*cryptoutilAppsSmKmsModel.AuditConfig{{Operation: cryptoutilAppsSmKmsModel.OperationSign}}, nil
	}
	config, err := service.GetAuditConfig(testContext, tenantID)
	require.NoError(t, err)
	require.Equal(t, cryptoutilAppsSmKmsModel.OperationSign, config.Operation)

	configRepo.upsertFn = func(context.Context, *cryptoutilAppsSmKmsModel.AuditConfig) error {
		return errJWKServiceInjected
	}

	require.ErrorContains(t, service.UpdateAuditConfig(testContext, tenantID, &cryptoutilAppsSmKmsModel.AuditConfig{}), "failed to update audit config")

	logRepo.deleteOlderThanFn = func(context.Context, googleUuid.UUID, int) (int64, error) {
		return 0, errJWKServiceInjected
	}
	_, err = service.CleanupOldLogs(testContext, tenantID, 1)
	require.ErrorContains(t, err, "failed to delete old audit logs")
}

func TestElasticJWKService_RepositoryAndMaterialErrors(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWKID := newTestUUID(t)
	elasticJWK := &cryptoutilAppsSmKmsModel.ElasticJWK{
		ID:           elasticJWKID,
		TenantID:     tenantID,
		Algorithm:    cryptoutilSharedMagic.JoseAlgES256,
		Use:          cryptoutilAppsSmKmsModel.KeyUseSig,
		MaxMaterials: 2,
	}

	elasticRepo := &elasticJWKRepositoryStub{
		createFn: func(context.Context, *cryptoutilAppsSmKmsModel.ElasticJWK) error {
			return errJWKServiceInjected
		},
	}
	materialRepo := &materialJWKRepositoryStub{}
	service := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	_, _, err := service.CreateElasticJWK(testContext, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 1)
	require.ErrorContains(t, err, "failed to create elastic JWK")

	elasticRepo.createFn = func(context.Context, *cryptoutilAppsSmKmsModel.ElasticJWK) error {
		return nil
	}

	var deleted bool

	elasticRepo.deleteFn = func(context.Context, googleUuid.UUID) error {
		deleted = true

		return nil
	}
	materialRepo.createFn = func(context.Context, *cryptoutilAppsSmKmsModel.MaterialJWK) error {
		return errJWKServiceInjected
	}
	_, _, err = service.CreateElasticJWK(testContext, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 1)
	require.ErrorContains(t, err, "failed to create initial material")
	require.True(t, deleted)

	elasticService, ok := service.(*elasticJWKServiceImpl)
	require.True(t, ok)

	_, err = elasticService.createMaterialJWK(testContext, &cryptoutilAppsSmKmsModel.ElasticJWK{Algorithm: testErrorUnsupported}, true)
	require.Error(t, err)
	require.ErrorContains(t, err, "unsupported algorithm for key generation")

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return nil, errJWKServiceInjected
	}
	_, err = service.GetElasticJWK(testContext, tenantID, elasticJWKID)
	require.ErrorContains(t, err, "failed to get elastic JWK")

	elasticRepo.listFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
		return nil, 0, errJWKServiceInjected
	}
	_, _, err = service.ListElasticJWKs(testContext, tenantID, 0, 1)
	require.ErrorContains(t, err, "failed to list elastic JWKs")

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return elasticJWK, nil
	}
	materialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		return nil, 0, errJWKServiceInjected
	}

	require.ErrorContains(t, service.DeleteElasticJWK(testContext, tenantID, elasticJWKID), "failed to list materials")

	materialRepo.listByElasticJWKFn = func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
		return []*cryptoutilAppsSmKmsModel.MaterialJWK{{ID: newTestUUID(t)}}, 1, nil
	}
	materialRepo.deleteFn = func(context.Context, googleUuid.UUID) error {
		return errJWKServiceInjected
	}

	require.ErrorContains(t, service.DeleteElasticJWK(testContext, tenantID, elasticJWKID), "failed to delete material")

	materialRepo.deleteFn = func(context.Context, googleUuid.UUID) error {
		return nil
	}
	elasticRepo.deleteFn = func(context.Context, googleUuid.UUID) error {
		return errJWKServiceInjected
	}

	require.ErrorContains(t, service.DeleteElasticJWK(testContext, tenantID, elasticJWKID), "failed to delete elastic JWK")

	elasticRepo.getByIDFn = func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
		return nil, errJWKServiceInjected
	}

	require.ErrorContains(t, service.DeleteElasticJWK(testContext, tenantID, elasticJWKID), "failed to get elastic JWK")
}

func TestElasticJWKService_CanceledMaterialCreationCleansUp(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticRepo := &elasticJWKRepositoryStub{}
	materialRepo := &materialJWKRepositoryStub{}

	var deleted bool

	elasticRepo.deleteFn = func(context.Context, googleUuid.UUID) error {
		deleted = true

		return nil
	}
	service := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)
	ctx, cancel := context.WithCancel(testContext)
	cancel()

	_, _, err := service.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 1)
	require.Error(t, err)
	require.True(t, deleted)
}
