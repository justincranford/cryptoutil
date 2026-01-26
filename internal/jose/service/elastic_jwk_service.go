// Copyright (c) 2025 Justin Cranford
//
//

// Package service provides JOSE-JA business logic services.
package service

import (
	"context"
	json "encoding/json"
	"fmt"
	"time"

	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// ElasticJWKService manages Elastic JWKs with multi-tenancy support.
type ElasticJWKService struct {
	elasticRepo  cryptoutilJoseRepository.ElasticJWKRepository
	materialRepo cryptoutilJoseRepository.MaterialJWKRepository
	jwkGenSvc    *cryptoutilSharedCryptoJose.JWKGenService
	barrierSvc   *cryptoutilAppsTemplateServiceServerBarrier.Service
	auditLogSvc  *AuditLogService // Optional: nil to disable audit logging.
}

// NewElasticJWKService creates a new ElasticJWKService.
func NewElasticJWKService(
	elasticRepo cryptoutilJoseRepository.ElasticJWKRepository,
	materialRepo cryptoutilJoseRepository.MaterialJWKRepository,
	jwkGenSvc *cryptoutilSharedCryptoJose.JWKGenService,
	barrierSvc *cryptoutilAppsTemplateServiceServerBarrier.Service,
) *ElasticJWKService {
	return &ElasticJWKService{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		jwkGenSvc:    jwkGenSvc,
		barrierSvc:   barrierSvc,
	}
}

// WithAuditLogging adds audit logging to the service.
func (s *ElasticJWKService) WithAuditLogging(auditLogSvc *AuditLogService) *ElasticJWKService {
	s.auditLogSvc = auditLogSvc

	return s
}

// CreateElasticJWKRequest contains parameters for creating an Elastic JWK.
type CreateElasticJWKRequest struct {
	TenantID     googleUuid.UUID
	RealmID      googleUuid.UUID
	KID          string // User-specified KID, or empty for auto-generated.
	KTY          string // RSA, EC, OKP, oct.
	ALG          string // RS256, ES256, A256GCM, etc.
	USE          string // sig, enc.
	MaxMaterials int    // Max materials per elastic JWK (default 1000).
}

// CreateElasticJWKResponse contains the created Elastic JWK and its first material.
type CreateElasticJWKResponse struct {
	ElasticJWK  *cryptoutilJoseDomain.ElasticJWK
	MaterialJWK *cryptoutilJoseDomain.MaterialJWK
	PublicJWK   joseJwk.Key
}

// CreateElasticJWK creates a new Elastic JWK with its first Material JWK.
func (s *ElasticJWKService) CreateElasticJWK(ctx context.Context, req *CreateElasticJWKRequest) (*CreateElasticJWKResponse, error) {
	// Generate IDs.
	elasticID := googleUuid.New()
	materialID := googleUuid.New()

	// Use provided KID or generate one.
	kid := req.KID
	if kid == "" {
		kid = googleUuid.New().String()
	}

	materialKID := googleUuid.New().String()

	// Set default max materials.
	maxMaterials := req.MaxMaterials
	if maxMaterials <= 0 {
		maxMaterials = 1000
	}

	// Generate the JWK based on key type and use.
	privateJWK, publicJWK, err := s.generateJWK(req.KTY, req.ALG, req.USE)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWK: %w", err)
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

	// For symmetric keys (oct), public key is the same as private key.
	// For asymmetric keys, publicJWK is the actual public key.
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
		// For symmetric keys, use the private key as the "public" key.
		publicJWKJWE = privateJWKJWE
		publicJWK = privateJWK
	}

	// Create the Elastic JWK record.
	elasticJWK := &cryptoutilJoseDomain.ElasticJWK{
		ID:                   elasticID,
		TenantID:             req.TenantID,
		RealmID:              req.RealmID,
		KID:                  kid,
		KTY:                  req.KTY,
		ALG:                  req.ALG,
		USE:                  req.USE,
		MaxMaterials:         maxMaterials,
		CurrentMaterialCount: 1, // Starting with first material.
		CreatedAt:            time.Now().UTC().UnixMilli(),
	}

	// Create the first Material JWK record.
	// BarrierVersion is tracked by the barrier service internally.
	materialJWK := &cryptoutilJoseDomain.MaterialJWK{
		ID:             materialID,
		ElasticJWKID:   elasticID,
		MaterialKID:    materialKID,
		PrivateJWKJWE:  string(privateJWKJWE),
		PublicJWKJWE:   string(publicJWKJWE),
		Active:         true,
		CreatedAt:      time.Now().UTC().UnixMilli(),
		RetiredAt:      nil,
		BarrierVersion: 1, // Version is embedded in the JWE.
	}

	// Store the Elastic JWK.
	if err := s.elasticRepo.Create(ctx, elasticJWK); err != nil {
		s.logAuditFailure(ctx, req.TenantID, req.RealmID, AuditOperationKeyGen, "elastic_jwk", kid, err, map[string]any{
			"kty": req.KTY,
			"alg": req.ALG,
			"use": req.USE,
		})

		return nil, fmt.Errorf("failed to create elastic JWK: %w", err)
	}

	// Store the Material JWK.
	if err := s.materialRepo.Create(ctx, materialJWK); err != nil {
		s.logAuditFailure(ctx, req.TenantID, req.RealmID, AuditOperationKeyGen, "elastic_jwk", kid, err, map[string]any{
			"kty": req.KTY,
			"alg": req.ALG,
			"use": req.USE,
		})

		return nil, fmt.Errorf("failed to create material JWK: %w", err)
	}

	// Log successful key generation.
	s.logAuditSuccess(ctx, req.TenantID, req.RealmID, AuditOperationKeyGen, "elastic_jwk", kid, map[string]any{
		"kty":          req.KTY,
		"alg":          req.ALG,
		"use":          req.USE,
		"material_kid": materialKID,
	})

	return &CreateElasticJWKResponse{
		ElasticJWK:  elasticJWK,
		MaterialJWK: materialJWK,
		PublicJWK:   publicJWK,
	}, nil
}

