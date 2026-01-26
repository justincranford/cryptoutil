// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// MaxMaterialsPerElasticJWK is the maximum number of material JWKs per elastic JWK.
const MaxMaterialsPerElasticJWK = 1000

// MaterialRotationResponse holds the result of a material rotation.
type MaterialRotationResponse struct {
	MaterialJWK *cryptoutilJoseDomain.MaterialJWK
	PublicJWK   joseJwk.Key
}

// RotateMaterial rotates the active material JWK for an elastic JWK.
// It enforces the 1000 material limit and returns an error if the limit is reached.
func (s *ElasticJWKService) RotateMaterial(ctx context.Context, tenantID, realmID, elasticJWKID googleUuid.UUID) (*MaterialRotationResponse, error) {
	// Get the elastic JWK to verify ownership and get algorithm info.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant and realm match (tenant isolation).
	if elasticJWK.TenantID != tenantID || elasticJWK.RealmID != realmID {
		return nil, fmt.Errorf("elastic JWK not found in specified tenant/realm")
	}

	// Check material count limit (CRITICAL: enforce 1000 limit).
	count, err := s.materialRepo.CountMaterials(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to count materials: %w", err)
	}

	if count >= MaxMaterialsPerElasticJWK {
		return nil, fmt.Errorf("elastic JWK %s at max %d materials, rotation blocked", elasticJWKID, MaxMaterialsPerElasticJWK)
	}

	// Generate new material JWK.
	privateJWK, publicJWK, err := s.generateJWK(elasticJWK.KTY, elasticJWK.ALG, elasticJWK.USE)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new JWK: %w", err)
	}

	// Encrypt the private key with barrier.
	privateJWKJSON, err := json.Marshal(privateJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private JWK: %w", err)
	}

	privateJWKJWE, err := s.barrierSvc.EncryptContentWithContext(ctx, privateJWKJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private JWK: %w", err)
	}

	// Handle public key (for symmetric keys, it's the same as private).
	var publicJWKJWE []byte

	if publicJWK != nil {
		publicJWKJSON, err := json.Marshal(publicJWK)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public JWK: %w", err)
		}

		publicJWKJWE, err = s.barrierSvc.EncryptContentWithContext(ctx, publicJWKJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt public JWK: %w", err)
		}
	} else {
		publicJWKJWE = privateJWKJWE
		publicJWK = privateJWK
	}

	// Create new material JWK record.
	materialID := googleUuid.New()
	materialKID := googleUuid.New().String()

	newMaterial := &cryptoutilJoseDomain.MaterialJWK{
		ID:             materialID,
		ElasticJWKID:   elasticJWKID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  string(privateJWKJWE),
		PublicJWKJWE:   string(publicJWKJWE),
		Active:         true,
		CreatedAt:      time.Now().UTC().UnixMilli(),
		RetiredAt:      nil,
		BarrierVersion: 1,
	}

	// Rotate atomically (retire old, insert new).
	if err := s.materialRepo.RotateMaterial(ctx, elasticJWKID, newMaterial); err != nil {
		s.logAuditFailure(ctx, tenantID, realmID, AuditOperationRotate, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": elasticJWKID.String(),
		})

		return nil, fmt.Errorf("failed to rotate material: %w", err)
	}

	// Increment material count in elastic JWK.
	if err := s.elasticRepo.IncrementMaterialCount(ctx, elasticJWKID); err != nil {
		s.logAuditFailure(ctx, tenantID, realmID, AuditOperationRotate, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": elasticJWKID.String(),
		})

		return nil, fmt.Errorf("failed to increment material count: %w", err)
	}

	// Log successful rotation.
	s.logAuditSuccess(ctx, tenantID, realmID, AuditOperationRotate, "elastic_jwk", elasticJWK.KID, map[string]any{
		"elastic_jwk_id": elasticJWKID.String(),
		"material_kid":   materialKID,
		"material_count": count + 1,
	})

	return &MaterialRotationResponse{
		MaterialJWK: newMaterial,
		PublicJWK:   publicJWK,
	}, nil
}

// CanRotate checks if the elastic JWK can accept another material rotation.
func (s *ElasticJWKService) CanRotate(ctx context.Context, elasticJWKID googleUuid.UUID) (bool, int64, error) {
	count, err := s.materialRepo.CountMaterials(ctx, elasticJWKID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to count materials: %w", err)
	}

	return count < MaxMaterialsPerElasticJWK, count, nil
}

// GetMaterialCount returns the current material count for an elastic JWK.
func (s *ElasticJWKService) GetMaterialCount(ctx context.Context, elasticJWKID googleUuid.UUID) (int64, error) {
	count, err := s.materialRepo.CountMaterials(ctx, elasticJWKID)
	if err != nil {
		return 0, fmt.Errorf("failed to count materials: %w", err)
	}

	return count, nil
}
