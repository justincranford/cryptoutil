// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"encoding/base64"
	"fmt"

	joseJADomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	jose "github.com/go-jose/go-jose/v4"
	googleUuid "github.com/google/uuid"
)

// JWEService provides JWE encryption and decryption operations.
type JWEService interface {
	// Encrypt encrypts plaintext using the active material key of an elastic JWK.
	Encrypt(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, plaintext []byte) (string, error)

	// Decrypt decrypts a JWE compact serialization.
	Decrypt(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, jweCompact string) ([]byte, error)

	// EncryptWithKID encrypts plaintext using a specific material key.
	EncryptWithKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, materialKID string, plaintext []byte) (string, error)
}

// jweServiceImpl implements JWEService.
type jweServiceImpl struct {
	elasticRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	barrierSvc   *cryptoutilBarrier.BarrierService
}

// NewJWEService creates a new JWEService.
func NewJWEService(
	elasticRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	barrierSvc *cryptoutilBarrier.BarrierService,
) JWEService {
	return &jweServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		barrierSvc:   barrierSvc,
	}
}

// Encrypt encrypts plaintext using the active material key.
func (s *jweServiceImpl) Encrypt(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, plaintext []byte) (string, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return "", fmt.Errorf("elastic JWK not found")
	}

	// Verify key use is for encryption.
	if elasticJWK.Use != joseJADomain.KeyUseEnc {
		return "", fmt.Errorf("elastic JWK is not configured for encryption (use=%s)", elasticJWK.Use)
	}

	// Get active material.
	material, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get active material: %w", err)
	}

	// Decode base64 encoded JWE string.
	publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
	if err != nil {
		return "", fmt.Errorf("failed to decode public JWK JWE: %w", err)
	}

	// Decrypt public JWK.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse public JWK.
	var publicJWK jose.JSONWebKey
	if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
		return "", fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Determine key algorithm and content encryption.
	keyAlg, contentEnc := mapToJWEAlgorithms(elasticJWK.Algorithm)
	if keyAlg == "" {
		return "", fmt.Errorf("unsupported algorithm for JWE: %s", elasticJWK.Algorithm)
	}

	// Create JWE encrypter.
	encrypter, err := jose.NewEncrypter(contentEnc, jose.Recipient{Algorithm: keyAlg, Key: publicJWK}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create encrypter: %w", err)
	}

	// Encrypt plaintext.
	jweObject, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Serialize to compact form.
	compact, err := jweObject.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize JWE: %w", err)
	}

	return compact, nil
}

// Decrypt decrypts a JWE compact serialization.
func (s *jweServiceImpl) Decrypt(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, jweCompact string) ([]byte, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Parse JWE.
	jweObject, err := jose.ParseEncrypted(jweCompact, []jose.KeyAlgorithm{jose.RSA_OAEP, jose.RSA_OAEP_256, jose.ECDH_ES, jose.ECDH_ES_A128KW, jose.ECDH_ES_A192KW, jose.ECDH_ES_A256KW, jose.A128KW, jose.A192KW, jose.A256KW, jose.A128GCMKW, jose.A192GCMKW, jose.A256GCMKW, jose.DIRECT}, []jose.ContentEncryption{jose.A128GCM, jose.A192GCM, jose.A256GCM, jose.A128CBC_HS256, jose.A192CBC_HS384, jose.A256CBC_HS512})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE: %w", err)
	}

	// Try to decrypt with all available materials.
	materials, _, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWKID, 0, cryptoutilMagic.JoseJADefaultListLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list materials: %w", err)
	}

	for _, material := range materials {
		plaintext, err := s.decryptWithMaterial(ctx, jweObject, material)
		if err == nil {
			return plaintext, nil
		}
	}

	return nil, fmt.Errorf("failed to decrypt JWE: no matching key found")
}

