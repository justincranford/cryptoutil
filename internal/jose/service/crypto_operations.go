// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	"fmt"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

// SignRequest contains parameters for signing data with an Elastic JWK.
type SignRequest struct {
	TenantID     googleUuid.UUID
	RealmID      googleUuid.UUID
	ElasticJWKID googleUuid.UUID
	Payload      []byte
}

// SignResponse contains the signed data and metadata.
type SignResponse struct {
	JWSMessage      *joseJws.Message
	JWSMessageBytes []byte
	MaterialKID     string
}

// Sign signs data using the active material JWK for an Elastic JWK.
// The material_kid is embedded in the JWS header for verification.
func (s *ElasticJWKService) Sign(ctx context.Context, req *SignRequest) (*SignResponse, error) {
	// Get the elastic JWK to verify ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, req.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant and realm match (tenant isolation).
	if elasticJWK.TenantID != req.TenantID || elasticJWK.RealmID != req.RealmID {
		return nil, fmt.Errorf("elastic JWK not found in specified tenant/realm")
	}

	// Verify this is a signing key.
	if elasticJWK.USE != cryptoutilSharedMagic.JoseKeyUseSig {
		return nil, fmt.Errorf("elastic JWK is not a signing key (use=%s)", elasticJWK.USE)
	}

	// Get the active material JWK.
	activeMaterial, err := s.materialRepo.GetActiveMaterial(ctx, req.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active material JWK: %w", err)
	}

	// Decrypt the private JWK with barrier service.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(activeMaterial.PrivateJWKJWE))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	// Parse the decrypted JWK.
	privateJWK, err := joseJwk.ParseKey(privateJWKJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Set the kid to the material_kid for tracking.
	if err := privateJWK.Set(joseJwk.KeyIDKey, activeMaterial.MaterialKID); err != nil {
		return nil, fmt.Errorf("failed to set kid on private JWK: %w", err)
	}

	// Sign using the shared crypto utility.
	jwsMessage, jwsMessageBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{privateJWK}, req.Payload)
	if err != nil {
		s.logAuditFailure(ctx, req.TenantID, req.RealmID, AuditOperationSign, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": req.ElasticJWKID.String(),
			"material_kid":   activeMaterial.MaterialKID,
		})

		return nil, fmt.Errorf("failed to sign payload: %w", err)
	}

	// Log successful sign operation.
	s.logAuditSuccess(ctx, req.TenantID, req.RealmID, AuditOperationSign, "elastic_jwk", elasticJWK.KID, map[string]any{
		"elastic_jwk_id": req.ElasticJWKID.String(),
		"material_kid":   activeMaterial.MaterialKID,
	})

	return &SignResponse{
		JWSMessage:      jwsMessage,
		JWSMessageBytes: jwsMessageBytes,
		MaterialKID:     activeMaterial.MaterialKID,
	}, nil
}

// EncryptRequest contains parameters for encrypting data with an Elastic JWK.
type EncryptRequest struct {
	TenantID     googleUuid.UUID
	RealmID      googleUuid.UUID
	ElasticJWKID googleUuid.UUID
	Plaintext    []byte
}

// EncryptResponse contains the encrypted data and metadata.
type EncryptResponse struct {
	JWEMessage      *joseJwe.Message
	JWEMessageBytes []byte
	MaterialKID     string
}

// Encrypt encrypts data using the active material JWK for an Elastic JWK.
// The material_kid is embedded in the JWE header for decryption.
func (s *ElasticJWKService) Encrypt(ctx context.Context, req *EncryptRequest) (*EncryptResponse, error) {
	// Get the elastic JWK to verify ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, req.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant and realm match (tenant isolation).
	if elasticJWK.TenantID != req.TenantID || elasticJWK.RealmID != req.RealmID {
		return nil, fmt.Errorf("elastic JWK not found in specified tenant/realm")
	}

	// Verify this is an encryption key.
	if elasticJWK.USE != cryptoutilSharedMagic.JoseKeyUseEnc {
		return nil, fmt.Errorf("elastic JWK is not an encryption key (use=%s)", elasticJWK.USE)
	}

	// Get the active material JWK.
	activeMaterial, err := s.materialRepo.GetActiveMaterial(ctx, req.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active material JWK: %w", err)
	}

	// Decrypt the public JWK with barrier service.
	// For symmetric keys, public_jwk_jwe contains the same key as private_jwk_jwe.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(activeMaterial.PublicJWKJWE))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse the decrypted JWK.
	publicJWK, err := joseJwk.ParseKey(publicJWKJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Set the kid to the material_kid for tracking.
	if err := publicJWK.Set(joseJwk.KeyIDKey, activeMaterial.MaterialKID); err != nil {
		return nil, fmt.Errorf("failed to set kid on public JWK: %w", err)
	}

	// Encrypt using the shared crypto utility.
	jweMessage, jweMessageBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{publicJWK}, req.Plaintext)
	if err != nil {
		s.logAuditFailure(ctx, req.TenantID, req.RealmID, AuditOperationEncrypt, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": req.ElasticJWKID.String(),
			"material_kid":   activeMaterial.MaterialKID,
		})

		return nil, fmt.Errorf("failed to encrypt plaintext: %w", err)
	}

	// Log successful encrypt operation.
	s.logAuditSuccess(ctx, req.TenantID, req.RealmID, AuditOperationEncrypt, "elastic_jwk", elasticJWK.KID, map[string]any{
		"elastic_jwk_id": req.ElasticJWKID.String(),
		"material_kid":   activeMaterial.MaterialKID,
	})

	return &EncryptResponse{
		JWEMessage:      jweMessage,
		JWEMessageBytes: jweMessageBytes,
		MaterialKID:     activeMaterial.MaterialKID,
	}, nil
}

