// Copyright (c) 2025 Justin Cranford

package crypto

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

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

			// Verify headers.
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

			// Verify headers.
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

			// Verify headers.
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

	// Verify headers.
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

func TestCreateJWSJWKFromKey_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setupFn func(t *testing.T) (*googleUuid.UUID, *joseJwa.SignatureAlgorithm, cryptoutilSharedCryptoKeygen.Key)
		wantErr string
	}{
		{
			name: "unsupported key type",
			setupFn: func(t *testing.T) (*googleUuid.UUID, *joseJwa.SignatureAlgorithm, cryptoutilSharedCryptoKeygen.Key) {
				t.Helper()

				kid := googleUuid.New()
				alg := joseJwa.RS256()

				return &kid, &alg, &cryptoutilSharedCryptoKeygen.KeyPair{Private: 12345, Public: nil}
			},
			wantErr: "invalid key type",
		},
		{
			name: "empty secret key",
			setupFn: func(t *testing.T) (*googleUuid.UUID, *joseJwa.SignatureAlgorithm, cryptoutilSharedCryptoKeygen.Key) {
				t.Helper()

				kid := googleUuid.Must(googleUuid.NewV7())
				alg := joseJwa.HS256()

				return &kid, &alg, cryptoutilSharedCryptoKeygen.SecretKey("")
			},
			wantErr: "invalid JWS JWK headers",
		},
		{
			name: "nil private key in pair",
			setupFn: func(t *testing.T) (*googleUuid.UUID, *joseJwa.SignatureAlgorithm, cryptoutilSharedCryptoKeygen.Key) {
				t.Helper()

				kid := googleUuid.Must(googleUuid.NewV7())
				alg := joseJwa.ES256()

				return &kid, &alg, &cryptoutilSharedCryptoKeygen.KeyPair{Private: nil, Public: nil}
			},
			wantErr: "invalid JWS JWK headers",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			kid, alg, key := tt.setupFn(t)
			_, _, _, _, _, err := CreateJWSJWKFromKey(kid, alg, key)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestCreateJWSJWKFromKey_SetKidSuccess(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.HS256()
	validKey, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	_, _, _, _, _, err = CreateJWSJWKFromKey(&kid, &alg, validKey)
	require.NoError(t, err)
}

func TestCreateJWSJWKFromKey_PublicKeyExtraction(t *testing.T) {
	t.Parallel()

	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	kid := googleUuid.New()
	alg := joseJwa.RS256()

	_, nonPublicJWK, publicJWK, _, clearPublicBytes, err := CreateJWSJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, clearPublicBytes)

	var kidFromPublic string

	require.NoError(t, publicJWK.Get(joseJwk.KeyIDKey, &kidFromPublic))
	require.Equal(t, kid.String(), kidFromPublic)
}

func TestCreateJWSJWKFromKey_HMACNoPublicKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := joseJwa.HS256()
	key := make(cryptoutilSharedCryptoKeygen.SecretKey, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(key)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWSJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.Empty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}
