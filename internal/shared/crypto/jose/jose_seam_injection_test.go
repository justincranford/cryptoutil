// Copyright (c) 2025 Justin Cranford
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

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// errInjected is the sentinel error used in all error injection tests.
var errInjected = errors.New("injected error")

// saveRestoreSeams saves current seam values and restores them on test cleanup.
func saveRestoreSeams(t *testing.T) {
	t.Helper()

	origKeySet := jwkKeySet
	origImport := jwkImport
	origMarshal := jsonMarshalFunc
	origPublicKey := jwkPublicKey
	origGenRSA := generateRSAKeyPair
	origGenECDSA := generateECDSAKeyPair
	origGenEDDSA := generateEDDSAKeyPair
	origGenHMAC := generateHMACKey
	origGenAES := generateAESKey

	t.Cleanup(func() {
		jwkKeySet = origKeySet
		jwkImport = origImport
		jsonMarshalFunc = origMarshal
		jwkPublicKey = origPublicKey
		generateRSAKeyPair = origGenRSA
		generateECDSAKeyPair = origGenECDSA
		generateEDDSAKeyPair = origGenEDDSA
		generateHMACKey = origGenHMAC
		generateAESKey = origGenAES
	})
}

// resetSeams restores all seams to their default (real) implementations.
func resetSeams() {
	jwkKeySet = func(key joseJwk.Key, name string, value any) error { return key.Set(name, value) }
	jwkImport = func(raw any) (joseJwk.Key, error) { return joseJwk.Import(raw) }
	jsonMarshalFunc = json.Marshal
	jwkPublicKey = func(key joseJwk.Key) (joseJwk.Key, error) { return key.PublicKey() }
	generateRSAKeyPair = cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair
	generateECDSAKeyPair = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair
	generateEDDSAKeyPair = cryptoutilSharedCryptoKeygen.GenerateEDDSAKeyPair
	generateHMACKey = cryptoutilSharedCryptoKeygen.GenerateHMACKey
	generateAESKey = cryptoutilSharedCryptoKeygen.GenerateAESKey
}

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
	// NOT PARALLEL: modifies package-level function variables.
	saveRestoreSeams(t)

	kid := googleUuid.Must(googleUuid.NewV7())
	enc := EncA256GCM
	aesAlg := AlgA256KW
	rsaAlg := AlgRSAOAEP256
	ecAlg := AlgECDHES

	aesKey := make([]byte, 32)
	_, _ = crand.Read(aesKey)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, 2048)
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
		{"AES", cryptoutilSharedCryptoKeygen.SecretKey(aesKey), &enc, &aesAlg, 7},
		{"RSA", &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &enc, &rsaAlg, 8},
		{"EC", &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecdhPriv, Public: ecdhPriv.PublicKey()}, &enc, &ecAlg, 8},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			jwkImport = countingImport(1)
			_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, tc.enc, tc.alg, tc.key)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors — fail on each Set call number.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				jwkKeySet = countingSet(i)
				_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, tc.enc, tc.alg, tc.key)
				require.Error(t, err)
			})
		}

		// Marshal error (first call).
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			jsonMarshalFunc = countingMarshal(1)
			_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, tc.enc, tc.alg, tc.key)
			require.Error(t, err)
		})
	}

	// PublicKey error — only for KeyPair types.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		resetSeams()

		jwkPublicKey = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error — second Marshal call for KeyPair.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		resetSeams()

		jsonMarshalFunc = countingMarshal(2)
		_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
	})

	// Set(ops) on public JWK — 8th Set call for RSA KeyPair.
	t.Run("RSA/set_public_ops_fail", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(8)
		_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &enc, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
	})
}

