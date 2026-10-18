// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"errors"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilAppsSmKmsRepository "cryptoutil/internal/apps/sm-kms/server/repository"

	googleUuid "github.com/google/uuid"
)

type elasticJWKRepositoryStub struct {
	cryptoutilAppsSmKmsRepository.ElasticJWKRepository
	createFn         func(context.Context, *cryptoutilAppsSmKmsModel.ElasticJWK) error
	getByIDFn        func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error)
	listFn           func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error)
	deleteFn         func(context.Context, googleUuid.UUID) error
	incrementCountFn func(context.Context, googleUuid.UUID) error
}

func (s *elasticJWKRepositoryStub) Create(ctx context.Context, jwk *cryptoutilAppsSmKmsModel.ElasticJWK) error {
	if s.createFn != nil {
		return s.createFn(ctx, jwk)
	}

	return nil
}

func (s *elasticJWKRepositoryStub) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}

	return nil, errors.New("unexpected get by id")
}

func (s *elasticJWKRepositoryStub) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsSmKmsModel.ElasticJWK, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, tenantID, offset, limit)
	}

	return nil, 0, errors.New("unexpected list")
}

func (s *elasticJWKRepositoryStub) Delete(ctx context.Context, id googleUuid.UUID) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}

	return nil
}

func (s *elasticJWKRepositoryStub) IncrementMaterialCount(ctx context.Context, id googleUuid.UUID) error {
	if s.incrementCountFn != nil {
		return s.incrementCountFn(ctx, id)
	}

	return nil
}

type materialJWKRepositoryStub struct {
	cryptoutilAppsSmKmsRepository.MaterialJWKRepository
	createFn           func(context.Context, *cryptoutilAppsSmKmsModel.MaterialJWK) error
	getByKIDFn         func(context.Context, string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error)
	getByIDFn          func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error)
	getActiveFn        func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error)
	listByElasticJWKFn func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error)
	rotateMaterialFn   func(context.Context, googleUuid.UUID, *cryptoutilAppsSmKmsModel.MaterialJWK) error
	retireMaterialFn   func(context.Context, googleUuid.UUID) error
	deleteFn           func(context.Context, googleUuid.UUID) error
	countMaterialsFn   func(context.Context, googleUuid.UUID) (int64, error)
}

func (s *materialJWKRepositoryStub) Create(ctx context.Context, material *cryptoutilAppsSmKmsModel.MaterialJWK) error {
	if s.createFn != nil {
		return s.createFn(ctx, material)
	}

	return nil
}

func (s *materialJWKRepositoryStub) GetByMaterialKID(ctx context.Context, kid string) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	if s.getByKIDFn != nil {
		return s.getByKIDFn(ctx, kid)
	}

	return nil, errors.New("unexpected get by KID")
}

func (s *materialJWKRepositoryStub) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(ctx, id)
	}

	return nil, errors.New("unexpected get by id")
}

func (s *materialJWKRepositoryStub) GetActiveMaterial(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsSmKmsModel.MaterialJWK, error) {
	if s.getActiveFn != nil {
		return s.getActiveFn(ctx, id)
	}

	return nil, errors.New("unexpected get active material")
}

func (s *materialJWKRepositoryStub) ListByElasticJWK(ctx context.Context, id googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsSmKmsModel.MaterialJWK, int64, error) {
	if s.listByElasticJWKFn != nil {
		return s.listByElasticJWKFn(ctx, id, offset, limit)
	}

	return nil, 0, errors.New("unexpected list materials")
}

func (s *materialJWKRepositoryStub) RotateMaterial(ctx context.Context, id googleUuid.UUID, material *cryptoutilAppsSmKmsModel.MaterialJWK) error {
	if s.rotateMaterialFn != nil {
		return s.rotateMaterialFn(ctx, id, material)
	}

	return nil
}

func (s *materialJWKRepositoryStub) RetireMaterial(ctx context.Context, id googleUuid.UUID) error {
	if s.retireMaterialFn != nil {
		return s.retireMaterialFn(ctx, id)
	}

	return nil
}

func (s *materialJWKRepositoryStub) Delete(ctx context.Context, id googleUuid.UUID) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}

	return nil
}

func (s *materialJWKRepositoryStub) CountMaterials(ctx context.Context, id googleUuid.UUID) (int64, error) {
	if s.countMaterialsFn != nil {
		return s.countMaterialsFn(ctx, id)
	}

	return 0, nil
}

type auditLogRepositoryStub struct {
	cryptoutilAppsSmKmsRepository.AuditLogRepository
	createFn           func(context.Context, *cryptoutilAppsSmKmsModel.AuditLogEntry) error
	listFn             func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error)
	listByElasticJWKFn func(context.Context, googleUuid.UUID, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error)
	listByOperationFn  func(context.Context, googleUuid.UUID, string, int, int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error)
	deleteOlderThanFn  func(context.Context, googleUuid.UUID, int) (int64, error)
}

func (s *auditLogRepositoryStub) Create(ctx context.Context, entry *cryptoutilAppsSmKmsModel.AuditLogEntry) error {
	if s.createFn != nil {
		return s.createFn(ctx, entry)
	}

	return nil
}

func (s *auditLogRepositoryStub) List(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, tenantID, offset, limit)
	}

	return nil, 0, errors.New("unexpected audit list")
}

func (s *auditLogRepositoryStub) ListByElasticJWK(ctx context.Context, id googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
	if s.listByElasticJWKFn != nil {
		return s.listByElasticJWKFn(ctx, id, offset, limit)
	}

	return nil, 0, errors.New("unexpected audit list by elastic JWK")
}

func (s *auditLogRepositoryStub) ListByOperation(ctx context.Context, tenantID googleUuid.UUID, operation string, offset, limit int) ([]*cryptoutilAppsSmKmsModel.AuditLogEntry, int64, error) {
	if s.listByOperationFn != nil {
		return s.listByOperationFn(ctx, tenantID, operation, offset, limit)
	}

	return nil, 0, errors.New("unexpected audit list by operation")
}

func (s *auditLogRepositoryStub) DeleteOlderThan(ctx context.Context, tenantID googleUuid.UUID, days int) (int64, error) {
	if s.deleteOlderThanFn != nil {
		return s.deleteOlderThanFn(ctx, tenantID, days)
	}

	return 0, nil
}

type auditConfigRepositoryStub struct {
	cryptoutilAppsSmKmsRepository.AuditConfigRepository
	shouldAuditFn func(context.Context, googleUuid.UUID, string) (bool, error)
	getAllFn      func(context.Context, googleUuid.UUID) ([]*cryptoutilAppsSmKmsModel.AuditConfig, error)
	upsertFn      func(context.Context, *cryptoutilAppsSmKmsModel.AuditConfig) error
}

func (s *auditConfigRepositoryStub) ShouldAudit(ctx context.Context, tenantID googleUuid.UUID, operation string) (bool, error) {
	if s.shouldAuditFn != nil {
		return s.shouldAuditFn(ctx, tenantID, operation)
	}

	return true, nil
}

func (s *auditConfigRepositoryStub) GetAllForTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*cryptoutilAppsSmKmsModel.AuditConfig, error) {
	if s.getAllFn != nil {
		return s.getAllFn(ctx, tenantID)
	}

	return nil, nil
}

func (s *auditConfigRepositoryStub) Upsert(ctx context.Context, config *cryptoutilAppsSmKmsModel.AuditConfig) error {
	if s.upsertFn != nil {
		return s.upsertFn(ctx, config)
	}

	return nil
}
