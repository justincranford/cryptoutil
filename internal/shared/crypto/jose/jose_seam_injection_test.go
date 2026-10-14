// Copyright (c) 2025-2026 Justin Cranford.
//
//

package crypto

import (
	"crypto/ecdh"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	json "encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/sm-kms/models"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// errInjected is the sentinel error used in all error injection tests.
var errInjected = errors.New("injected error")

// countingSet returns a jwkKeySet that fails on the Nth call.
func countingSet(failOn int) func(joseJwk.Key, string, any) error {
	var count atomic.Int32

	return func(key joseJwk.Key, name string, value any) error {
		if int(count.Add(1)) == failOn {
			return errInjected
		}

		return key.Set(name, value)
	}
}

// countingImport returns a jwkImport that fails on the Nth call.
func countingImport(failOn int) func(any) (joseJwk.Key, error) {
	var count atomic.Int32

	return func(raw any) (joseJwk.Key, error) {
		if int(count.Add(1)) == failOn {
			return nil, errInjected
		}

		return joseJwk.Import(raw)
	}
}

// countingMarshal returns a jsonMarshalFunc that fails on the Nth call.
func countingMarshal(failOn int) func(any) ([]byte, error) {
	var count atomic.Int32

	return func(v any) ([]byte, error) {
		if int(count.Add(1)) == failOn {
			return nil, errInjected
		}

		return json.Marshal(v)
	}
}

// TestCreateJWEJWKFromKey_SeamErrors tests all error injection paths.
func TestCreateJWEJWKFromKey_SeamErrors(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := EncA256GCM
	aesAlg := AlgA256KW
	rsaAlg := AlgRSAOAEP256
	ecAlg := AlgECDHES

	aesKey := make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(aesKey)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	ecdhPriv, err := ecdh.P521().GenerateKey(crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     cryptoutilSharedCryptoKeygen.Key
		enc     *joseJwa.ContentEncryptionAlgorithm
		alg     *joseJwa.KeyEncryptionAlgorithm
		maxSets int
	}{
		{testKeygenAESLabel, cryptoutilSharedCryptoKeygen.SecretKey(aesKey), &enc, &aesAlg, cryptoutilSharedMagic.GitRecentActivityDays},
		{cryptoutilSharedMagic.KeyTypeRSA, &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &enc, &rsaAlg, cryptoutilSharedMagic.IMMinPasswordLength},
		{testEC, &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecdhPriv, Public: ecdhPriv.PublicKey()}, &enc, &ecAlg, cryptoutilSharedMagic.IMMinPasswordLength},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.importJWK = countingImport(1)
			_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, tc.enc, tc.alg, tc.key, operations)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors â€” fail on each Set call number.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				t.Parallel()

				operations := newJOSEOperations()
				operations.setJWK = countingSet(i)
				_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, tc.enc, tc.alg, tc.key, operations)
				require.Error(t, err)
			})
		}

		// Marshal error (first call).
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.marshalJSON = countingMarshal(1)
			_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, tc.enc, tc.alg, tc.key, operations)
			require.Error(t, err)
		})
	}

	// PublicKey error â€” only for KeyPair types.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.publicKeyJWK = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error â€” second Marshal call for KeyPair.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.marshalJSON = countingMarshal(2)
		_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
	})

	// Set(ops) on public JWK â€” 8th Set call for RSA KeyPair.
	t.Run("RSA/set_public_ops_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.setJWK = countingSet(cryptoutilSharedMagic.IMMinPasswordLength)
		_, _, _, _, _, err := createJWEJWKFromKeyWithOperations(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
	})
}

