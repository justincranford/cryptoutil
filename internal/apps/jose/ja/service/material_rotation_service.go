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
	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	joseJARepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
)

// MaterialRotationService provides business logic for material JWK rotation.
type MaterialRotationService interface {
	// RotateMaterial creates a new active material and marks previous as inactive.
	RotateMaterial(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*joseJADomain.MaterialJWK, error)

	// RetireMaterial marks a material JWK as retired (no longer usable for signing/encryption).
	RetireMaterial(ctx context.Context, tenantID, elasticJWKID, materialID googleUuid.UUID) error

	// ListMaterials lists all materials for an elastic JWK.
	ListMaterials(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) ([]*joseJADomain.MaterialJWK, error)

	// GetActiveMaterial gets the currently active material for an elastic JWK.
	GetActiveMaterial(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*joseJADomain.MaterialJWK, error)

	// GetMaterialByKID gets a material by its KID (for decryption/verification).
	GetMaterialByKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, kid string) (*joseJADomain.MaterialJWK, error)
}

// materialRotationServiceImpl implements MaterialRotationService.
type materialRotationServiceImpl struct {
	elasticRepo  joseJARepository.ElasticJWKRepository
	materialRepo joseJARepository.MaterialJWKRepository
	jwkGenSvc    *cryptoutilJose.JWKGenService
	barrierSvc   *cryptoutilBarrier.BarrierService
}

// NewMaterialRotationService creates a new MaterialRotationService.
func NewMaterialRotationService(
	elasticRepo joseJARepository.ElasticJWKRepository,
	materialRepo joseJARepository.MaterialJWKRepository,
	jwkGenSvc *cryptoutilJose.JWKGenService,
	barrierSvc *cryptoutilBarrier.BarrierService,
) MaterialRotationService {
	return &materialRotationServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		jwkGenSvc:    jwkGenSvc,
		barrierSvc:   barrierSvc,
	}
}

// RotateMaterial creates a new active material and marks previous as inactive.
func (s *materialRotationServiceImpl) RotateMaterial(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*joseJADomain.MaterialJWK, error) {
	// Verify tenant ownership and get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Check if max materials reached.
	count, err := s.materialRepo.CountMaterials(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to count materials: %w", err)
	}

	if int(count) >= elasticJWK.MaxMaterials {
		return nil, fmt.Errorf("max materials reached (%d), cannot rotate", elasticJWK.MaxMaterials)
	}

	// Create new material.
	newMaterial, err := s.createMaterialJWK(ctx, elasticJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to create new material: %w", err)
	}

	// Rotate material in repository (marks current active as inactive, activates new one).
	if err := s.materialRepo.RotateMaterial(ctx, elasticJWKID, newMaterial); err != nil {
		return nil, fmt.Errorf("failed to rotate material: %w", err)
	}

	// Update elastic JWK material count.
	if err := s.elasticRepo.IncrementMaterialCount(ctx, elasticJWKID); err != nil {
		// Log but don't fail - material was rotated successfully.
		fmt.Printf("warning: failed to update elastic JWK material count: %v\n", err)
	}

	return newMaterial, nil
}

// RetireMaterial marks a material JWK as retired.
func (s *materialRotationServiceImpl) RetireMaterial(ctx context.Context, tenantID, elasticJWKID, materialID googleUuid.UUID) error {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return fmt.Errorf("elastic JWK not found")
	}

	// Verify material belongs to this elastic JWK.
	material, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return fmt.Errorf("failed to get material: %w", err)
	}

	if material.ElasticJWKID != elasticJWKID {
		return fmt.Errorf("material not found for this elastic JWK")
	}

	// Retire material.
	if err := s.materialRepo.RetireMaterial(ctx, materialID); err != nil {
		return fmt.Errorf("failed to retire material: %w", err)
	}

	return nil
}

// ListMaterials lists all materials for an elastic JWK.
func (s *materialRotationServiceImpl) ListMaterials(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) ([]*joseJADomain.MaterialJWK, error) {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// List materials.
	materials, _, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWKID, 0, cryptoutilMagic.JoseJADefaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	return materials, nil
}

// GetActiveMaterial gets the currently active material for an elastic JWK.
func (s *materialRotationServiceImpl) GetActiveMaterial(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*joseJADomain.MaterialJWK, error) {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Get active material.
	material, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active material: %w", err)
	}

	return material, nil
}

