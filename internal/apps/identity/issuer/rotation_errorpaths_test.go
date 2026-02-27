// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"
)

// failingKeyGenerator implements KeyGenerator and always returns errors.
type failingKeyGenerator struct {
	signingErr    error
	encryptionErr error
}

func (f *failingKeyGenerator) GenerateSigningKey(_ context.Context, _ string) (*SigningKey, error) {
	return nil, f.signingErr
}

func (f *failingKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	return nil, f.encryptionErr
}

// TestGenerateRSASigningKey_InvalidAlgorithm tests invalid RSA algorithm.
func TestGenerateRSASigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateRSASigningKey(ctx, "RS999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid RSA algorithm")
}

// TestGenerateECDSASigningKey_InvalidAlgorithm tests invalid ECDSA algorithm.
func TestGenerateECDSASigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateECDSASigningKey(ctx, "ES999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid ECDSA algorithm")
}

// TestGenerateHMACSigningKey_InvalidAlgorithm tests invalid HMAC algorithm.
func TestGenerateHMACSigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateHMACSigningKey(ctx, "HS999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid HMAC algorithm")
}

// TestEcdsaCurveName_UnknownCurve tests ecdsaCurveName with unsupported curve.
func TestEcdsaCurveName_UnknownCurve(t *testing.T) {
	t.Parallel()

	result := ecdsaCurveName(elliptic.P224())
	testify.Equal(t, "", result)
}

// TestGetPublicKeys_ExpiredAndInvalidKeys tests filtering of expired/invalid keys.
func TestGetPublicKeys_ExpiredAndInvalidKeys(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	testify.NoError(t, err)

	mgr.signingKeys = append(mgr.signingKeys, &SigningKey{
		KeyID:         "expired-key",
		Key:           ecKey,
		Algorithm:     cryptoutilSharedMagic.JoseAlgES256,
		CreatedAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
		ExpiresAt:     time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
		Active:        false,
		ValidForVerif: true,
	})

	mgr.signingKeys = append(mgr.signingKeys, &SigningKey{
		KeyID:         "not-valid-for-verif",
		Key:           ecKey,
		Algorithm:     cryptoutilSharedMagic.JoseAlgES256,
		CreatedAt:     time.Now().UTC(),
		ExpiresAt:     time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
		Active:        false,
		ValidForVerif: false,
	})

	keys := mgr.GetPublicKeys()
	testify.Empty(t, keys)
}

// TestRotateSigningKey_GeneratorFailure tests signing key rotation with failing generator.
func TestRotateSigningKey_GeneratorFailure(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing key generation error"),
		encryptionErr: fmt.Errorf("mock encryption key generation error"),
	}

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), failGen, nil)
	testify.NoError(t, err)

	err = mgr.RotateSigningKey(context.Background(), cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate signing key")
}

// TestRotateEncryptionKey_GeneratorFailure tests encryption key rotation with failing generator.
func TestRotateEncryptionKey_GeneratorFailure(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing key generation error"),
		encryptionErr: fmt.Errorf("mock encryption key generation error"),
	}

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), failGen, nil)
	testify.NoError(t, err)

	err = mgr.RotateEncryptionKey(context.Background())
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate encryption key")
}

// TestRotateEncryptionKey_MaxKeysExceeded tests encryption key pruning when max keys exceeded.
func TestRotateEncryptionKey_MaxKeysExceeded(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	policy := &KeyRotationPolicy{
		RotationInterval:    time.Hour,
		GracePeriod:         time.Minute,
		MaxActiveKeys:       2,
		AutoRotationEnabled: false,
	}

	mgr, err := NewKeyRotationManager(policy, gen, nil)
	testify.NoError(t, err)

	ctx := context.Background()

	for i := 0; i < 3; i++ {
		err = mgr.RotateEncryptionKey(ctx)
		testify.NoError(t, err)
	}

	mgr.mu.RLock()
	keyCount := len(mgr.encryptionKeys)
	mgr.mu.RUnlock()

	testify.LessOrEqual(t, keyCount, policy.MaxActiveKeys)
}

// TestStartAutoRotation_WithErrors tests auto rotation continues after errors.
func TestStartAutoRotation_WithErrors(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing error"),
		encryptionErr: fmt.Errorf("mock encryption error"),
	}

	policy := &KeyRotationPolicy{
		RotationInterval:    cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond,
		GracePeriod:         time.Millisecond,
		MaxActiveKeys:       cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		AutoRotationEnabled: true,
	}

	mgr, err := NewKeyRotationManager(policy, failGen, nil)
	testify.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond)
	defer cancel()

	mgr.StartAutoRotation(ctx, cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm)
	testify.Error(t, ctx.Err())
}

// partialFailKeyGenerator succeeds for signing but fails for encryption.
type partialFailKeyGenerator struct {
	ProductionKeyGenerator
}

func (p *partialFailKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	return nil, fmt.Errorf("encryption key generation failed")
}

// TestStartAutoRotation_SigningSucceedsEncryptionFails tests the second error continue in auto rotation.
func TestStartAutoRotation_SigningSucceedsEncryptionFails(t *testing.T) {
	t.Parallel()

	policy := &KeyRotationPolicy{
		RotationInterval:    cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond,
		GracePeriod:         time.Millisecond,
		MaxActiveKeys:       cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries,
		AutoRotationEnabled: true,
	}

	mgr, err := NewKeyRotationManager(policy, &partialFailKeyGenerator{}, nil)
	testify.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.IMMaxUsernameLength*time.Millisecond)
	defer cancel()

	mgr.StartAutoRotation(ctx, cryptoutilSharedMagic.JoseAlgES256)
	testify.Error(t, ctx.Err())
}
