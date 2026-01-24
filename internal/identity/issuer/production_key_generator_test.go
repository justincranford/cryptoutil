// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	"context"
	ecdsa "crypto/ecdsa"
	rsa "crypto/rsa"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

func TestProductionKeyGenerator_GenerateSigningKey_RS256(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "RS256")

	require.NoError(t, err, "RS256 key generation should succeed")
	require.NotNil(t, key, "RS256 key should not be nil")
	require.Equal(t, "RS256", key.Algorithm, "Algorithm should be RS256")
	require.NotEmpty(t, key.KeyID, "KeyID should be generated")
	require.False(t, key.Active, "New key should not be active")
	require.False(t, key.ValidForVerif, "New key should not be valid for verification")

	rsaKey, ok := key.Key.(*rsa.PrivateKey)
	require.True(t, ok, "Key material should be RSA private key")
	require.Equal(t, cryptoutilIdentityMagic.RSA2048KeySize, rsaKey.N.BitLen(), "RS256 should use 2048-bit RSA key")
}

func TestProductionKeyGenerator_GenerateSigningKey_RS384(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "RS384")

	require.NoError(t, err, "RS384 key generation should succeed")
	require.NotNil(t, key, "RS384 key should not be nil")
	require.Equal(t, "RS384", key.Algorithm, "Algorithm should be RS384")

	rsaKey, ok := key.Key.(*rsa.PrivateKey)
	require.True(t, ok, "Key material should be RSA private key")
	require.Equal(t, cryptoutilIdentityMagic.RSA3072KeySize, rsaKey.N.BitLen(), "RS384 should use 3072-bit RSA key")
}

func TestProductionKeyGenerator_GenerateSigningKey_RS512(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "RS512")

	require.NoError(t, err, "RS512 key generation should succeed")
	require.NotNil(t, key, "RS512 key should not be nil")
	require.Equal(t, "RS512", key.Algorithm, "Algorithm should be RS512")

	rsaKey, ok := key.Key.(*rsa.PrivateKey)
	require.True(t, ok, "Key material should be RSA private key")
	require.Equal(t, cryptoutilIdentityMagic.RSA4096KeySize, rsaKey.N.BitLen(), "RS512 should use 4096-bit RSA key")
}

func TestProductionKeyGenerator_GenerateSigningKey_ES256(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "ES256")

	require.NoError(t, err, "ES256 key generation should succeed")
	require.NotNil(t, key, "ES256 key should not be nil")
	require.Equal(t, "ES256", key.Algorithm, "Algorithm should be ES256")

	ecKey, ok := key.Key.(*ecdsa.PrivateKey)
	require.True(t, ok, "Key material should be ECDSA private key")
	require.Equal(t, 256, ecKey.Curve.Params().BitSize, "ES256 should use P-256 curve")
}

func TestProductionKeyGenerator_GenerateSigningKey_ES384(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "ES384")

	require.NoError(t, err, "ES384 key generation should succeed")
	require.NotNil(t, key, "ES384 key should not be nil")
	require.Equal(t, "ES384", key.Algorithm, "Algorithm should be ES384")

	ecKey, ok := key.Key.(*ecdsa.PrivateKey)
	require.True(t, ok, "Key material should be ECDSA private key")
	require.Equal(t, 384, ecKey.Curve.Params().BitSize, "ES384 should use P-384 curve")
}

func TestProductionKeyGenerator_GenerateSigningKey_ES512(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "ES512")

	require.NoError(t, err, "ES512 key generation should succeed")
	require.NotNil(t, key, "ES512 key should not be nil")
	require.Equal(t, "ES512", key.Algorithm, "Algorithm should be ES512")

	ecKey, ok := key.Key.(*ecdsa.PrivateKey)
	require.True(t, ok, "Key material should be ECDSA private key")
	require.Equal(t, 521, ecKey.Curve.Params().BitSize, "ES512 should use P-521 curve")
}

func TestProductionKeyGenerator_GenerateSigningKey_HS256(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "HS256")

	require.NoError(t, err, "HS256 key generation should succeed")
	require.NotNil(t, key, "HS256 key should not be nil")
	require.Equal(t, "HS256", key.Algorithm, "Algorithm should be HS256")

	hmacKey, ok := key.Key.([]byte)
	require.True(t, ok, "Key material should be byte slice")
	require.Equal(t, cryptoutilIdentityMagic.HMACSHA256KeySize, len(hmacKey), "HS256 should use 32-byte HMAC key")
}

