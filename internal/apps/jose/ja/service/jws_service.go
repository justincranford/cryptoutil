// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"encoding/base64"
	"fmt"

	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	joseJARepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	googleUuid "github.com/google/uuid"
)

// JWSService provides JWS signing and verification operations.
type JWSService interface {
	// Sign signs payload using the active material key of an elastic JWK.
	Sign(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, payload []byte) (string, error)

	// Verify verifies a JWS compact serialization.
	Verify(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, jwsCompact string) ([]byte, error)

	// SignWithKID signs payload using a specific material key.
	SignWithKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, materialKID string, payload []byte) (string, error)
}

// jwsServiceImpl implements JWSService.
type jwsServiceImpl struct {
	elasticRepo  joseJARepository.ElasticJWKRepository
	materialRepo joseJARepository.MaterialJWKRepository
	barrierSvc   *cryptoutilBarrier.BarrierService
}

// NewJWSService creates a new JWSService.
func NewJWSService(
	elasticRepo joseJARepository.ElasticJWKRepository,
	materialRepo joseJARepository.MaterialJWKRepository,
	barrierSvc *cryptoutilBarrier.BarrierService,
) JWSService {
	return &jwsServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		barrierSvc:   barrierSvc,
	}
}

// Sign signs payload using the active material key.
func (s *jwsServiceImpl) Sign(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, payload []byte) (string, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return "", fmt.Errorf("elastic JWK not found")
	}

	// Verify key use is for signing.
	if elasticJWK.Use != joseJADomain.KeyUseSig {
		return "", fmt.Errorf("elastic JWK is not configured for signing (use=%s)", elasticJWK.Use)
	}

	// Get active material.
	material, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get active material: %w", err)
	}

	return s.signWithMaterial(ctx, elasticJWK, material, payload)
}

// Verify verifies a JWS compact serialization.
func (s *jwsServiceImpl) Verify(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, jwsCompact string) ([]byte, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Parse JWS to determine algorithm.
	algorithms := []jose.SignatureAlgorithm{
		jose.RS256, jose.RS384, jose.RS512,
		jose.PS256, jose.PS384, jose.PS512,
		jose.ES256, jose.ES384, jose.ES512,
		jose.EdDSA,
		jose.HS256, jose.HS384, jose.HS512,
	}

	jwsObject, err := jose.ParseSigned(jwsCompact, algorithms)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS: %w", err)
	}

	// Try to verify with all available materials.
	materials, _, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWKID, 0, cryptoutilMagic.JoseJADefaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	for _, material := range materials {
		verifiedPayload, verifyErr := s.verifyWithMaterial(ctx, jwsObject, material)
		if verifyErr == nil {
			return verifiedPayload, nil
		}
	}

	return nil, fmt.Errorf("failed to verify JWS: no matching key found")
}

// SignWithKID signs payload using a specific material key.
func (s *jwsServiceImpl) SignWithKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, materialKID string, payload []byte) (string, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return "", fmt.Errorf("elastic JWK not found")
	}

	// Verify key use is for signing.
	if elasticJWK.Use != joseJADomain.KeyUseSig {
		return "", fmt.Errorf("elastic JWK is not configured for signing (use=%s)", elasticJWK.Use)
	}

	// Find material by KID.
	material, err := s.materialRepo.GetByMaterialKID(ctx, materialKID)
	if err != nil {
		return "", fmt.Errorf("failed to get material by KID: %w", err)
	}

	// Verify material belongs to elastic JWK.
	if material.ElasticJWKID != elasticJWKID {
		return "", fmt.Errorf("material key does not belong to elastic JWK")
	}

	return s.signWithMaterial(ctx, elasticJWK, material, payload)
}

// signWithMaterial signs payload using a specific material key.
func (s *jwsServiceImpl) signWithMaterial(ctx context.Context, elasticJWK *joseJADomain.ElasticJWK, material *joseJADomain.MaterialJWK, payload []byte) (string, error) {
	// Decode base64 encoded JWE string.
	privateJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PrivateJWKJWE)
	if err != nil {
		return "", fmt.Errorf("failed to decode private JWK JWE: %w", err)
	}

	// Decrypt private JWK.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, privateJWKEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	// Parse private JWK.
	var privateJWK jose.JSONWebKey
	if err := privateJWK.UnmarshalJSON(privateJWKJSON); err != nil {
		return "", fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Determine signing algorithm.
	sigAlg := mapToSignatureAlgorithm(elasticJWK.Algorithm)
	if sigAlg == "" {
		return "", fmt.Errorf("unsupported algorithm for JWS: %s", elasticJWK.Algorithm)
	}

	// Create signer.
	signerOpts := &jose.SignerOptions{}
	signerOpts.WithHeader(jose.HeaderKey("kid"), material.MaterialKID)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: sigAlg, Key: privateJWK}, signerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	// Sign payload.
	jwsObject, err := signer.Sign(payload)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Serialize to compact form.
	compact, err := jwsObject.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize JWS: %w", err)
	}

	return compact, nil
}

// verifyWithMaterial verifies a JWS with a specific material key.
func (s *jwsServiceImpl) verifyWithMaterial(ctx context.Context, jwsObject *jose.JSONWebSignature, material *joseJADomain.MaterialJWK) ([]byte, error) {
	// Decode base64 encoded JWE string.
	publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public JWK JWE: %w", err)
	}

	// Decrypt public JWK.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse public JWK.
	var publicJWK jose.JSONWebKey
	if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
		return nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Attempt verification.
	verifiedPayload, err := jwsObject.Verify(publicJWK)
	if err != nil {
		return nil, err //nolint:wrapcheck // Expected to fail for non-matching keys.
	}

	return verifiedPayload, nil
}

// mapToSignatureAlgorithm maps algorithm string to JWS signature algorithm.
func mapToSignatureAlgorithm(algorithm string) jose.SignatureAlgorithm {
	switch algorithm {
	case "RS256", "RSA/2048":
		return jose.RS256
	case "RS384", "RSA/3072":
		return jose.RS384
	case "RS512", "RSA/4096":
		return jose.RS512
	case "PS256":
		return jose.PS256
	case "PS384":
		return jose.PS384
	case "PS512":
		return jose.PS512
	case "ES256", "EC/P256":
		return jose.ES256
	case "ES384", "EC/P384":
		return jose.ES384
	case "ES512", "EC/P521":
		return jose.ES512
	case "EdDSA", "OKP/Ed25519":
		return jose.EdDSA
	case "HS256", "oct/256":
		return jose.HS256
	case "HS384", "oct/384":
		return jose.HS384
	case "HS512", "oct/512":
		return jose.HS512
	default:
		return ""
	}
}