// VerifyRequest contains parameters for verifying a JWS signature.
type VerifyRequest struct {
	TenantID        googleUuid.UUID
	JWSMessageBytes []byte
}

// VerifyResponse contains the verified payload and metadata.
type VerifyResponse struct {
	Payload     []byte
	MaterialKID string
}

// Verify verifies a JWS signature using the material JWK identified by the kid in the header.
// This supports historical materials (retired_at != NULL still usable for verification).
func (s *ElasticJWKService) Verify(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	// Parse the JWS message to extract the material_kid.
	jwsMessage, err := joseJws.Parse(req.JWSMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message: %w", err)
	}

	// Extract the kid (material_kid) from the JWS header.
	materialKID, _, err := cryptoutilSharedCryptoJose.ExtractKidAlgFromJWSMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kid from JWS header: %w", err)
	}

	// Look up the material JWK by material_kid globally (includes historical materials).
	material, err := s.materialRepo.GetByMaterialKIDGlobal(ctx, materialKID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	// Get the elastic JWK to verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, material.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant matches (tenant isolation).
	if elasticJWK.TenantID != req.TenantID {
		return nil, fmt.Errorf("material JWK not found for tenant")
	}

	// Decrypt the public JWK with barrier service.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(material.PublicJWKJWE))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse the decrypted JWK.
	publicJWK, err := joseJwk.ParseKey(publicJWKJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Set the kid to match the JWS header.
	if err := publicJWK.Set(joseJwk.KeyIDKey, material.MaterialKID); err != nil {
		return nil, fmt.Errorf("failed to set kid on public JWK: %w", err)
	}

	// Verify using the shared crypto utility.
	payload, err := cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{publicJWK}, req.JWSMessageBytes)
	if err != nil {
		s.logAuditFailure(ctx, req.TenantID, elasticJWK.RealmID, AuditOperationVerify, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": elasticJWK.ID.String(),
			"material_kid":   material.MaterialKID,
		})

		return nil, fmt.Errorf("failed to verify JWS signature: %w", err)
	}

	// Log successful verify operation.
	s.logAuditSuccess(ctx, req.TenantID, elasticJWK.RealmID, AuditOperationVerify, "elastic_jwk", elasticJWK.KID, map[string]any{
		"elastic_jwk_id": elasticJWK.ID.String(),
		"material_kid":   material.MaterialKID,
	})

	return &VerifyResponse{
		Payload:     payload,
		MaterialKID: material.MaterialKID,
	}, nil
}

// DecryptRequest contains parameters for decrypting a JWE message.
type DecryptRequest struct {
	TenantID        googleUuid.UUID
	JWEMessageBytes []byte
}

// DecryptResponse contains the decrypted plaintext and metadata.
type DecryptResponse struct {
	Plaintext   []byte
	MaterialKID string
}