func TestProductionKeyGenerator_GenerateSigningKey_HS384(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "HS384")

	require.NoError(t, err, "HS384 key generation should succeed")
	require.NotNil(t, key, "HS384 key should not be nil")
	require.Equal(t, "HS384", key.Algorithm, "Algorithm should be HS384")

	hmacKey, ok := key.Key.([]byte)
	require.True(t, ok, "Key material should be byte slice")
	require.Equal(t, cryptoutilIdentityMagic.HMACSHA384KeySize, len(hmacKey), "HS384 should use 48-byte HMAC key")
}

func TestProductionKeyGenerator_GenerateSigningKey_HS512(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "HS512")

	require.NoError(t, err, "HS512 key generation should succeed")
	require.NotNil(t, key, "HS512 key should not be nil")
	require.Equal(t, "HS512", key.Algorithm, "Algorithm should be HS512")

	hmacKey, ok := key.Key.([]byte)
	require.True(t, ok, "Key material should be byte slice")
	require.Equal(t, cryptoutilIdentityMagic.HMACSHA512KeySize, len(hmacKey), "HS512 should use 64-byte HMAC key")
}

func TestProductionKeyGenerator_GenerateSigningKey_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateSigningKey(ctx, "INVALID")

	require.Error(t, err, "Unsupported algorithm should fail")
	require.Nil(t, key, "Key should be nil on error")
	require.Contains(t, err.Error(), "unsupported signing algorithm", "Error should mention unsupported algorithm")
}

func TestProductionKeyGenerator_GenerateEncryptionKey(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := generator.GenerateEncryptionKey(ctx)

	require.NoError(t, err, "Encryption key generation should succeed")
	require.NotNil(t, key, "Encryption key should not be nil")
	require.NotEmpty(t, key.KeyID, "KeyID should be generated")
	require.Equal(t, cryptoutilIdentityMagic.AES256KeySize, len(key.Key), "Encryption key should be 32 bytes (AES-256)")
	require.False(t, key.Active, "New key should not be active")
	require.False(t, key.ValidForDecr, "New key should not be valid for decryption")
	require.NotZero(t, key.CreatedAt, "CreatedAt should be set")
	require.NotZero(t, key.ExpiresAt, "ExpiresAt should be set")
}

func TestProductionKeyGenerator_KeyUniqueness(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	// Generate multiple keys and ensure they are unique.
	key1, err1 := generator.GenerateSigningKey(ctx, "RS256")
	key2, err2 := generator.GenerateSigningKey(ctx, "RS256")

	require.NoError(t, err1, "First key generation should succeed")
	require.NoError(t, err2, "Second key generation should succeed")
	require.NotEqual(t, key1.KeyID, key2.KeyID, "Key IDs should be unique")
	require.NotEqual(t, key1.Key, key2.Key, "Key materials should be different")
}

func TestProductionKeyGenerator_Integration_KeyRotationManager(t *testing.T) {
	t.Parallel()

	generator := cryptoutilIdentityIssuer.NewProductionKeyGenerator()
	ctx := context.Background()

	policy := cryptoutilIdentityIssuer.DefaultKeyRotationPolicy()

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		policy,
		generator,
		func(keyID string) {
			t.Logf("Key rotated: %s", keyID)
		},
	)
	require.NoError(t, err, "KeyRotationManager creation should succeed")

	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err, "RotateSigningKey should succeed with ProductionKeyGenerator")

	activeKey, err := keyRotationMgr.GetActiveSigningKey()
	require.NoError(t, err, "GetActiveSigningKey should succeed")
	require.Equal(t, "RS256", activeKey.Algorithm, "Active key should be RS256")

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err, "RotateEncryptionKey should succeed with ProductionKeyGenerator")

	activeEncKey, err := keyRotationMgr.GetActiveEncryptionKey()
	require.NoError(t, err, "GetActiveEncryptionKey should succeed")
	require.NotNil(t, activeEncKey, "Active encryption key should not be nil")
}
