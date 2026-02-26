// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"crypto/elliptic"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestCreateJWKFromKey_HMAC_HS512(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct512
	key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.DefaultTracesBatchSize)
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
	require.Equal(t, AlgHS512, algVal)
}

// TestCreateJWKFromKey_AES_A128GCM tests AES Oct128 key creation.
func TestCreateJWKFromKey_AES_A128GCM(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct128
	key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.TLSSelfSignedCertSerialNumberBits)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify algorithm is set (stored as string in JWK).
	require.True(t, nonPublicJWK.Has(joseJwk.AlgorithmKey))
	require.True(t, nonPublicJWK.Has(joseJwk.KeyUsageKey))
}

// TestCreateJWKFromKey_AES_A192GCM tests AES Oct192 key creation.
func TestCreateJWKFromKey_AES_A192GCM(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct192
	key, err := cryptoutilSharedCryptoKeygen.GenerateAESKey(cryptoutilSharedMagic.SymmetricKeySize192)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.Empty(t, publicBytes)

	// Verify algorithm is set (stored as string in JWK).
	require.True(t, nonPublicJWK.Has(joseJwk.AlgorithmKey))
}

// TestCreateJWKFromKey_RSA tests RSA key pair creation.
func TestCreateJWKFromKey_RSA(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.RSA2048
	key, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify both keys have KID set.
	keyID, ok := nonPublicJWK.KeyID()
	require.True(t, ok)
	require.Equal(t, kid.String(), keyID)

	pubKeyID, ok := publicJWK.KeyID()
	require.True(t, ok)
	require.Equal(t, kid.String(), pubKeyID)

	// Verify key type.
	var ktyVal joseJwa.KeyType

	require.NoError(t, nonPublicJWK.Get(joseJwk.KeyTypeKey, &ktyVal))
	require.Equal(t, KtyRSA, ktyVal)
}

// TestCreateJWKFromKey_ECDSA tests ECDSA key pair creation.
func TestCreateJWKFromKey_ECDSA(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.ECP256
	key, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	var ktyVal joseJwa.KeyType

	require.NoError(t, nonPublicJWK.Get(joseJwk.KeyTypeKey, &ktyVal))
	require.Equal(t, KtyEC, ktyVal)
}

// TestCreateJWKFromKey_EdDSA tests EdDSA key pair creation.
func TestCreateJWKFromKey_EdDSA(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.OKPEd25519
	key, err := cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair(cryptoutilSharedMagic.EdCurveEd25519)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, key)
	require.NoError(t, err)
	require.NotNil(t, retKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK)
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	var ktyVal joseJwa.KeyType

	require.NoError(t, nonPublicJWK.Get(joseJwk.KeyTypeKey, &ktyVal))
	require.Equal(t, KtyOKP, ktyVal)
}

// TestCreateJWKFromKey_NilKid tests error with nil KID.
func TestCreateJWKFromKey_NilKid(t *testing.T) {
	t.Parallel()

	alg := cryptoutilOpenapiModel.Oct256
	key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(nil, &alg, key)
	require.Error(t, err)
	require.Nil(t, retKid)
	require.Nil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.Empty(t, nonPublicBytes)
	require.Empty(t, publicBytes)
	require.Contains(t, err.Error(), "JWK kid must be valid")
}

// TestCreateJWKFromKey_NilAlg tests error with nil algorithm.
func TestCreateJWKFromKey_NilAlg(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	key, err := cryptoutilSharedCryptoKeygen.GenerateHMACKey(cryptoutilSharedMagic.MaxUnsealSharedSecrets)
	require.NoError(t, err)

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, nil, key)
	require.Error(t, err)
	require.Nil(t, retKid)
	require.Nil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.Empty(t, nonPublicBytes)
	require.Empty(t, publicBytes)
	require.Contains(t, err.Error(), "JWK alg must be non-nil")
}

// TestCreateJWKFromKey_NilKey tests error with nil key.
func TestCreateJWKFromKey_NilKey(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct256

	retKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Nil(t, retKid)
	require.Nil(t, nonPublicJWK)
	require.Nil(t, publicJWK)
	require.Empty(t, nonPublicBytes)
	require.Empty(t, publicBytes)
	require.Contains(t, err.Error(), "JWK key material must be non-nil")
}