// EncryptWithKID encrypts plaintext using a specific material key.
func (s *jweServiceImpl) EncryptWithKID(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, materialKID string, plaintext []byte) (string, error) {
	// Get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant ownership.
	if elasticJWK.TenantID != tenantID {
		return "", fmt.Errorf("elastic JWK not found")
	}

	// Verify key use is for encryption.
	if elasticJWK.Use != joseJADomain.KeyUseEnc {
		return "", fmt.Errorf("elastic JWK is not configured for encryption (use=%s)", elasticJWK.Use)
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

	// Decode base64 encoded JWE string.
	publicJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PublicJWKJWE)
	if err != nil {
		return "", fmt.Errorf("failed to decode public JWK JWE: %w", err)
	}

	// Decrypt public JWK.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse public JWK.
	var publicJWK jose.JSONWebKey
	if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
		return "", fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Determine key algorithm and content encryption.
	keyAlg, contentEnc := mapToJWEAlgorithms(elasticJWK.Algorithm)
	if keyAlg == "" {
		return "", fmt.Errorf("unsupported algorithm for JWE: %s", elasticJWK.Algorithm)
	}

	// Create JWE encrypter.
	encrypter, err := jose.NewEncrypter(contentEnc, jose.Recipient{Algorithm: keyAlg, Key: publicJWK}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create encrypter: %w", err)
	}

	// Encrypt plaintext.
	jweObject, err := encrypter.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Serialize to compact form.
	compact, err := jweObject.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize JWE: %w", err)
	}

	return compact, nil
}

// decryptWithMaterial attempts to decrypt a JWE with a specific material key.
func (s *jweServiceImpl) decryptWithMaterial(ctx context.Context, jweObject *jose.JSONWebEncryption, material *joseJADomain.MaterialJWK) ([]byte, error) {
	// Decode base64 encoded JWE string.
	privateJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PrivateJWKJWE)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private JWK JWE: %w", err)
	}

	// Decrypt private JWK.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, privateJWKEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	// Parse private JWK.
	var privateJWK jose.JSONWebKey
	if err := privateJWK.UnmarshalJSON(privateJWKJSON); err != nil {
		return nil, fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Attempt decryption.
	plaintext, err := jweObject.Decrypt(privateJWK)
	if err != nil {
		return nil, err //nolint:wrapcheck // Expected to fail for non-matching keys.
	}

	return plaintext, nil
}

// mapToJWEAlgorithms maps algorithm string to JWE key algorithm and content encryption.
func mapToJWEAlgorithms(algorithm string) (jose.KeyAlgorithm, jose.ContentEncryption) {
	switch algorithm {
	case cryptoutilMagic.JoseKeyTypeRSA2048, cryptoutilMagic.JoseKeyTypeRSA3072, cryptoutilMagic.JoseKeyTypeRSA4096,
		cryptoutilMagic.JoseAlgRSAOAEP, cryptoutilMagic.JoseAlgRSAOAEP256:
		return jose.RSA_OAEP_256, jose.A256GCM
	case cryptoutilMagic.JoseKeyTypeECP256, cryptoutilMagic.JoseKeyTypeECP384, cryptoutilMagic.JoseKeyTypeECP521,
		cryptoutilMagic.JoseAlgECDHES:
		return jose.ECDH_ES_A256KW, jose.A256GCM
	case "A128KW":
		return jose.A128KW, jose.A128GCM
	case "A192KW":
		return jose.A192KW, jose.A192GCM
	case "A256KW":
		return jose.A256KW, jose.A256GCM
	case "A128GCMKW":
		return jose.A128GCMKW, jose.A128GCM
	case "A192GCMKW":
		return jose.A192GCMKW, jose.A192GCM
	case "A256GCMKW":
		return jose.A256GCMKW, jose.A256GCM
	case cryptoutilMagic.JoseAlgDir, cryptoutilMagic.JoseKeyTypeOct128:
		return jose.DIRECT, jose.A128GCM
	case cryptoutilMagic.JoseKeyTypeOct192:
		return jose.DIRECT, jose.A192GCM
	case cryptoutilMagic.JoseKeyTypeOct256:
		return jose.DIRECT, jose.A256GCM
	default:
		return "", ""
	}
}
