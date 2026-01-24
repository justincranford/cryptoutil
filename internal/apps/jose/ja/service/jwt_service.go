// Copyright (c) 2025 Justin Cranford
//

// Package service provides business logic services for jose-ja.
package service

import (
	"context"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"time"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	googleUuid "github.com/google/uuid"
)

// JWTClaims represents standard JWT claims.
type JWTClaims struct {
	Issuer    string                 `json:"iss,omitempty"`
	Subject   string                 `json:"sub,omitempty"`
	Audience  []string               `json:"aud,omitempty"`
	ExpiresAt *time.Time             `json:"exp,omitempty"`
	NotBefore *time.Time             `json:"nbf,omitempty"`
	IssuedAt  *time.Time             `json:"iat,omitempty"`
	JTI       string                 `json:"jti,omitempty"`
	Custom    map[string]interface{} `json:"-"` // Additional custom claims.
}

// JWTService provides business logic for JWT operations.
type JWTService interface {
	// CreateJWT creates a signed JWT with the given claims.
	CreateJWT(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, claims *JWTClaims) (string, error)

	// ValidateJWT validates a JWT and returns the claims.
	ValidateJWT(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, token string) (*JWTClaims, error)

	// CreateEncryptedJWT creates an encrypted JWT (JWE containing JWT).
	CreateEncryptedJWT(ctx context.Context, tenantID, signingKeyID, encryptionKeyID googleUuid.UUID, claims *JWTClaims) (string, error)
}

// jwtServiceImpl implements JWTService.
type jwtServiceImpl struct {
	elasticRepo  cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	barrierSvc   *cryptoutilAppsTemplateServiceServerBarrier.Service
}

// NewJWTService creates a new JWTService.
func NewJWTService(
	elasticRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	barrierSvc *cryptoutilAppsTemplateServiceServerBarrier.Service,
) JWTService {
	return &jwtServiceImpl{
		elasticRepo:  elasticRepo,
		materialRepo: materialRepo,
		barrierSvc:   barrierSvc,
	}
}

// CreateJWT creates a signed JWT with the given claims.
func (s *jwtServiceImpl) CreateJWT(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, claims *JWTClaims) (string, error) {
	// Verify tenant ownership and get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return "", fmt.Errorf("elastic JWK not found")
	}

	// Validate key use for signing.
	if elasticJWK.Use != cryptoutilAppsJoseJaDomain.KeyUseSig {
		return "", fmt.Errorf("key is not configured for signing (use=%s)", elasticJWK.Use)
	}

	// Get active material.
	material, err := s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
	if err != nil {
		return "", fmt.Errorf("failed to get active material: %w", err)
	}

	// Decode base64 encoded JWE string.
	privateJWKEncrypted, err := base64.StdEncoding.DecodeString(material.PrivateJWKJWE)
	if err != nil {
		return "", fmt.Errorf("failed to decode private JWK JWE: %w", err)
	}

	// Decrypt private JWK using barrier service.
	privateJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, privateJWKEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private JWK: %w", err)
	}

	// Parse private JWK.
	var privateJWK jose.JSONWebKey
	if err := privateJWK.UnmarshalJSON(privateJWKJSON); err != nil {
		return "", fmt.Errorf("failed to parse private JWK: %w", err)
	}

	// Set KID on private JWK.
	privateJWK.KeyID = material.MaterialKID

	// Create JWT signer.
	sigAlg := jose.SignatureAlgorithm(elasticJWK.Algorithm)
	signerOpts := (&jose.SignerOptions{}).WithHeader(jose.HeaderKey("kid"), material.MaterialKID)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: sigAlg, Key: privateJWK}, signerOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %w", err)
	}

	// Build claims map.
	claimsMap := s.buildClaimsMap(claims)

	// Create and sign JWT.
	builder := jwt.Signed(signer).Claims(claimsMap)

	token, err := builder.Serialize()
	if err != nil {
		return "", fmt.Errorf("failed to create JWT: %w", err)
	}

	return token, nil
}

