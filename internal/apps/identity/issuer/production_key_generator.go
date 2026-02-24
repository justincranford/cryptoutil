// Copyright (c) 2025 Justin Cranford
//
//

package issuer

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ProductionKeyGenerator implements KeyGenerator for production use.
type ProductionKeyGenerator struct{}

// NewProductionKeyGenerator creates a new production key generator.
func NewProductionKeyGenerator() *ProductionKeyGenerator {
	return &ProductionKeyGenerator{}
}

// GenerateSigningKey generates a signing key for the specified algorithm.
func (g *ProductionKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*SigningKey, error) {
	switch algorithm {
	case cryptoutilSharedMagic.AlgorithmRS256, cryptoutilSharedMagic.AlgorithmRS384, cryptoutilSharedMagic.AlgorithmRS512:
		return g.generateRSASigningKey(ctx, algorithm)
	case cryptoutilSharedMagic.AlgorithmES256, cryptoutilSharedMagic.AlgorithmES384, cryptoutilSharedMagic.AlgorithmES512:
		return g.generateECDSASigningKey(ctx, algorithm)
	case cryptoutilSharedMagic.AlgorithmHS256, cryptoutilSharedMagic.AlgorithmHS384, cryptoutilSharedMagic.AlgorithmHS512:
		return g.generateHMACSigningKey(ctx, algorithm)
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("unsupported signing algorithm: %s", algorithm),
		)
	}
}

// GenerateEncryptionKey generates an AES-256 encryption key.
func (g *ProductionKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	keyBytes := make([]byte, cryptoutilSharedMagic.AES256KeySize)

	if _, err := crand.Read(keyBytes); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyGenerationFailed,
			fmt.Errorf("failed to generate AES-256 key: %w", err),
		)
	}

	now := time.Now().UTC()

	return &EncryptionKey{
		KeyID:        googleUuid.NewString(),
		Key:          keyBytes,
		CreatedAt:    now,
		ExpiresAt:    now.Add(cryptoutilSharedMagic.DefaultKeyRotationInterval + cryptoutilSharedMagic.DefaultKeyGracePeriod),
		Active:       false,
		ValidForDecr: false,
	}, nil
}

// generateRSASigningKey generates an RSA signing key.
func (g *ProductionKeyGenerator) generateRSASigningKey(_ context.Context, algorithm string) (*SigningKey, error) {
	var keySize int

	switch algorithm {
	case "RS256":
		keySize = cryptoutilSharedMagic.RSA2048KeySize
	case "RS384":
		keySize = cryptoutilSharedMagic.RSA3072KeySize
	case "RS512":
		keySize = cryptoutilSharedMagic.RSA4096KeySize
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("invalid RSA algorithm: %s", algorithm),
		)
	}

	privateKey, err := rsa.GenerateKey(crand.Reader, keySize)
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyGenerationFailed,
			fmt.Errorf("failed to generate RSA key: %w", err),
		)
	}

	now := time.Now().UTC()

	return &SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           privateKey,
		Algorithm:     algorithm,
		CreatedAt:     now,
		ExpiresAt:     now.Add(cryptoutilSharedMagic.DefaultKeyRotationInterval + cryptoutilSharedMagic.DefaultKeyGracePeriod),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

// generateECDSASigningKey generates an ECDSA signing key.
func (g *ProductionKeyGenerator) generateECDSASigningKey(_ context.Context, algorithm string) (*SigningKey, error) {
	var curve elliptic.Curve

	switch algorithm {
	case "ES256":
		curve = elliptic.P256()
	case "ES384":
		curve = elliptic.P384()
	case "ES512":
		curve = elliptic.P521()
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("invalid ECDSA algorithm: %s", algorithm),
		)
	}

	privateKey, err := ecdsa.GenerateKey(curve, crand.Reader)
	if err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyGenerationFailed,
			fmt.Errorf("failed to generate ECDSA key: %w", err),
		)
	}

	now := time.Now().UTC()

	return &SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           privateKey,
		Algorithm:     algorithm,
		CreatedAt:     now,
		ExpiresAt:     now.Add(cryptoutilSharedMagic.DefaultKeyRotationInterval + cryptoutilSharedMagic.DefaultKeyGracePeriod),
		Active:        false,
		ValidForVerif: false,
	}, nil
}

// generateHMACSigningKey generates an HMAC signing key.
func (g *ProductionKeyGenerator) generateHMACSigningKey(_ context.Context, algorithm string) (*SigningKey, error) {
	var keySize int

	switch algorithm {
	case "HS256":
		keySize = cryptoutilSharedMagic.HMACSHA256KeySize
	case "HS384":
		keySize = cryptoutilSharedMagic.HMACSHA384KeySize
	case "HS512":
		keySize = cryptoutilSharedMagic.HMACSHA512KeySize
	default:
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrInvalidConfiguration,
			fmt.Errorf("invalid HMAC algorithm: %s", algorithm),
		)
	}

	keyBytes := make([]byte, keySize)

	if _, err := crand.Read(keyBytes); err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrKeyGenerationFailed,
			fmt.Errorf("failed to generate HMAC key: %w", err),
		)
	}

	now := time.Now().UTC()

	return &SigningKey{
		KeyID:         googleUuid.NewString(),
		Key:           keyBytes,
		Algorithm:     algorithm,
		CreatedAt:     now,
		ExpiresAt:     now.Add(cryptoutilSharedMagic.DefaultKeyRotationInterval + cryptoutilSharedMagic.DefaultKeyGracePeriod),
		Active:        false,
		ValidForVerif: false,
	}, nil
}