// GetMaterialByKID gets a material by its KID.
func (s *materialRotationServiceImpl) GetMaterialByKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, kid string) (*joseJADomain.MaterialJWK, error) {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Get material by KID.
	material, err := s.materialRepo.GetByMaterialKID(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get material by KID: %w", err)
	}

	// Verify material belongs to this elastic JWK.
	if material.ElasticJWKID != elasticJWKID {
		return nil, fmt.Errorf("material not found for this elastic JWK")
	}

	return material, nil
}

// createMaterialJWK generates a new material JWK.
func (s *materialRotationServiceImpl) createMaterialJWK(ctx context.Context, elasticJWK *joseJADomain.ElasticJWK) (*joseJADomain.MaterialJWK, error) {
	// Generate material ID.
	materialID := googleUuid.New()
	materialKID := materialID.String()

	// Convert algorithm to GenerateAlgorithm.
	genAlg := mapToGenerateAlgorithmForRotation(elasticJWK.Algorithm)
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

	// Base64 encode for storage as strings.
	privateJWE := base64.StdEncoding.EncodeToString(privateJWEBytes)
	publicJWE := base64.StdEncoding.EncodeToString(publicJWEBytes)

	// Create material JWK record.
	materialJWK := &joseJADomain.MaterialJWK{
		ID:             materialID,
		ElasticJWKID:   elasticJWK.ID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  privateJWE,
		PublicJWKJWE:   publicJWE,
		BarrierVersion: 1,
		Active:         true,
		CreatedAt:      time.Now(),
	}

	return materialJWK, nil
}

// mapToGenerateAlgorithmForRotation maps algorithm string to OpenAPI GenerateAlgorithm.
// This is duplicated from elastic_jwk_service.go - consider extracting to shared utility.
func mapToGenerateAlgorithmForRotation(algorithm string) *cryptoutilOpenapiModel.GenerateAlgorithm {
	var alg cryptoutilOpenapiModel.GenerateAlgorithm

	switch algorithm {
	case cryptoutilMagic.JoseAlgRS256, cryptoutilMagic.JoseAlgRS384, cryptoutilMagic.JoseAlgRS512, cryptoutilMagic.JoseKeyTypeRSA2048:
		alg = cryptoutilOpenapiModel.RSA2048
	case cryptoutilMagic.JoseAlgPS256, cryptoutilMagic.JoseAlgPS384, cryptoutilMagic.JoseAlgPS512, cryptoutilMagic.JoseKeyTypeRSA3072:
		alg = cryptoutilOpenapiModel.RSA3072
	case cryptoutilMagic.JoseKeyTypeRSA4096:
		alg = cryptoutilOpenapiModel.RSA4096
	case cryptoutilMagic.JoseAlgES256, cryptoutilMagic.JoseKeyTypeECP256:
		alg = cryptoutilOpenapiModel.ECP256
	case cryptoutilMagic.JoseAlgES384, cryptoutilMagic.JoseKeyTypeECP384:
		alg = cryptoutilOpenapiModel.ECP384
	case cryptoutilMagic.JoseAlgES512, cryptoutilMagic.JoseKeyTypeECP521:
		alg = cryptoutilOpenapiModel.ECP521
	case cryptoutilMagic.JoseAlgEdDSA, cryptoutilMagic.JoseKeyTypeOKPEd25519:
		alg = cryptoutilOpenapiModel.OKPEd25519
	case cryptoutilMagic.JoseKeyTypeOct128, cryptoutilMagic.JoseEncA128GCM:
		alg = cryptoutilOpenapiModel.Oct128
	case cryptoutilMagic.JoseKeyTypeOct192, cryptoutilMagic.JoseEncA192GCM:
		alg = cryptoutilOpenapiModel.Oct192
	case cryptoutilMagic.JoseKeyTypeOct256, cryptoutilMagic.JoseEncA256GCM:
		alg = cryptoutilOpenapiModel.Oct256
	case cryptoutilMagic.JoseKeyTypeOct384, cryptoutilMagic.JoseEncA128CBCHS256:
		alg = cryptoutilOpenapiModel.Oct384
	case cryptoutilMagic.JoseKeyTypeOct512, cryptoutilMagic.JoseEncA256CBCHS512:
		alg = cryptoutilOpenapiModel.Oct512
	default:
		return nil
	}

	return &alg
}