// TestCreateJWSJWKFromKey_SeamErrors tests all error injection paths for JWS JWK creation.
func TestCreateJWSJWKFromKey_SeamErrors(t *testing.T) {
	// NOT PARALLEL: modifies package-level function variables.
	saveRestoreSeams(t)

	kid := googleUuid.Must(googleUuid.NewV7())
	hmacAlg := AlgHS256
	rsaAlg := AlgRS256
	ecAlg := AlgES256
	edAlg := AlgEdDSA

	hmacKey := make([]byte, 32)
	_, _ = crand.Read(hmacKey)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, 2048)
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
		{"HMAC", cryptoutilSharedCryptoKeygen.SecretKey(hmacKey), &hmacAlg, 5},
		{"RSA", &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &rsaAlg, 6},
		{"EC", &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecPriv, Public: &ecPriv.PublicKey}, &ecAlg, 6},
		{"EdDSA", &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}, &edAlg, 6},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			jwkImport = countingImport(1)
			_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, tc.alg, tc.key)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				jwkKeySet = countingSet(i)
				_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, tc.alg, tc.key)
				require.Error(t, err)
			})
		}

		// Marshal error.
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			jsonMarshalFunc = countingMarshal(1)
			_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, tc.alg, tc.key)
			require.Error(t, err)
		})
	}

	// PublicKey error.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		resetSeams()

		jwkPublicKey = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		resetSeams()

		jsonMarshalFunc = countingMarshal(2)
		_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
	})

	// Set(ops) on public key — 7th Set call for RSA.
	t.Run("RSA/set_public_ops_fail", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(7)
		_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
	})
}