// GetElasticJWK retrieves an Elastic JWK by tenant, realm, and KID.
func (s *ElasticJWKService) GetElasticJWK(ctx context.Context, tenantID, realmID googleUuid.UUID, kid string) (*cryptoutilJoseDomain.ElasticJWK, error) {
	elasticJWK, err := s.elasticRepo.Get(ctx, tenantID, realmID, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	return elasticJWK, nil
}

// ListElasticJWKs retrieves all Elastic JWKs for a tenant/realm.
func (s *ElasticJWKService) ListElasticJWKs(ctx context.Context, tenantID, realmID googleUuid.UUID, offset, limit int) ([]cryptoutilJoseDomain.ElasticJWK, error) {
	elasticJWKs, err := s.elasticRepo.List(ctx, tenantID, realmID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list elastic JWKs: %w", err)
	}

	return elasticJWKs, nil
}

// GetActiveMaterialJWK retrieves the active Material JWK for an Elastic JWK.
func (s *ElasticJWKService) GetActiveMaterialJWK(ctx context.Context, elasticJWKID googleUuid.UUID) (*cryptoutilJoseDomain.MaterialJWK, joseJwk.Key, joseJwk.Key, error) {
	materialJWK, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get active material JWK: %w", err)
	}

	// Decrypt the keys.
	privateJWK, publicJWK, err := s.decryptMaterialKeys(ctx, materialJWK)
	if err != nil {
		return nil, nil, nil, err
	}

	return materialJWK, privateJWK, publicJWK, nil
}

// GetMaterialJWKByKID retrieves a Material JWK by its material KID (for historical access).
func (s *ElasticJWKService) GetMaterialJWKByKID(ctx context.Context, elasticJWKID googleUuid.UUID, materialKID string) (*cryptoutilJoseDomain.MaterialJWK, joseJwk.Key, joseJwk.Key, error) {
	materialJWK, err := s.materialRepo.GetByMaterialKID(ctx, elasticJWKID, materialKID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get material JWK by KID: %w", err)
	}

	// Decrypt the keys.
	privateJWK, publicJWK, err := s.decryptMaterialKeys(ctx, materialJWK)
	if err != nil {
		return nil, nil, nil, err
	}

	return materialJWK, privateJWK, publicJWK, nil
}

// decryptMaterialKeys decrypts the private and public JWKs from a Material JWK.
func (s *ElasticJWKService) decryptMaterialKeys(ctx context.Context, materialJWK *cryptoutilJoseDomain.MaterialJWK) (joseJwk.Key, joseJwk.Key, error) {
	// Decrypt private JWK.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(materialJWK.PrivateJWKJWE))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	privateJWK, err := joseJwk.ParseKey(privateJWKJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Decrypt public JWK.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, []byte(materialJWK.PublicJWKJWE))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	publicJWK, err := joseJwk.ParseKey(publicJWKJSON)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public JWK: %w", err)
	}

	return privateJWK, publicJWK, nil
}

// generateJWK generates a new JWK based on key type and algorithm.
func (s *ElasticJWKService) generateJWK(_, alg, use string) (joseJwk.Key, joseJwk.Key, error) {
	// Use the appropriate generation method based on use type.
	if use == "sig" {
		return s.generateSigningJWK(alg)
	}

	return s.generateEncryptionJWK(alg)
}

// generateSigningJWK generates a signing JWK based on algorithm.
func (s *ElasticJWKService) generateSigningJWK(alg string) (joseJwk.Key, joseJwk.Key, error) {
	// Map algorithm string to JWA signature algorithm.
	sigAlg, err := mapToJWASignatureAlgorithm(alg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map signature algorithm: %w", err)
	}

	_, privateJWK, publicJWK, _, _, err := s.jwkGenSvc.GenerateJWSJWK(sigAlg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate signing JWK: %w", err)
	}

	return privateJWK, publicJWK, nil
}