// ValidateJWT validates a JWT and returns the claims.
func (s *jwtServiceImpl) ValidateJWT(ctx context.Context, tenantID, elasticJWKID googleUuid.UUID, token string) (*JWTClaims, error) {
	// Verify tenant ownership and get elastic JWK.
	elasticJWK, err := s.elasticRepo.GetByID(ctx, elasticJWKID)
	if err != nil {
		return nil, fmt.Errorf("failed to get elastic JWK: %w", err)
	}

	if elasticJWK.TenantID != tenantID {
		return nil, fmt.Errorf("elastic JWK not found")
	}

	// Parse JWT to extract KID.
	parsedJWT, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{jose.SignatureAlgorithm(elasticJWK.Algorithm)})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Get the KID from header.
	if len(parsedJWT.Headers) == 0 {
		return nil, fmt.Errorf("JWT has no headers")
	}

	kid := parsedJWT.Headers[0].KeyID

	// Get material by KID.
	var material *cryptoutilAppsJoseJaDomain.MaterialJWK

	if kid != "" {
		material, err = s.materialRepo.GetByMaterialKID(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get material by KID: %w", err)
		}

		// Verify the material belongs to the correct elastic JWK.
		if material.ElasticJWKID != elasticJWKID {
			return nil, fmt.Errorf("material KID does not belong to this elastic JWK")
		}
	} else {
		// Fallback to active material.
		material, err = s.materialRepo.GetActiveMaterial(ctx, elasticJWKID)
		if err != nil {
			return nil, fmt.Errorf("failed to get active material: %w", err)
		}
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

	// Verify and extract claims.
	var claimsMap map[string]interface{}
	if err := parsedJWT.Claims(publicJWK, &claimsMap); err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	// Convert claims map to JWTClaims.
	claims := s.parseClaimsMap(claimsMap)

	// Validate expiration.
	now := time.Now()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(now) {
		return nil, fmt.Errorf("JWT has expired")
	}

	// Validate not-before.
	if claims.NotBefore != nil && claims.NotBefore.After(now) {
		return nil, fmt.Errorf("JWT is not yet valid")
	}

	return claims, nil
}

// CreateEncryptedJWT creates an encrypted JWT (JWE containing signed JWT).
func (s *jwtServiceImpl) CreateEncryptedJWT(ctx context.Context, tenantID, signingKeyID, encryptionKeyID googleUuid.UUID, claims *JWTClaims) (string, error) {
	// First, create a signed JWT.
	signedJWT, err := s.CreateJWT(ctx, tenantID, signingKeyID, claims)
	if err != nil {
		return "", fmt.Errorf("failed to create signed JWT: %w", err)
	}

	// Get encryption key.
	encryptionKey, err := s.elasticRepo.GetByID(ctx, encryptionKeyID)
	if err != nil {
		return "", fmt.Errorf("failed to get encryption key: %w", err)
	}

	if encryptionKey.TenantID != tenantID {
		return "", fmt.Errorf("encryption key not found")
	}

	// Validate key use for encryption.
	if encryptionKey.Use != cryptoutilAppsJoseJaDomain.KeyUseEnc {
		return "", fmt.Errorf("key is not configured for encryption (use=%s)", encryptionKey.Use)
	}

	// Get active material for encryption.
	encMaterial, err := s.materialRepo.GetActiveMaterial(ctx, encryptionKeyID)
	if err != nil {
		return "", fmt.Errorf("failed to get encryption material: %w", err)
	}

	// Decode base64 encoded JWE string.
	publicJWKEncrypted, err := base64.StdEncoding.DecodeString(encMaterial.PublicJWKJWE)
	if err != nil {
		return "", fmt.Errorf("failed to decode public JWK JWE: %w", err)
	}

	// Decrypt public JWK for encryption using barrier service.
	publicJWKJSON, err := s.barrierSvc.DecryptContentWithContext(ctx, publicJWKEncrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt public JWK: %w", err)
	}

	// Parse public JWK.
	var publicJWK jose.JSONWebKey
	if err := publicJWK.UnmarshalJSON(publicJWKJSON); err != nil {
		return "", fmt.Errorf("failed to parse public JWK: %w", err)
	}

	// Determine encryption algorithms.
	keyAlg, contentEnc := mapToJWEAlgorithms(encryptionKey.Algorithm)
	if keyAlg == "" {
		return "", fmt.Errorf("unsupported algorithm for JWE: %s", encryptionKey.Algorithm)
	}

	// Create encrypter with content type header.
	encrypterOpts := (&jose.EncrypterOptions{}).WithContentType("JWT")

	encrypter, err := jose.NewEncrypter(contentEnc, jose.Recipient{Algorithm: keyAlg, Key: publicJWK}, encrypterOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create encrypter: %w", err)
	}

	// Encrypt the signed JWT.
	jweObject, err := encrypter.Encrypt([]byte(signedJWT))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt JWT: %w", err)
	}

	// Serialize to compact format.
	compact, err := jweObject.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize encrypted JWT: %w", err)
	}

	return compact, nil
}

