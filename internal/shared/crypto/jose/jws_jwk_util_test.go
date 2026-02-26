// Copyright (c) 2025 Justin Cranford

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// TestCreateJWSJWKFromKey_HMAC tests HMAC key import.
func TestCreateJWSJWKFromKey_HMAC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{cryptoutilSharedMagic.JoseAlgHS256, joseJwa.HS256(), cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes},
		{cryptoutilSharedMagic.JoseAlgHS384, joseJwa.HS384(), cryptoutilSharedMagic.HMACSHA384KeySize},
		{cryptoutilSharedMagic.JoseAlgHS512, joseJwa.HS512(), cryptoutilSharedMagic.MinSerialNumberBits},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			key := make(cryptoutilSharedCryptoKeygen.SecretKey, tt.keySize)
			_, err := crand.Read(key)
			require.NoError(t, err)

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, key)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.Nil(t, publicJWK) // HMAC has no public key
			require.NotEmpty(t, nonPublicBytes)
			require.Empty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyOCT, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_RSA tests RSA key import.
func TestCreateJWSJWKFromKey_RSA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		alg     joseJwa.SignatureAlgorithm
		keySize int
	}{
		{cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, joseJwa.RS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
		{cryptoutilSharedMagic.JoseAlgRS384, joseJwa.RS384(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
		{cryptoutilSharedMagic.JoseAlgRS512, joseJwa.RS512(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
		{cryptoutilSharedMagic.JoseAlgPS256, joseJwa.PS256(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
		{cryptoutilSharedMagic.JoseAlgPS384, joseJwa.PS384(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
		{cryptoutilSharedMagic.JoseAlgPS512, joseJwa.PS512(), cryptoutilSharedMagic.DefaultMetricsBatchSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			privateKey, err := rsa.GenerateKey(crand.Reader, tt.keySize)
			require.NoError(t, err)

			keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
				Private: privateKey,
				Public:  &privateKey.PublicKey,
			}

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyRSA, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_ECDSA tests ECDSA key import.
func TestCreateJWSJWKFromKey_ECDSA(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		alg   joseJwa.SignatureAlgorithm
		curve elliptic.Curve
	}{
		{cryptoutilSharedMagic.JoseAlgES256, joseJwa.ES256(), elliptic.P256()},
		{cryptoutilSharedMagic.JoseAlgES384, joseJwa.ES384(), elliptic.P384()},
		{cryptoutilSharedMagic.JoseAlgES512, joseJwa.ES512(), elliptic.P521()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid := googleUuid.New()
			privateKey, err := ecdsa.GenerateKey(tt.curve, crand.Reader)
			require.NoError(t, err)

			keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
				Private: privateKey,
				Public:  &privateKey.PublicKey,
			}

			resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &tt.alg, keyPair)
			require.NoError(t, err)
			require.Equal(t, &kid, resultKid)
			require.NotNil(t, nonPublicJWK)
			require.NotNil(t, publicJWK)
			require.NotEmpty(t, nonPublicBytes)
			require.NotEmpty(t, publicBytes)

			// Verify headers
			keyID, ok := nonPublicJWK.KeyID()
			require.True(t, ok)
			require.Equal(t, kid.String(), keyID)

			alg, ok := nonPublicJWK.Algorithm()
			require.True(t, ok)
			require.Equal(t, tt.alg, alg)

			require.Equal(t, KtyEC, nonPublicJWK.KeyType())

			usage, ok := nonPublicJWK.KeyUsage()
			require.True(t, ok)
			require.Equal(t, joseJwk.ForSignature.String(), usage)
		})
	}
}

// TestCreateJWSJWKFromKey_EdDSA tests EdDSA key import.
func TestCreateJWSJWKFromKey_EdDSA(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	_, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  privateKey.Public(),
	}

	alg := joseJwa.EdDSA()

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify headers
	keyID, ok := nonPublicJWK.KeyID()
	require.True(t, ok)
	require.Equal(t, kid.String(), keyID)

	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, alg, algVal)

	require.Equal(t, KtyOKP, nonPublicJWK.KeyType())

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, joseJwk.ForSignature.String(), usage)
}

// TestCreateJWSJWKFromKey_UnsupportedKeyType tests error for unsupported key types.
func TestCreateJWSJWKFromKey_UnsupportedKeyType(t *testing.T) {
	t.Parallel()

	// Test with KeyPair containing unsupported key type (int instead of crypto key)
	kid := googleUuid.New()
	alg := joseJwa.RS256()

	// KeyPair with invalid Private field type will hit default case in switch
	invalidKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: 12345, // Invalid type - not *rsa.PrivateKey, *ecdsa.PrivateKey, or ed25519.PrivateKey
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWSJWKFromKey_NilKid tests error for nil KID.
func TestCreateJWSJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	alg := joseJwa.RS256()
	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	_, _, _, _, _, err = CreateJWSJWKFromKey(nil, &alg, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_NilAlg tests error for nil algorithm.
func TestCreateJWSJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()

	// Use simple KeyPair with valid key type to test nil alg validation
	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	keyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: privateKey,
		Public:  &privateKey.PublicKey,
	}

	// Should error on nil alg validation before key type checks
	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, nil, keyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

func TestCreateJWSJWKFromKey_SetKidError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.HS256()
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, &alg, validKey)
	require.NoError(t, err)
}

// TestCreateJWSJWKFromKey_NilKey tests error for nil key.
func TestCreateJWSJWKFromKey_NilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.RS256()

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

func TestCreateJWSJWKFromKey_PublicKeyExtraction(t *testing.T) {
	t.Parallel()

	// Generate RSA key pair.
	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	kid := googleUuid.New()
	alg := joseJwa.RS256()

	// Create JWK from key pair.
	_, nonPublicJWK, publicJWK, _, clearPublicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, clearPublicBytes)

	// Verify public key extracted successfully.
	var kidFromPublic string

	require.NoError(t, publicJWK.Get(joseJwk.KeyIDKey, &kidFromPublic))
	require.Equal(t, kid.String(), kidFromPublic)
}

func TestCreateJWSJWKFromKey_HMACNoPublicKey(t *testing.T) {
	t.Parallel()

	// HMAC has no public key, so publicJWK should be nil and clearPublicBytes empty.
	kid := googleUuid.New()
	alg := joseJwa.HS256()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // HMAC should have no public key
	require.Empty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

// TestCreateJWSJWKFromKey_ImportSecretKeyError tests validation error for empty HMAC key.
func TestCreateJWSJWKFromKey_ImportSecretKeyError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS256()
	// Empty SecretKey fails validation before import
	emptyKey := cryptoutilSharedCryptoKeygen.SecretKey("")

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, emptyKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}

// TestCreateJWSJWKFromKey_ImportKeyPairError tests validation error for nil KeyPair.
func TestCreateJWSJWKFromKey_ImportKeyPairError(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.ES256()
	// KeyPair with nil Private fails validation before import
	invalidKeyPair := &cryptoutilSharedCryptoKeygen.KeyPair{
		Private: nil,
		Public:  nil,
	}

	_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, invalidKeyPair)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWS JWK headers")
}
