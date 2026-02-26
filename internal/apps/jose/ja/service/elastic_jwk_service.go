// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
)

// ElasticJWKService provides business logic for Elastic JWK operations.
type ElasticJWKService interface {
	// CreateElasticJWK creates a new elastic JWK container and initial material.
	CreateElasticJWK(ctx context.Context, tenantID googleUuid.UUID, algorithm, use string, maxMaterials int) (*cryptoutilAppsJoseJaDomain.ElasticJWK, *cryptoutilAppsJoseJaDomain.MaterialJWK, error)

	// GetElasticJWK retrieves an elastic JWK by ID.
	GetElasticJWK(ctx context.Context, tenantID, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.ElasticJWK, error)

	// ListElasticJWKs lists elastic JWKs for a tenant with pagination.
	ListElasticJWKs(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.ElasticJWK, int64, error)

	// DeleteElasticJWK deletes an elastic JWK and all its materials.
	DeleteElasticJWK(ctx context.Context, tenantID, id googleUuid.UUID) error
}

// elasticJWKServiceImpl implements ElasticJWKService.
type elasticJWKServiceImpl struct {
	elasticRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	jwkGenSvc    *cryptoutilSharedCryptoJose.JWKGenService
	barrierSvc   *cryptoutilAppsTemplateServiceServerBarrier.Service
}

// NewElasticJWKService creates a new ElasticJWKService.
func NewElasticJWKService(
	elasticRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	jwkGenSvc *cryptoutilSharedCryptoJose.JWKGenService,
	barrierSvc *cryptoutilAppsTemplateServiceServerBarrier.Service,
) ElasticJWKService {
	return &elasticJWKServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		jwkGenSvc:    jwkGenSvc,
		barrierSvc:   barrierSvc,
	}
}

// CreateElasticJWK creates a new elastic JWK container with initial material key.
func (s *elasticJWKServiceImpl) CreateElasticJWK(ctx context.Context, tenantID googleUuid.UUID, algorithm, use string, maxMaterials int) (*cryptoutilAppsJoseJaDomain.ElasticJWK, *cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	// Validate algorithm and derive key type.
	keyType := mapAlgorithmToKeyType(algorithm)
	if keyType == "" {
		return nil, nil, fmt.Errorf("invalid algorithm: %s", algorithm)
	}

	// Validate use.
	if use != cryptoutilAppsJoseJaDomain.KeyUseSig && use != cryptoutilAppsJoseJaDomain.KeyUseEnc {
		return nil, nil, fmt.Errorf("invalid key use: %s (must be 'sig' or 'enc')", use)
	}

	// Set default max materials.
	if maxMaterials <= 0 {
		maxMaterials = cryptoutilSharedMagic.JoseJADefaultMaxMaterials
	}

	// Generate elastic JWK ID and KID.
	elasticID := googleUuid.New()
	elasticKID := elasticID.String()

	// Create elastic JWK record.
	elasticJWK := &cryptoutilAppsJoseJaDomain.ElasticJWK{
		ID:                   elasticID,
		TenantID:             tenantID,
		KID:                  elasticKID,
		KeyType:              keyType,
		Algorithm:            algorithm,
		Use:                  use,
		MaxMaterials:         maxMaterials,
		CurrentMaterialCount: 1,
		CreatedAt:            time.Now().UTC(),
	}

	// Store elastic JWK.
	if err := s.elasticRepo.Create(ctx, elasticJWK); err != nil {
		return nil, nil, fmt.Errorf("failed to create elastic JWK: %w", err)
	}

	// Create initial material.
	material, err := s.createMaterialJWK(ctx, elasticJWK, true)
	if err != nil {
		// Clean up elastic JWK on failure.
		_ = s.elasticRepo.Delete(ctx, elasticID)

		return nil, nil, fmt.Errorf("failed to create initial material: %w", err)
	}

	return elasticJWK, material, nil
}

// GetElasticJWK retrieves an elastic JWK by ID.
func (s *elasticJWKServiceImpl) GetElasticJWK(ctx context.Context, tenantID, id googleUuid.UUID) (*cryptoutilAppsJoseJaDomain.ElasticJWK, error) {
	elasticJWK, err := s.elasticRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	return elasticJWK, nil
}