// Decrypt decrypts a JWE message using the material JWK identified by the kid in the header.
// This supports historical materials (retired_at != NULL still usable for decryption).
func (s *ElasticJWKService) Decrypt(ctx context.Context, req *DecryptRequest) (*DecryptResponse, error) {
	// Parse the JWE message to extract the material_kid.
	jweMessage, err := joseJwe.Parse(req.JWEMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}

	// Extract the kid (material_kid) from the JWE header.
	materialKID, err := cryptoutilSharedCryptoJose.ExtractKidFromJWEMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to extract kid from JWE header: %w", err)
	}

	// Look up the material JWK by material_kid globally (includes historical materials).
	material, err := s.materialRepo.GetByMaterialKIDGlobal(ctx, materialKID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	// Get the elastic JWK to verify tenant ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, material.ElasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant matches (tenant isolation).
	if elasticJWK.TenantID != req.TenantID {
		return nil, fmt.Errorf("material JWK not found for tenant")
	}

	// Decrypt the private JWK with barrier service.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(material.PrivateJWKJWE))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	// Parse the decrypted JWK.
	privateJWK, err := joseJwk.ParseKey(privateJWKJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Set the kid to match the JWE header.
	if err := privateJWK.Set(joseJwk.KeyIDKey, material.MaterialKID); err != nil {
		return nil, fmt.Errorf("failed to set kid on private JWK: %w", err)
	}

	// Decrypt using the shared crypto utility.
	plaintext, err := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{privateJWK}, req.JWEMessageBytes)
	if err != nil {
		s.logAuditFailure(ctx, req.TenantID, elasticJWK.RealmID, AuditOperationDecrypt, "elastic_jwk", elasticJWK.KID, err, map[string]any{
			"elastic_jwk_id": elasticJWK.ID.String(),
			"material_kid":   material.MaterialKID,
		})

		return nil, fmt.Errorf("failed to decrypt JWE message: %w", err)
	}

	// Log successful decrypt operation.
	s.logAuditSuccess(ctx, req.TenantID, elasticJWK.RealmID, AuditOperationDecrypt, "elastic_jwk", elasticJWK.KID, map[string]any{
		"elastic_jwk_id": elasticJWK.ID.String(),
		"material_kid":   material.MaterialKID,
	})

	return &DecryptResponse{
		Plaintext:   plaintext,
		MaterialKID: material.MaterialKID,
	}, nil
}

// GetDecryptedMaterialJWK returns the decrypted private JWK for a specific material.
// This is a utility function for internal use (e.g., JWKS endpoint needs public keys).
func (s *ElasticJWKService) GetDecryptedMaterialJWK(ctx context.Context, materialJWKID googleUuid.UUID) (privateJWK, publicJWK joseJwk.Key, err error) {
	// Get the material JWK.
	material, err := s.materialRepo.GetByID(ctx, materialJWKID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get material JWK: %w", err)
	}

	// Decrypt the private JWK.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(material.PrivateJWKJWE))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	privateJWK, err = joseJwk.ParseKey(privateJWKJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Decrypt the public JWK.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(material.PublicJWKJWE))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	publicJWK, err = joseJwk.ParseKey(publicJWKJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	return privateJWK, publicJWK, nil
}

// GetDecryptedPublicJWKs returns all decrypted public JWKs for an Elastic JWK.
// This is used by the JWKS endpoint to expose public keys.
func (s *ElasticJWKService) GetDecryptedPublicJWKs(ctx context.Context, tenantID, realmID, elasticJWKID googleUuid.UUID) ([]joseJwk.Key, error) {
	// Get the elastic JWK to verify ownership.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	// Verify tenant and realm match (tenant isolation).
	if elasticJWK.TenantID != tenantID || elasticJWK.RealmID != realmID {
		return nil, fmt.Errorf("elastic JWK not found in specified tenant/realm")
	}

	// For symmetric keys (oct), we should not expose the key via JWKS.
	if elasticJWK.KTY == "oct" {
		return nil, fmt.Errorf("symmetric keys cannot be exposed via JWKS")
	}

	// Get all material JWKs for this elastic JWK.
	// Use a large limit to get all materials (max is 1000 per elastic JWK).
	materials, err := s.materialRepo.ListByElasticJWK(ctx, elasticJWKID, 0, MaxMaterialsPerElasticJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to list material JWKs: %w", err)
	}

	publicJWKs := make([]joseJwk.Key, 0, len(materials))

	for _, material := range materials {
		// Decrypt the public JWK.
		publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(material.PublicJWKJWE))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt public JWK for material %s: %w", material.MaterialKID, err)
		}

		publicJWK, err := joseJwk.ParseKey(publicJWKJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public JWK for material %s: %w", material.MaterialKID, err)
		}

		// Set the kid to the material_kid.
		if err := publicJWK.Set(joseJwk.KeyIDKey, material.MaterialKID); err != nil {
			return nil, fmt.Errorf("failed to set kid on public JWK for material %s: %w", material.MaterialKID, err)
		}

		publicJWKs = append(publicJWKs, publicJWK)
	}

	return publicJWKs, nil
}