// TestCreateJWSJWKFromKey_SeamErrors tests all error injection paths for JWS JWK creation.
func TestCreateJWSJWKFromKey_SeamErrors(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	hmacAlg := AlgHS256
	rsaAlg := AlgRS256
	ecAlg := AlgES256
	edAlg := AlgEdDSA

	hmacKey := make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(hmacKey)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	ecPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     cryptoutilSharedCryptoKeygen.Key
		alg     *joseJwa.SignatureAlgorithm
		maxSets int
	}{
		{testKeygenHMACLabel, cryptoutilSharedCryptoKeygen.SecretKey(hmacKey), &hmacAlg, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		{cryptoutilSharedMagic.KeyTypeRSA, &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &rsaAlg, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{testEC, &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecPriv, Public: &ecPriv.PublicKey}, &ecAlg, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{cryptoutilSharedMagic.JoseAlgEdDSA, &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}, &edAlg, cryptoutilSharedMagic.DefaultEmailOTPLength},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.importJWK = countingImport(1)
			_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				t.Parallel()

				operations := newJOSEOperations()
				operations.setJWK = countingSet(i)
				_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
				require.Error(t, err)
			})
		}

		// Marshal error.
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.marshalJSON = countingMarshal(1)
			_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
			require.Error(t, err)
		})
	}

	// PublicKey error.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.publicKeyJWK = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.marshalJSON = countingMarshal(2)
		_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
	})

	// Set(ops) on public key â€” 7th Set call for RSA.
	t.Run("RSA/set_public_ops_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.setJWK = countingSet(cryptoutilSharedMagic.GitRecentActivityDays)
		_, _, _, _, _, err := createJWSJWKFromKeyWithOperations(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
	})
}

// TestCreateJWKFromKey_SeamErrors tests all error injection paths for generic JWK creation.
func TestCreateJWKFromKey_SeamErrors(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	hmacAlg := cryptoutilOpenapiModel.Oct256
	hmac384Alg := cryptoutilOpenapiModel.Oct384
	hmac512Alg := cryptoutilOpenapiModel.Oct512
	aes128Alg := cryptoutilOpenapiModel.Oct128
	aes192Alg := cryptoutilOpenapiModel.Oct192
	rsaAlg := cryptoutilOpenapiModel.RSA2048
	ecAlg := cryptoutilOpenapiModel.ECP256
	edAlg := cryptoutilOpenapiModel.OKPEd25519

	hmacKey := make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
	_, _ = crand.Read(hmacKey)

	hmac384Key := make([]byte, cryptoutilSharedMagic.HMACSHA384KeySize)
	_, _ = crand.Read(hmac384Key)

	hmac512Key := make([]byte, cryptoutilSharedMagic.MinSerialNumberBits)
	_, _ = crand.Read(hmac512Key)

	aes128Key := make([]byte, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
	_, _ = crand.Read(aes128Key)

	aes192Key := make([]byte, cryptoutilSharedMagic.HoursPerDay)
	_, _ = crand.Read(aes192Key)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	ecPriv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	tests := []struct {
		name    string
		key     cryptoutilSharedCryptoKeygen.Key
		alg     *cryptoutilOpenapiModel.GenerateAlgorithm
		maxSets int
	}{
		{testHmac256, cryptoutilSharedCryptoKeygen.SecretKey(hmacKey), &hmacAlg, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{testHmac384, cryptoutilSharedCryptoKeygen.SecretKey(hmac384Key), &hmac384Alg, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{testHmac512, cryptoutilSharedCryptoKeygen.SecretKey(hmac512Key), &hmac512Alg, cryptoutilSharedMagic.DefaultEmailOTPLength},
		{testAes128, cryptoutilSharedCryptoKeygen.SecretKey(aes128Key), &aes128Alg, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		{testAes192, cryptoutilSharedCryptoKeygen.SecretKey(aes192Key), &aes192Alg, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries},
		{cryptoutilSharedMagic.KeyTypeRSA, &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &rsaAlg, test3},
		{testEC, &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecPriv, Public: &ecPriv.PublicKey}, &ecAlg, test3},
		{cryptoutilSharedMagic.JoseAlgEdDSA, &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}, &edAlg, test3},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.importJWK = countingImport(1)
			_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				t.Parallel()

				operations := newJOSEOperations()
				operations.setJWK = countingSet(i)
				_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
				require.Error(t, err)
			})
		}

		// Marshal error.
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			t.Parallel()

			operations := newJOSEOperations()
			operations.marshalJSON = countingMarshal(1)
			_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, tc.alg, tc.key, operations)
			require.Error(t, err)
		})
	}

	// PublicKey error.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.publicKeyJWK = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		t.Parallel()

		operations := newJOSEOperations()
		operations.marshalJSON = countingMarshal(2)
		_, _, _, _, _, err := createJWKFromKeyWithOperations(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, operations)
		require.Error(t, err)
	})
}
