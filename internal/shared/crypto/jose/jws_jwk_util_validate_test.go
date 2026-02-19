// Copyright (c) 2025 Justin Cranford

package crypto

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/stretchr/testify/require"
)

func TestValidateOrGenerateJWSEcdsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEcdsaJWK(validKey, joseJwa.ES256(), elliptic.P256())
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEcdsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEcdsaJWK(wrongKey, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEcdsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	// KeyPair with nil private key.
	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: (*ecdsa.PrivateKey)(nil),
		Public:  &ecdsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil ECDSA private key")
}

func TestValidateOrGenerateJWSEcdsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Create KeyPair with valid private and nil public.
	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: validKey.Private,
		Public:  (*ecdsa.PublicKey)(nil),
	}

	validated, err := validateOrGenerateJWSEcdsaJWK(keyPair, joseJwa.ES256(), elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid nil ECDSA public key")
}

func TestValidateOrGenerateJWSEddsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid Ed25519 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSEddsaJWK(validKey, joseJwa.EdDSA(), "Ed25519")
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSEddsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateJWSEddsaJWK(wrongKey, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSEddsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSEddsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	_ = publicKey

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSEddsaJWK(keyPair, joseJwa.EdDSA(), "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSHMACJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid HMAC 256 key.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateJWSHMACJWK(validKey, joseJwa.HS256(), 256)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSHMACJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSHMACJWK(wrongKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSHMACJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilSharedCryptoKeygen.SecretKey

	result, err := validateOrGenerateJWSHMACJWK(nilKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestValidateOrGenerateJWSHMACJWK_WrongKeyLength(t *testing.T) {
	t.Parallel()

	// Generate 512-bit key but expect 256-bit.
	wrongLengthKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(512)
	require.NoError(t, err)

	result, err := validateOrGenerateJWSHMACJWK(wrongLengthKey, joseJwa.HS256(), 256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid key length")
}

func TestValidateOrGenerateJWSRSAJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	validKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(validKey, joseJwa.RS256(), 2048)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateJWSRSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(256)
	require.NoError(t, err)

	validated, err := validateOrGenerateJWSRSAJWK(wrongKey, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateJWSRSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateJWSRSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateJWSRSAJWK(keyPair, joseJwa.RS256(), 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWSJWKFromKey_ECDSA_AllCurves tests ECDSA with all curves.