// ListElasticJWKs lists elastic JWKs for a tenant with pagination.
func (s *elasticJWKServiceImpl) ListElasticJWKs(ctx context.Context, tenantID googleUuid.UUID, offset, limit int) ([]*cryptoutilAppsJoseJaDomain.ElasticJWK, int64, error) {
	elasticJWKs, total, err := s.elasticRepo.List(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list elastic JWKs: %w", err)
	}

	return elasticJWKs, total, nil
}

// DeleteElasticJWK deletes an elastic JWK and all its materials.
func (s *elasticJWKServiceImpl) DeleteElasticJWK(ctx context.Context, tenantID, id googleUuid.UUID) error {
	// Verify ownership first.
	elasticJWK, err := s.GetElasticJWK(ctx, tenantID, id)
	if err != nil {
		return err
	}

	// Delete all materials first.
	materials, _, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWK.ID, 0, cryptoutilSharedMagic.JoseJADefaultListLimit)
	if err != nil {
		return fmt.Errorf("failed to list materials for deletion: %w", err)
	}

	for _, material := range materials {
		if err := s.materialRepo.Delete(ctx, material.ID); err != nil {
			return fmt.Errorf("failed to delete material %s: %w", material.ID, err)
		}
	}

	// Delete elastic JWK.
	if err := s.elasticRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete elastic JWK: %w", err)
	}

	return nil
}

// createMaterialJWK generates and stores a new material key for an elastic JWK.
func (s *elasticJWKServiceImpl) createMaterialJWK(ctx context.Context, elasticJWK *cryptoutilAppsJoseJaDomain.ElasticJWK, active bool) (*cryptoutilAppsJoseJaDomain.MaterialJWK, error) {
	// Generate material ID.
	materialID := googleUuid.New()
	materialKID := materialID.String()

	// Convert algorithm to GenerateAlgorithm.
	genAlg := mapToGenerateAlgorithm(elasticJWK.Algorithm)
	if genAlg == nil {
		return nil, fmt.Errorf("unsupported algorithm for key generation: %s", elasticJWK.Algorithm)
	}

	// Generate JWK using JWKGenService.
	kid, privateJWK, publicJWK, privateJWKBytes, publicJWKBytes, err := s.jwkGenSvc.GenerateJWK(genAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWK: %w", err)
	}

	// Use generated KID if not already set.
	if kid != nil {
		materialKID = kid.String()
	}

	// Set KID on JWKs.
	// Note: For symmetric keys (oct), publicJWK is nil - symmetric keys have no separate public key.
	if err := privateJWK.Set("kid", materialKID); err != nil {
		return nil, fmt.Errorf("failed to set private JWK kid: %w", err)
	}

	// For asymmetric keys, set KID on public JWK as well.
	if publicJWK != nil {
		if err := publicJWK.Set("kid", materialKID); err != nil {
			return nil, fmt.Errorf("failed to set public JWK kid: %w", err)
		}
	}

	// Encrypt private and public JWKs with barrier.
	// For symmetric keys (oct), use private JWK for both since there's no separate public key.
	privateJWEBytes, err := s.barrierSvc.EncryptContentWithContext(ctx, privateJWKBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private JWK: %w", err)
	}

	// For symmetric keys, publicJWKBytes is nil - use privateJWKBytes for both fields.
	publicBytesToEncrypt := publicJWKBytes
	if publicBytesToEncrypt == nil {
		publicBytesToEncrypt = privateJWKBytes
	}

	publicJWEBytes, err := s.barrierSvc.EncryptContentWithContext(ctx, publicBytesToEncrypt)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt public JWK: %w", err)
	}

	// Convert encrypted bytes to base64 strings for storage.
	privateJWE := base64.StdEncoding.EncodeToString(privateJWEBytes)
	publicJWE := base64.StdEncoding.EncodeToString(publicJWEBytes)

	// Create material JWK record.
	// Note: MaterialJWK.PrivateJWKJWE and PublicJWKJWE are strings.
	materialJWK := &cryptoutilAppsJoseJaDomain.MaterialJWK{
		ID:             materialID,
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  privateJWE,
		PublicJWKJWE:   publicJWE,
		Active:         active,
		CreatedAt:      time.Now().UTC(),
		BarrierVersion: 1, // Initial barrier version.
	}

	// Store material JWK.
	if err := s.materialRepo.Create(ctx, materialJWK); err != nil {
		return nil, fmt.Errorf("failed to create material JWK: %w", err)
	}

	return materialJWK, nil
}

