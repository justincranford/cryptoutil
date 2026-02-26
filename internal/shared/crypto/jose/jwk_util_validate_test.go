// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestExtractKty_NilJWK(t *testing.T) {
	t.Parallel()

	kty, err := ExtractKty(nil)
	require.Error(t, err)
	require.Nil(t, kty)
	require.Contains(t, err.Error(), "invalid jwk")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

// TestExtractKty_MissingKeyType removed - JWX v3 always sets kty header on Import.
// Error path "failed to get kty header" is unreachable in normal usage.
// ExtractKty nil check tested in TestExtractKty_NilJWK.
func TestValidateOrGenerateRSAJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid RSA key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateRSAJWK(validKey, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateRSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes))

	validated, err := validateOrGenerateRSAJWK(wrongKey, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateRSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateRSAJWK(keyPair, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateRSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateRSAJWK(keyPair, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEcdsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateEcdsaJWK(validKey, elliptic.P256())
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateEcdsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes))

	validated, err := validateOrGenerateEcdsaJWK(wrongKey, elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateEcdsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  &ecdsa.PublicKey{},
	}

	validated, err := validateOrGenerateEcdsaJWK(keyPair, elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEcdsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateEcdsaJWK(keyPair, elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEddsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid Ed25519 key pair.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateEddsaJWK(validKey, cryptoutilSharedMagic.EdCurveEd25519)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateEddsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilSharedCryptoKeygen.SecretKey(make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes))

	validated, err := validateOrGenerateEddsaJWK(wrongKey, cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateEddsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateEddsaJWK(keyPair, cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEddsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	_ = publicKey

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateEddsaJWK(keyPair, cryptoutilSharedMagic.EdCurveEd25519)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateHMACJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid HMAC 256 key.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateHMACJWK(validKey, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateHMACJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	validated, err := validateOrGenerateHMACJWK(wrongKey, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateAESJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid AES 256 key.
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateAESJWK(validKey, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateAESJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	validated, err := validateOrGenerateAESJWK(wrongKey, cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWKFromKey_HMAC_HS256 tests HMAC Oct256 key creation.
func TestCreateJWKFromKey_HMAC_HS256(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct256
	key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.Equal(t, kid, *retKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify JWK headers.
	keyID, ok := nonPublicJWK.KeyID()
	require.True(t, ok)
	require.Equal(t, kid.String(), keyID)

	var algVal joseJwa.SignatureAlgorithm

	require.NoError(t, nonPublicJWK.Get(joseJwk.AlgorithmKey, &algVal))
	require.Equal(t, AlgHS256, algVal)

	var useVal string

	require.NoError(t, nonPublicJWK.Get(joseJwk.KeyUsageKey, &useVal))
	require.Equal(t, string(joseJwk.ForSignature), useVal)
}

// TestCreateJWKFromKey_HMAC_HS384 tests HMAC Oct384 key creation.
func TestCreateJWKFromKey_HMAC_HS384(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct384
	key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.SymmetricKeySize384)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	var algVal joseJwa.SignatureAlgorithm

	require.NoError(t, nonPublicJWK.Get(joseJwk.AlgorithmKey, &algVal))
	require.Equal(t, AlgHS384, algVal)
}

// TestCreateJWKFromKey_HMAC_HS512 tests HMAC Oct512 key creation.
