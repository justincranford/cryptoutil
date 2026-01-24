// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"encoding/base64"
	"fmt"

	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	googleUuid "github.com/google/uuid"
)

// JWKSService provides business logic for JWKS (JSON Web Key Set) operations.
type JWKSService interface {
	// GetJWKS returns the JWKS (public keys only) for a tenant.
	GetJWKS(ctx context.Context, tenantID googleUuid.UUID) (*jose.JSONWebKeySet, error)

	// GetJWKSForElasticKey returns the JWKS for a specific elastic JWK.
	GetJWKSForElasticKey(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*jose.JSONWebKeySet, error)

	// GetPublicJWK returns a single public JWK by KID.
	GetPublicJWK(ctx context.Context, tenantID googleUuid.UUID, kid string) (*jose.JSONWebKey, error)
}

// jwksServiceImpl implements JWKSService.
type jwksServiceImpl struct {
	elasticRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	barrierSvc   *cryptoutilBarrier.BarrierService
}

// NewJWKSService creates a new JWKSService.
func NewJWKSService(
	elasticRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	barrierSvc *cryptoutilBarrier.BarrierService,
) JWKSService {
	return &jwksServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		barrierSvc:   barrierSvc,
	}
}

// GetJWKS returns the JWKS (public keys only) for a tenant.
func (s *jwksServiceImpl) GetJWKS(ctx context.Context, tenantID googleUuid.UUID) (*jose.JSONWebKeySet, error) {
	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{},
	}

	// List all elastic JWKs for tenant.
	elasticJWKs, _, err := s.elasticRepo.List(ctx, tenantID, 0, cryptoutilMagic.JoseJADefaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list elastic JWKs: %w", err)
	}

	// For each elastic JWK, get active material's public key.
	for _, elasticJWK := range elasticJWKs {
		material, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWK.ID)
		if err != nil {
			// Skip if no active material.
			continue
		}

		// Decode base64 encoded JWE string.
		publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
		if err != nil {
			continue // Skip on decode failure.
		}

		// Decrypt public JWK using barrier service.
		publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
		if err != nil {
			continue // Skip on decryption failure.
		}

		// Parse public JWK.
		var publicJWK jose.JSONWebKey
		if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
			continue // Skip on parse failure.
		}

		// Ensure the public key is set correctly.
		publicJWK.KeyID = material.MaterialKID
		publicJWK.Use = elasticJWK.Use
		publicJWK.Algorithm = elasticJWK.Algorithm

		jwks.Keys = append(jwks.Keys, publicJWK)
	}

	return jwks, nil
}

// GetJWKSForElasticKey returns the JWKS for a specific elastic JWK.
func (s *jwksServiceImpl) GetJWKSForElasticKey(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID) (*jose.JSONWebKeySet, error) {
	// Verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	jwks := &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{},
	}

	// List all materials for this elastic JWK.
	materials, _, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWKID, 0, cryptoutilMagic.JoseJADefaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	// For each material, add public key to JWKS.
	for _, material := range materials {
		// Skip retired materials.
		if material.RetiredAt != nil {
			continue
		}

		// Decode base64 encoded JWE string.
		publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
		if err != nil {
			continue // Skip on decode failure.
		}

		// Decrypt public JWK using barrier service.
		publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
		if err != nil {
			continue // Skip on decryption failure.
		}

		// Parse public JWK.
		var publicJWK jose.JSONWebKey
		if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
			continue // Skip on parse failure.
		}

		// Set metadata.
		publicJWK.KeyID = material.MaterialKID
		publicJWK.Use = elasticJWK.Use
		publicJWK.Algorithm = elasticJWK.Algorithm

		jwks.Keys = append(jwks.Keys, publicJWK)
	}

	return jwks, nil
}

// GetPublicJWK returns a single public JWK by KID.
func (s *jwksServiceImpl) GetPublicJWK(ctx context.Context, tenantID googleUuid.UUID, kid string) (*jose.JSONWebKey, error) {
	// Get material by KID directly.
	material, err := s.materialRepo.GetByMaterialKID(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("JWK with KID %s not found: %w", kid, err)
	}

	// Get the elastic JWK to verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, material.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("JWK with KID %s not found", kid)
	}

	// Decode base64 encoded JWE string.
	publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public JWK JWE: %w", err)
	}

	// Decrypt public JWK using barrier service.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse public JWK.
	var publicJWK jose.JSONWebKey
	if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
		return nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Set metadata.
	publicJWK.KeyID = material.MaterialKID
	publicJWK.Use = elasticJWK.Use
	publicJWK.Algorithm = elasticJWK.Algorithm

	return &publicJWK, nil
}