// buildClaimsMap converts JWTClaims to a map for JWT creation.
func (s *jwtServiceImpl) buildClaimsMap(claims *JWTClaims) map[string]interface{} {
	claimsMap := make(map[string]interface{})

	if claims.Issuer != "" {
		claimsMap["iss"] = claims.Issuer
	}

	if claims.Subject != "" {
		claimsMap["sub"] = claims.Subject
	}

	if len(claims.Audience) > 0 {
		if len(claims.Audience) == 1 {
			claimsMap["aud"] = claims.Audience[0]
		} else {
			claimsMap["aud"] = claims.Audience
		}
	}

	if claims.ExpiresAt != nil {
		claimsMap["exp"] = claims.ExpiresAt.Unix()
	}

	if claims.NotBefore != nil {
		claimsMap["nbf"] = claims.NotBefore.Unix()
	}

	if claims.IssuedAt != nil {
		claimsMap["iat"] = claims.IssuedAt.Unix()
	}

	if claims.JTI != "" {
		claimsMap["jti"] = claims.JTI
	}

	// Add custom claims.
	for k, v := range claims.Custom {
		claimsMap[k] = v
	}

	return claimsMap
}

// parseClaimsMap converts a claims map back to JWTClaims.
func (s *jwtServiceImpl) parseClaimsMap(claimsMap map[string]interface{}) *JWTClaims {
	claims := &JWTClaims{
		Custom: make(map[string]interface{}),
	}

	for k, v := range claimsMap {
		switch k {
		case "iss":
			if str, ok := v.(string); ok {
				claims.Issuer = str
			}
		case "sub":
			if str, ok := v.(string); ok {
				claims.Subject = str
			}
		case "aud":
			switch a := v.(type) {
			case string:
				claims.Audience = []string{a}
			case []interface{}:
				for _, item := range a {
					if str, ok := item.(string); ok {
						claims.Audience = append(claims.Audience, str)
					}
				}
			}
		case "exp":
			if f, ok := v.(float64); ok {
				t := time.Unix(int64(f), 0)
				claims.ExpiresAt = &t
			} else if n, ok := v.(json.Number); ok {
				i, _ := n.Int64()
				t := time.Unix(i, 0)
				claims.ExpiresAt = &t
			}
		case "nbf":
			if f, ok := v.(float64); ok {
				t := time.Unix(int64(f), 0)
				claims.NotBefore = &t
			} else if n, ok := v.(json.Number); ok {
				i, _ := n.Int64()
				t := time.Unix(i, 0)
				claims.NotBefore = &t
			}
		case "iat":
			if f, ok := v.(float64); ok {
				t := time.Unix(int64(f), 0)
				claims.IssuedAt = &t
			} else if n, ok := v.(json.Number); ok {
				i, _ := n.Int64()
				t := time.Unix(i, 0)
				claims.IssuedAt = &t
			}
		case "jti":
			if str, ok := v.(string); ok {
				claims.JTI = str
			}
		default:
			claims.Custom[k] = v
		}
	}

	return claims
}