// mapAlgorithmToKeyType maps algorithm string to key type.
func mapAlgorithmToKeyType(algorithm string) string {
	switch algorithm {
	case cryptoutilSharedMagic.JoseAlgRS256, cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS512,
		cryptoutilSharedMagic.JoseAlgPS256, cryptoutilSharedMagic.JoseAlgPS384, cryptoutilSharedMagic.JoseAlgPS512,
		cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilSharedMagic.JoseKeyTypeRSA3072, cryptoutilSharedMagic.JoseKeyTypeRSA4096:
		return cryptoutilAppsJoseJaDomain.KeyTypeRSA
	case cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseAlgES512,
		cryptoutilSharedMagic.JoseKeyTypeECP256, cryptoutilSharedMagic.JoseKeyTypeECP384, cryptoutilSharedMagic.JoseKeyTypeECP521:
		return cryptoutilAppsJoseJaDomain.KeyTypeEC
	case cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilSharedMagic.JoseKeyTypeOKPEd25519:
		return cryptoutilAppsJoseJaDomain.KeyTypeOKP
	case cryptoutilSharedMagic.JoseEncA128GCM, cryptoutilSharedMagic.JoseEncA192GCM, cryptoutilSharedMagic.JoseEncA256GCM,
		cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilSharedMagic.JoseEncA192CBCHS384, cryptoutilSharedMagic.JoseEncA256CBCHS512,
		cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilSharedMagic.JoseKeyTypeOct256,
		cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilSharedMagic.JoseKeyTypeOct512:
		return cryptoutilAppsJoseJaDomain.KeyTypeOct
	default:
		return ""
	}
}

// mapToGenerateAlgorithm maps algorithm string to OpenAPI GenerateAlgorithm.
func mapToGenerateAlgorithm(algorithm string) *cryptoutilOpenapiModel.GenerateAlgorithm {
	var alg cryptoutilOpenapiModel.GenerateAlgorithm

	switch algorithm {
	case cryptoutilSharedMagic.JoseAlgRS256, cryptoutilSharedMagic.JoseAlgRS384, cryptoutilSharedMagic.JoseAlgRS512, cryptoutilSharedMagic.JoseKeyTypeRSA2048:
		alg = cryptoutilOpenapiModel.RSA2048
	case cryptoutilSharedMagic.JoseAlgPS256, cryptoutilSharedMagic.JoseAlgPS384, cryptoutilSharedMagic.JoseAlgPS512, cryptoutilSharedMagic.JoseKeyTypeRSA3072:
		alg = cryptoutilOpenapiModel.RSA3072
	case cryptoutilSharedMagic.JoseKeyTypeRSA4096:
		alg = cryptoutilOpenapiModel.RSA4096
	case cryptoutilSharedMagic.JoseAlgES256, cryptoutilSharedMagic.JoseKeyTypeECP256:
		alg = cryptoutilOpenapiModel.ECP256
	case cryptoutilSharedMagic.JoseAlgES384, cryptoutilSharedMagic.JoseKeyTypeECP384:
		alg = cryptoutilOpenapiModel.ECP384
	case cryptoutilSharedMagic.JoseAlgES512, cryptoutilSharedMagic.JoseKeyTypeECP521:
		alg = cryptoutilOpenapiModel.ECP521
	case cryptoutilSharedMagic.JoseAlgEdDSA, cryptoutilSharedMagic.JoseKeyTypeOKPEd25519:
		alg = cryptoutilOpenapiModel.OKPEd25519
	case cryptoutilSharedMagic.JoseKeyTypeOct128, cryptoutilSharedMagic.JoseEncA128GCM:
		alg = cryptoutilOpenapiModel.Oct128
	case cryptoutilSharedMagic.JoseKeyTypeOct192, cryptoutilSharedMagic.JoseEncA192GCM:
		alg = cryptoutilOpenapiModel.Oct192
	case cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilSharedMagic.JoseEncA256GCM:
		alg = cryptoutilOpenapiModel.Oct256
	case cryptoutilSharedMagic.JoseKeyTypeOct384, cryptoutilSharedMagic.JoseEncA128CBCHS256, cryptoutilSharedMagic.JoseEncA192CBCHS384:
		alg = cryptoutilOpenapiModel.Oct384
	case cryptoutilSharedMagic.JoseKeyTypeOct512, cryptoutilSharedMagic.JoseEncA256CBCHS512:
		alg = cryptoutilOpenapiModel.Oct512
	default:
		return nil
	}

	return &alg
}