// TestCreateJWKFromKey_SeamErrors tests all error injection paths for generic JWK creation.
func TestCreateJWKFromKey_SeamErrors(t *testing.T) {
	// NOT PARALLEL: modifies package-level function variables.
	saveRestoreSeams(t)

	kid := googleUuid.Must(googleUuid.NewV7())
	hmacAlg := cryptoutilOpenapiModel.Oct256
	hmac384Alg := cryptoutilOpenapiModel.Oct384
	hmac512Alg := cryptoutilOpenapiModel.Oct512
	aes128Alg := cryptoutilOpenapiModel.Oct128
	aes192Alg := cryptoutilOpenapiModel.Oct192
	rsaAlg := cryptoutilOpenapiModel.RSA2048
	ecAlg := cryptoutilOpenapiModel.ECP256
	edAlg := cryptoutilOpenapiModel.OKPEd25519

	hmacKey := make([]byte, 32)
	_, _ = crand.Read(hmacKey)

	hmac384Key := make([]byte, 48)
	_, _ = crand.Read(hmac384Key)

	hmac512Key := make([]byte, 64)
	_, _ = crand.Read(hmac512Key)

	aes128Key := make([]byte, 16)
	_, _ = crand.Read(aes128Key)

	aes192Key := make([]byte, 24)
	_, _ = crand.Read(aes192Key)

	rsaPriv, err := rsa.GenerateKey(crand.Reader, 2048)
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
		{"HMAC256", cryptoutilSharedCryptoKeygen.SecretKey(hmacKey), &hmacAlg, 6},
		{"HMAC384", cryptoutilSharedCryptoKeygen.SecretKey(hmac384Key), &hmac384Alg, 6},
		{"HMAC512", cryptoutilSharedCryptoKeygen.SecretKey(hmac512Key), &hmac512Alg, 6},
		{"AES128", cryptoutilSharedCryptoKeygen.SecretKey(aes128Key), &aes128Alg, 5},
		{"AES192", cryptoutilSharedCryptoKeygen.SecretKey(aes192Key), &aes192Alg, 5},
		{"RSA", &cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey}, &rsaAlg, 3},
		{"EC", &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecPriv, Public: &ecPriv.PublicKey}, &ecAlg, 3},
		{"EdDSA", &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}, &edAlg, 3},
	}

	for _, tc := range tests {
		// Import error.
		t.Run(tc.name+"/import_fail", func(t *testing.T) {
			jwkImport = countingImport(1)
			_, _, _, _, _, err := CreateJWKFromKey(&kid, tc.alg, tc.key)
			require.Error(t, err)
			require.True(t, errors.Is(err, errInjected))
		})

		// Set errors.
		for i := 1; i <= tc.maxSets; i++ {
			t.Run(fmt.Sprintf("%s/set_fail_%d", tc.name, i), func(t *testing.T) {
				jwkKeySet = countingSet(i)
				_, _, _, _, _, err := CreateJWKFromKey(&kid, tc.alg, tc.key)
				require.Error(t, err)
			})
		}

		// Marshal error.
		t.Run(tc.name+"/marshal_fail_1", func(t *testing.T) {
			jsonMarshalFunc = countingMarshal(1)
			_, _, _, _, _, err := CreateJWKFromKey(&kid, tc.alg, tc.key)
			require.Error(t, err)
		})
	}

	// PublicKey error.
	t.Run("RSA/publickey_fail", func(t *testing.T) {
		resetSeams()

		jwkPublicKey = func(_ joseJwk.Key) (joseJwk.Key, error) { return nil, errInjected }
		_, _, _, _, _, err := CreateJWKFromKey(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	// Marshal(public) error.
	t.Run("RSA/marshal_public_fail", func(t *testing.T) {
		resetSeams()

		jsonMarshalFunc = countingMarshal(2)
		_, _, _, _, _, err := CreateJWKFromKey(&kid, &rsaAlg,
			&cryptoutilSharedCryptoKeygen.KeyPair{Private: rsaPriv, Public: &rsaPriv.PublicKey})
		require.Error(t, err)
	})
}

// TestEdDSASetError tests Set error in the ed25519 type switch branch directly.
func TestEdDSASetError(t *testing.T) {
	saveRestoreSeams(t)

	kid := googleUuid.Must(googleUuid.NewV7())
	edAlg := AlgEdDSA

	_, edPriv, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	edKey := &cryptoutilSharedCryptoKeygen.KeyPair{Private: edPriv, Public: edPriv.Public()}

	// JWS: Set error on 1st Set (kty OKP)
	t.Run("JWS_EdDSA_set1", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWSJWKFromKey(&kid, &edAlg, edKey)
		require.Error(t, err, "expected Set error for JWS EdDSA kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})

	// JWE: Set error on 1st Set (kty OKP) — use ECDH-ES with ecdh key
	ecdhAlg := AlgECDHES
	encAlg := EncA256GCM
	ecdhPrivKey, err := ecdh.P521().GenerateKey(crand.Reader)
	require.NoError(t, err)

	ecdhKey := &cryptoutilSharedCryptoKeygen.KeyPair{Private: ecdhPrivKey, Public: ecdhPrivKey.PublicKey()}

	t.Run("JWE_ECDH_set1", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWEJWKFromKey(&kid, &encAlg, &ecdhAlg, ecdhKey)
		require.Error(t, err, "expected Set error for JWE ECDH kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})

	// JWK: Set error on 1st Set (kty OKP)
	genAlg := cryptoutilOpenapiModel.OKPEd25519

	t.Run("JWK_EdDSA_set1", func(t *testing.T) {
		resetSeams()

		jwkKeySet = countingSet(1)
		_, _, _, _, _, err := CreateJWKFromKey(&kid, &genAlg, edKey)
		require.Error(t, err, "expected Set error for JWK EdDSA kty Set")
		require.True(t, errors.Is(err, errInjected), "error should wrap errInjected, got: %v", err)
	})
}

// TestValidateOrGenerate_KeygenErrors tests keygen error paths in validateOrGenerate* functions.
func TestValidateOrGenerate_KeygenErrors(t *testing.T) {
	// NOT PARALLEL: modifies package-level function variables.
	saveRestoreSeams(t)

	t.Run("RSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateRSAKeyPair = func(_ int) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateRSAJWK(nil, cryptoutilSharedMagic.RSAKeySize2048)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("ECDSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateECDSAKeyPair = func(_ elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateEcdsaJWK(nil, elliptic.P256())
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("EDDSA_keygen_error", func(t *testing.T) {
		resetSeams()

		generateEDDSAKeyPair = func(_ string) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateEddsaJWK(nil, "Ed25519")
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("HMAC_keygen_error", func(t *testing.T) {
		resetSeams()

		generateHMACKey = func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateHMACJWK(nil, cryptoutilSharedMagic.HMACKeySize256)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})

	t.Run("AES_keygen_error", func(t *testing.T) {
		resetSeams()

		generateAESKey = func(_ int) (cryptoutilSharedCryptoKeygen.SecretKey, error) {
			return nil, errInjected
		}
		_, err := validateOrGenerateAESJWK(nil, cryptoutilSharedMagic.AESKeySize256)
		require.Error(t, err)
		require.True(t, errors.Is(err, errInjected))
	})
}