// generateEncryptionJWK generates an encryption JWK based on algorithm.
func (s *ElasticJWKService) generateEncryptionJWK(alg string) (joseJwk.Key, joseJwk.Key, error) {
	// Map algorithm string to JWA encryption algorithms.
	enc, keyAlg, err := mapToJWAEncryptionAlgorithms(alg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to map encryption algorithm: %w", err)
	}

	_, privateJWK, publicJWK, _, _, err := s.jwkGenSvc.GenerateJWEJWK(&enc, &keyAlg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate encryption JWK: %w", err)
	}

	return privateJWK, publicJWK, nil
}

// mapToJWASignatureAlgorithm maps an algorithm string to a JWA signature algorithm.
func mapToJWASignatureAlgorithm(alg string) (joseJwa.SignatureAlgorithm, error) {
	switch alg {
	case "RS256":
		return cryptoutilSharedCryptoJose.AlgRS256, nil
	case "RS384":
		return cryptoutilSharedCryptoJose.AlgRS384, nil
	case "RS512":
		return cryptoutilSharedCryptoJose.AlgRS512, nil
	case "PS256":
		return cryptoutilSharedCryptoJose.AlgPS256, nil
	case "PS384":
		return cryptoutilSharedCryptoJose.AlgPS384, nil
	case "PS512":
		return cryptoutilSharedCryptoJose.AlgPS512, nil
	case "ES256":
		return cryptoutilSharedCryptoJose.AlgES256, nil
	case "ES384":
		return cryptoutilSharedCryptoJose.AlgES384, nil
	case "ES512":
		return cryptoutilSharedCryptoJose.AlgES512, nil
	case "EdDSA":
		return cryptoutilSharedCryptoJose.AlgEdDSA, nil
	case "HS256":
		return cryptoutilSharedCryptoJose.AlgHS256, nil
	case "HS384":
		return cryptoutilSharedCryptoJose.AlgHS384, nil
	case "HS512":
		return cryptoutilSharedCryptoJose.AlgHS512, nil
	default:
		return cryptoutilSharedCryptoJose.AlgSigInvalid, fmt.Errorf("unsupported signature algorithm: %s", alg)
	}
}

// mapToJWAEncryptionAlgorithms maps an algorithm string to JWA content encryption and key encryption algorithms.
func mapToJWAEncryptionAlgorithms(alg string) (joseJwa.ContentEncryptionAlgorithm, joseJwa.KeyEncryptionAlgorithm, error) {
	switch alg {
	case "A128GCM":
		return cryptoutilSharedCryptoJose.EncA128GCM, cryptoutilSharedCryptoJose.AlgDir, nil
	case "A192GCM":
		return cryptoutilSharedCryptoJose.EncA192GCM, cryptoutilSharedCryptoJose.AlgDir, nil
	case "A256GCM":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgDir, nil
	case "A128CBC-HS256":
		return cryptoutilSharedCryptoJose.EncA128CBCHS256, cryptoutilSharedCryptoJose.AlgDir, nil
	case "A192CBC-HS384":
		return cryptoutilSharedCryptoJose.EncA192CBCHS384, cryptoutilSharedCryptoJose.AlgDir, nil
	case "A256CBC-HS512":
		return cryptoutilSharedCryptoJose.EncA256CBCHS512, cryptoutilSharedCryptoJose.AlgDir, nil
	case "RSA-OAEP":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgRSAOAEP, nil
	case "RSA-OAEP-256":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgRSAOAEP256, nil
	case "RSA-OAEP-384":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgRSAOAEP384, nil
	case "RSA-OAEP-512":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgRSAOAEP512, nil
	case "ECDH-ES":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgECDHES, nil
	case "ECDH-ES+A128KW":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgECDHESA128KW, nil
	case "ECDH-ES+A192KW":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgECDHESA192KW, nil
	case "ECDH-ES+A256KW":
		return cryptoutilSharedCryptoJose.EncA256GCM, cryptoutilSharedCryptoJose.AlgECDHESA256KW, nil
	default:
		return cryptoutilSharedCryptoJose.EncInvalid, cryptoutilSharedCryptoJose.AlgEncInvalid, fmt.Errorf("unsupported encryption algorithm: %s", alg)
	}
}

// logAuditSuccess logs a successful operation if audit logging is enabled.
func (s *ElasticJWKService) logAuditSuccess(ctx context.Context, tenantID, realmID googleUuid.UUID, operation, resourceType, resourceID string, metadata map[string]any) {
	if s.auditLogSvc == nil {
		return
	}

	_, _ = s.auditLogSvc.LogSuccess(ctx, tenantID, realmID, operation, resourceType, resourceID, metadata)
}

// logAuditFailure logs a failed operation if audit logging is enabled.
func (s *ElasticJWKService) logAuditFailure(ctx context.Context, tenantID, realmID googleUuid.UUID, operation, resourceType, resourceID string, err error, metadata map[string]any) {
	if s.auditLogSvc == nil {
		return
	}

	_, _ = s.auditLogSvc.LogFailure(ctx, tenantID, realmID, operation, resourceType, resourceID, err, metadata)
}
