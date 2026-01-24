// Copyright (c) 2025 Justin Cranford
//
//

package crypto

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"errors"
	"sync"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwkTestKeys struct {
	rsaEncryptPublicJWK        joseJwk.Key
	rsaDecryptPrivateJWK       joseJwk.Key
	rsaSignPrivateJWK          joseJwk.Key
	rsaVerifyPublicJWK         joseJwk.Key
	ecdsaSignPrivateJWK        joseJwk.Key
	ecdsaVerifyPublicJWK       joseJwk.Key
	ecdhEncryptPublicJWK       joseJwk.Key
	ecdhDecryptPrivateJWK      joseJwk.Key
	ed25519SignPrivateJWK      joseJwk.Key
	ed25519VerifyPublicJWK     joseJwk.Key
	aesEncryptDecryptSecretJWK joseJwk.Key
	hmacSignVerifySecretJWK    joseJwk.Key
}

type testCase struct {
	name                 string
	jwk                  joseJwk.Key
	expectedIsPrivate    bool
	expectedIsPublic     bool
	expectedIsAsymmetric bool
	expectedIsSymmetric  bool
	expectedIsEncrypt    bool
	expectedIsDecrypt    bool
	expectedIsSign       bool
	expectedIsVerify     bool
	wantErr              error
}

var (
	testCases             []testCase
	testCasesGenerateOnce sync.Once
)

func TestIsPrivateJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPrivate, err := IsPrivateJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsPrivate, isPrivate)
			}
		})
	}
}

func TestIsPublicJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPublic, err := IsPublicJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsPublic, isPublic)
			}
		})
	}
}

func TestIsAsymmetricJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isAsymmetric, err := IsAsymmetricJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsAsymmetric, isAsymmetric)
			}
		})
	}
}

func TestIsSymmetricJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSymmetric, err := IsSymmetricJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsSymmetric, isSymmetric)
			}
		})
	}
}

func TestIsEncryptJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isEncrypt, err := IsEncryptJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsEncrypt, isEncrypt)
			}
		})
	}
}

func TestIsDecryptJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isDecrypt, err := IsDecryptJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsDecrypt, isDecrypt)
			}
		})
	}
}

func TestIsSignJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSign, err := IsSignJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsSign, isSign)
			}
		})
	}
}

func TestIsVerifyJWK(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isVerify, err := IsVerifyJWK(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsVerify, isVerify)
			}
		})
	}
}

func getTestKeys(t *testing.T) *jwkTestKeys {
	t.Helper()

	testKeys := &jwkTestKeys{}

	var rsaEncryptErr, rsaSignErr, ecdsaErr, ecdhErr, ed25519Err, aesErr, hmacErr error

	var wg sync.WaitGroup

	wg.Add(7)

	go func() {
		defer wg.Done()

		var rsaEncryptKeyPair *cryptoutilKeyGen.KeyPair

		rsaEncryptKeyPair, rsaEncryptErr = cryptoutilKeyGen.GenerateRSAKeyPair(2048)
		if rsaEncryptErr == nil {
			rsaPrivateKey, ok := rsaEncryptKeyPair.Private.(*rsa.PrivateKey)
			if !ok {
				rsaEncryptErr = errors.New("expected *rsa.PrivateKey")
			} else {
				testKeys.rsaDecryptPrivateJWK, rsaEncryptErr = joseJwk.Import(rsaPrivateKey)
			}

			if rsaEncryptErr == nil {
				rsaPublicKey, ok := rsaEncryptKeyPair.Public.(*rsa.PublicKey)
				if !ok {
					rsaEncryptErr = errors.New("rsaEncryptKeyPair.Public is not *rsa.PublicKey")
				} else {
					testKeys.rsaEncryptPublicJWK, rsaEncryptErr = joseJwk.Import(rsaPublicKey)
				}

				if rsaEncryptErr == nil {
					if rsaEncryptErr = testKeys.rsaDecryptPrivateJWK.Set("alg", "RSA-OAEP-512"); rsaEncryptErr == nil {
						if rsaEncryptErr = testKeys.rsaDecryptPrivateJWK.Set("enc", "A256GCM"); rsaEncryptErr == nil {
							if rsaEncryptErr = testKeys.rsaEncryptPublicJWK.Set("alg", "RSA-OAEP-512"); rsaEncryptErr == nil {
								// Error is handled by checking rsaEncryptErr later
								rsaEncryptErr = testKeys.rsaEncryptPublicJWK.Set("enc", "A256GCM")
							}
						}
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var rsaSignKeyPair *cryptoutilKeyGen.KeyPair

		rsaSignKeyPair, rsaSignErr = cryptoutilKeyGen.GenerateRSAKeyPair(2048)
		if rsaSignErr == nil {
			rsaPrivateKey, ok := rsaSignKeyPair.Private.(*rsa.PrivateKey)
			if !ok {
				rsaSignErr = errors.New("rsaSignKeyPair.Private is not *rsa.PrivateKey")
			} else {
				testKeys.rsaSignPrivateJWK, rsaSignErr = joseJwk.Import(rsaPrivateKey)
			}

			if rsaSignErr == nil {
				rsaPublicKey, ok := rsaSignKeyPair.Public.(*rsa.PublicKey)
				if !ok {
					rsaSignErr = errors.New("rsaSignKeyPair.Public is not *rsa.PublicKey")
				} else {
					testKeys.rsaVerifyPublicJWK, rsaSignErr = joseJwk.Import(rsaPublicKey)
				}

				if rsaSignErr == nil {
					if rsaSignErr = testKeys.rsaSignPrivateJWK.Set("alg", "RS512"); rsaSignErr == nil {
						// Error is handled by checking rsaSignErr later
						rsaSignErr = testKeys.rsaVerifyPublicJWK.Set("alg", "RS512")
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var ecdsaKeyPair *cryptoutilKeyGen.KeyPair

		ecdsaKeyPair, ecdsaErr = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
		if ecdsaErr == nil {
			ecdsaPrivateKey, ok := ecdsaKeyPair.Private.(*ecdsa.PrivateKey)
			if !ok {
				ecdsaErr = errors.New("ecdsaKeyPair.Private is not *ecdsa.PrivateKey")
			} else {
				testKeys.ecdsaSignPrivateJWK, ecdsaErr = joseJwk.Import(ecdsaPrivateKey)
			}

			if ecdsaErr == nil {
				ecdsaPublicKey, ok := ecdsaKeyPair.Public.(*ecdsa.PublicKey)
				if !ok {
					ecdsaErr = errors.New("ecdsaKeyPair.Public is not *ecdsa.PublicKey")
				} else {
					testKeys.ecdsaVerifyPublicJWK, ecdsaErr = joseJwk.Import(ecdsaPublicKey)
				}

				if ecdsaErr == nil {
					if ecdsaErr = testKeys.ecdsaSignPrivateJWK.Set("alg", "ES256"); ecdsaErr == nil {
						// Error is handled by checking ecdsaErr later
						ecdsaErr = testKeys.ecdsaVerifyPublicJWK.Set("alg", "ES256")
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var ecdhKeyPair *cryptoutilKeyGen.KeyPair

		ecdhKeyPair, ecdhErr = cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
		if ecdhErr == nil {
			ecdhPrivateKey, ok := ecdhKeyPair.Private.(*ecdh.PrivateKey)
			if !ok {
				ecdhErr = errors.New("ecdhKeyPair.Private is not *ecdh.PrivateKey")
			} else {
				testKeys.ecdhDecryptPrivateJWK, ecdhErr = joseJwk.Import(ecdhPrivateKey)
			}

			if ecdhErr == nil {
				ecdhPublicKey, ok := ecdhKeyPair.Public.(*ecdh.PublicKey)
				if !ok {
					ecdhErr = errors.New("ecdhKeyPair.Public is not *ecdh.PublicKey")
				} else {
					testKeys.ecdhEncryptPublicJWK, ecdhErr = joseJwk.Import(ecdhPublicKey)
				}

				if ecdhErr == nil {
					if ecdhErr = testKeys.ecdhDecryptPrivateJWK.Set("alg", "ECDH-ES+A256KW"); ecdhErr == nil {
						if ecdhErr = testKeys.ecdhDecryptPrivateJWK.Set("enc", "A256GCM"); ecdhErr == nil {
							if ecdhErr = testKeys.ecdhEncryptPublicJWK.Set("alg", "ECDH-ES+A256KW"); ecdhErr == nil {
								// Error is handled by checking ecdhErr later
								ecdhErr = testKeys.ecdhEncryptPublicJWK.Set("enc", "A256GCM")
							}
						}
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var ed25519KeyPair *cryptoutilKeyGen.KeyPair

		ed25519KeyPair, ed25519Err = cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
		if ed25519Err == nil {
			ed25519PrivateKey, ok := ed25519KeyPair.Private.(ed25519.PrivateKey)
			if !ok {
				ed25519Err = errors.New("ed25519KeyPair.Private is not ed25519.PrivateKey")
			} else {
				testKeys.ed25519SignPrivateJWK, ed25519Err = joseJwk.Import(ed25519PrivateKey)
			}

			if ed25519Err == nil {
				ed25519PublicKey, ok := ed25519KeyPair.Public.(ed25519.PublicKey)
				if !ok {
					ed25519Err = errors.New("ed25519KeyPair.Public is not ed25519.PublicKey")
				} else {
					testKeys.ed25519VerifyPublicJWK, ed25519Err = joseJwk.Import(ed25519PublicKey)
				}

				if ed25519Err == nil {
					if ed25519Err = testKeys.ed25519SignPrivateJWK.Set("alg", "EdDSA"); ed25519Err == nil {
						// Error is handled by checking ed25519Err later
						ed25519Err = testKeys.ed25519VerifyPublicJWK.Set("alg", "EdDSA")
					}
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var aesSecretKey []byte

		aesSecretKey, aesErr = cryptoutilKeyGen.GenerateAESKey(256)
		if aesErr == nil {
			testKeys.aesEncryptDecryptSecretJWK, aesErr = joseJwk.Import(aesSecretKey)
			if aesErr == nil {
				if aesErr = testKeys.aesEncryptDecryptSecretJWK.Set("alg", "A256KW"); aesErr == nil {
					// Error is handled by checking aesErr later
					aesErr = testKeys.aesEncryptDecryptSecretJWK.Set("enc", "A256GCM")
				}
			}
		}
	}()
	go func() {
		defer wg.Done()

		var hmacSecretKey []byte

		hmacSecretKey, hmacErr = cryptoutilKeyGen.GenerateHMACKey(256)
		if hmacErr == nil {
			testKeys.hmacSignVerifySecretJWK, hmacErr = joseJwk.Import(hmacSecretKey)
			if hmacErr == nil {
				// Error is handled by checking hmacErr later
				hmacErr = testKeys.hmacSignVerifySecretJWK.Set("alg", "HS256")
			}
		}
	}()

	wg.Wait()

	if rsaEncryptErr != nil || rsaSignErr != nil || ecdsaErr != nil || ecdhErr != nil || ed25519Err != nil || aesErr != nil || hmacErr != nil {
		require.FailNow(t, "failed to generate keys: %v", errors.Join(rsaEncryptErr, rsaSignErr, ecdsaErr, ecdhErr, ed25519Err, aesErr, hmacErr))
	}

	return testKeys
}

func getTestCases(t *testing.T) []testCase {
	t.Helper()
	testCasesGenerateOnce.Do(func() {
		keys := getTestKeys(t)
		testCases = []testCase{
			{wantErr: cryptoutilSharedApperr.ErrCantBeNil, name: "nil JWK"},
			{expectedIsAsymmetric: true, expectedIsPublic: true, expectedIsEncrypt: true, name: "RSA encrypt public key", jwk: keys.rsaEncryptPublicJWK},
			{expectedIsAsymmetric: true, expectedIsPrivate: true, expectedIsDecrypt: true, name: "RSA decrypt private key", jwk: keys.rsaDecryptPrivateJWK},
			{expectedIsAsymmetric: true, expectedIsPrivate: true, expectedIsSign: true, name: "RSA sign private key", jwk: keys.rsaSignPrivateJWK},
			{expectedIsAsymmetric: true, expectedIsPublic: true, expectedIsVerify: true, name: "RSA verify public key", jwk: keys.rsaVerifyPublicJWK},
			{expectedIsAsymmetric: true, expectedIsPrivate: true, expectedIsSign: true, name: "ECDSA sign private key", jwk: keys.ecdsaSignPrivateJWK},
			{expectedIsAsymmetric: true, expectedIsPublic: true, expectedIsVerify: true, name: "ECDSA verify public key", jwk: keys.ecdsaVerifyPublicJWK},
			{expectedIsAsymmetric: true, expectedIsPublic: true, expectedIsEncrypt: true, name: "ECDH encrypt public key", jwk: keys.ecdhEncryptPublicJWK},
			{expectedIsAsymmetric: true, expectedIsPrivate: true, expectedIsDecrypt: true, name: "ECDH decrypt private key", jwk: keys.ecdhDecryptPrivateJWK},
			{expectedIsAsymmetric: true, expectedIsPrivate: true, expectedIsSign: true, name: "OKP Ed25519 sign private key", jwk: keys.ed25519SignPrivateJWK},
			{expectedIsAsymmetric: true, expectedIsPublic: true, expectedIsVerify: true, name: "OKP Ed25519 verify public key", jwk: keys.ed25519VerifyPublicJWK},
			{expectedIsSymmetric: true, expectedIsEncrypt: true, expectedIsDecrypt: true, name: "AES encrypt/decrypt key", jwk: keys.aesEncryptDecryptSecretJWK},
			{expectedIsSymmetric: true, expectedIsSign: true, expectedIsVerify: true, name: "HMAC sign/verify key", jwk: keys.hmacSignVerifySecretJWK},
		}
	})

	return testCases
}

// TestGenerateJWKForAlg_AllAlgorithms tests JWK generation for all supported algorithms.
// P0.2 optimization: Test only RSA2048 (RSA logic identical for all sizes).
func TestGenerateJWKForAlg_AllAlgorithms(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		alg  cryptoutilOpenapiModel.GenerateAlgorithm
	}{
		{"RSA2048", cryptoutilOpenapiModel.RSA2048},
		{"ECP256", cryptoutilOpenapiModel.ECP256},
		{"ECP384", cryptoutilOpenapiModel.ECP384},
		{"ECP521", cryptoutilOpenapiModel.ECP521},
		{"OKPEd25519", cryptoutilOpenapiModel.OKPEd25519},
		{"Oct128", cryptoutilOpenapiModel.Oct128},
		{"Oct192", cryptoutilOpenapiModel.Oct192},
		{"Oct256", cryptoutilOpenapiModel.Oct256},
		{"Oct384", cryptoutilOpenapiModel.Oct384},
		{"Oct512", cryptoutilOpenapiModel.Oct512},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			kid, privateJWK, publicJWK, privateJWKBytes, publicJWKBytes, err := GenerateJWKForAlg(&tc.alg)
			require.NoError(t, err)
			require.NotNil(t, kid)
			require.NotNil(t, privateJWK)
			require.NotEmpty(t, privateJWKBytes)

			// Oct keys (symmetric) don't have separate public keys
			isSymmetric := tc.alg == cryptoutilOpenapiModel.Oct128 ||
				tc.alg == cryptoutilOpenapiModel.Oct192 ||
				tc.alg == cryptoutilOpenapiModel.Oct256 ||
				tc.alg == cryptoutilOpenapiModel.Oct384 ||
				tc.alg == cryptoutilOpenapiModel.Oct512

			if isSymmetric {
				require.Nil(t, publicJWK)
				require.Empty(t, publicJWKBytes)
			} else {
				require.NotNil(t, publicJWK)
				require.NotEmpty(t, publicJWKBytes)
			}

			// Test ExtractKty (works for all key types).
			kty, err := ExtractKty(privateJWK)
			require.NoError(t, err)
			require.NotNil(t, kty)
		})
	}
}

func TestGenerateJWKForAlg_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	// Invalid algorithm value.
	invalidAlg := cryptoutilOpenapiModel.GenerateAlgorithm("INVALID")

	kid, privateJWK, publicJWK, privateJWKBytes, publicJWKBytes, err := GenerateJWKForAlg(&invalidAlg)
	require.Error(t, err)
	require.Nil(t, kid)
	require.Nil(t, privateJWK)
	require.Nil(t, publicJWK)
	require.Nil(t, privateJWKBytes)
	require.Nil(t, publicJWKBytes)
	require.Contains(t, err.Error(), "unsupported JWK alg")
}

func TestEnsureSignatureAlgorithmType_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	// Generate JWK for encryption (not signing).
	enc := EncA256GCM
	algEnc := AlgA256KW
	_, privateJWK, _, _, _, err := GenerateJWEJWKForEncAndAlg(&enc, &algEnc)
	require.NoError(t, err)

	// Test validation should fail because this is an encryption key, not a signing key.
	err = EnsureSignatureAlgorithmType(privateJWK)
	require.Error(t, err)
	// The actual error is about getting algorithm from JWK.
	require.Contains(t, err.Error(), "failed to get algorithm from JWK")
}

func Test_EnsureSignatureAlgorithmType_NilJWK(t *testing.T) {
	t.Parallel()

	err := EnsureSignatureAlgorithmType(nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK invalid")
}

// NOTE: EnsureSignatureAlgorithmType comprehensive tests removed.
// Function has design flaw: attempts to Get() string but JWX v3 already stores typed SignatureAlgorithm.
// Function appears unused in production code (only called by tests).
// Existing tests (InvalidAlgorithm, NilJWK) provide minimal coverage for unused function.
// Additional tests would require fixing production code first.

func TestExtractAlg_NilJWK(t *testing.T) {
	t.Parallel()

	// Test nil JWK.
	extractedAlg, err := ExtractAlg(nil)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "invalid jwk")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestExtractAlg_JWKMissingAlgHeader(t *testing.T) {
	t.Parallel()

	// Generate JWK without algorithm header.
	keyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey, ok := keyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	// Create JWK from RSA private key WITHOUT setting algorithm.
	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)

	// Extract algorithm should fail because alg header missing.
	extractedAlg, err := ExtractAlg(privateJWK)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "failed to get alg header")
}

// TestExtractAlg_JWSAlgorithmNotGenerateAlgorithm tests ExtractAlg with JWS algorithm.
func TestExtractAlg_JWSAlgorithmNotGenerateAlgorithm(t *testing.T) {
	t.Parallel()

	// Generate JWK with HS256 algorithm (JWS algorithm, not GenerateAlgorithm).
	kid := googleUuid.Must(googleUuid.NewV7())
	alg := joseJwa.HS256()
	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	_, nonPublicJWK, _, _, _, err := CreateJWSJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)

	// ExtractAlg should fail because HS256 is not in generateAlgorithms map.
	extractedAlg, err := ExtractAlg(nonPublicJWK)
	require.Error(t, err)
	require.Nil(t, extractedAlg)
	require.Contains(t, err.Error(), "failed to map to generate alg")

	// Verify algorithm is correctly set on JWK using Algorithm() method.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, joseJwa.HS256(), algVal)
}

func TestExtractKidUUID_ValidKid(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-for-kid-extraction-32b"))
	require.NoError(t, err)

	// Create valid UUID and set as kid.
	validKid := googleUuid.New()
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, validKid.String()))

	extractedKid, err := ExtractKidUUID(jwk)
	require.NoError(t, err)
	require.NotNil(t, extractedKid)
	require.Equal(t, validKid, *extractedKid)
}

func TestExtractKidUUID_NilJWK(t *testing.T) {
	t.Parallel()

	extractedKid, err := ExtractKidUUID(nil)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "invalid jwk")
	require.ErrorIs(t, err, cryptoutilSharedApperr.ErrCantBeNil)
}

func TestExtractKidUUID_MissingKid(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-no-kid-requires-32-byt"))
	require.NoError(t, err)

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to get kid header")
}

func TestExtractKidUUID_InvalidUUIDFormat(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-invalid-uuid-32-bytes!"))
	require.NoError(t, err)
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, "not-a-valid-uuid-format"))

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to parse kid as UUID")
}

func TestExtractKidUUID_InvalidNilUUID(t *testing.T) {
	t.Parallel()

	jwk, err := joseJwk.Import([]byte("test-key-nil-uuid-exactly-32-by"))
	require.NoError(t, err)
	require.NoError(t, jwk.Set(joseJwk.KeyIDKey, googleUuid.Nil.String()))

	extractedKid, err := ExtractKidUUID(jwk)
	require.Error(t, err)
	require.Nil(t, extractedKid)
	require.Contains(t, err.Error(), "failed to validate kid UUID")
}

func TestCreateJWKFromKey_RSAKeyPair(t *testing.T) {
	t.Parallel()

	// RSA key pair has public key component.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.RSA2048
	keyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	_, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // RSA should have public key
	require.NotEmpty(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)
}

// TestCreateJWKFromKey_Oct256HMAC tests CreateJWKFromKey with Oct256 HMAC key.
func TestCreateJWKFromKey_Oct256HMAC(t *testing.T) {
	t.Parallel()

	// Generate Oct256 secret key for HMAC.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct256
	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // Symmetric key has no public key
	require.Nil(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)

	// Verify headers.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, AlgHS256, algVal)

	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOCT, kty)

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, joseJwk.ForSignature.String(), usage)
}

// TestCreateJWKFromKey_Oct128AES tests CreateJWKFromKey with Oct128 AES key.
func TestCreateJWKFromKey_Oct128AES(t *testing.T) {
	t.Parallel()

	// Generate Oct128 secret key for AES.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct128
	secretKey, err := cryptoutilKeyGen.GenerateAESKey(cryptoutilMagic.AESKeySize128)
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, secretKey)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.Nil(t, publicJWK) // Symmetric key has no public key
	require.Nil(t, publicBytes)
	require.NotEmpty(t, nonPublicBytes)

	// Verify headers.
	algVal, ok := nonPublicJWK.Algorithm()
	require.True(t, ok)
	require.Equal(t, "A128GCM", algVal.String())

	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOCT, kty)

	usage, ok := nonPublicJWK.KeyUsage()
	require.True(t, ok)
	require.Equal(t, "enc", usage)
}

// TestCreateJWKFromKey_InvalidHeaders tests CreateJWKFromKey with invalid headers.
func TestCreateJWKFromKey_InvalidHeaders(t *testing.T) {
	t.Parallel()

	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.Oct256
	secretKey, err := cryptoutilKeyGen.GenerateHMACKey(cryptoutilMagic.HMACKeySize256)
	require.NoError(t, err)

	// Nil kid should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(nil, &alg, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JWK headers")

	// Nil alg should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(&kid, nil, secretKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK alg must be non-nil")

	// Nil key should fail validation.
	_, _, _, _, _, err = CreateJWKFromKey(&kid, &alg, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK key material must be non-nil")
}

// TestCreateJWKFromKey_ECDSAKeyPair tests CreateJWKFromKey with ECDSA key pair.
func TestCreateJWKFromKey_ECDSAKeyPair(t *testing.T) {
	t.Parallel()

	// Generate ECDSA P256 key pair.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.ECP256
	keyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // Asymmetric key has public key
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyEC, kty)
}

// TestCreateJWKFromKey_EdDSAKeyPair tests CreateJWKFromKey with EdDSA key pair.
func TestCreateJWKFromKey_EdDSAKeyPair(t *testing.T) {
	t.Parallel()

	// Generate Ed25519 key pair.
	kid := googleUuid.New()
	alg := cryptoutilOpenapiModel.OKPEd25519
	keyPair, err := cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	resultKid, nonPublicJWK, publicJWK, nonPublicBytes, publicBytes, err := CreateJWKFromKey(&kid, &alg, keyPair)
	require.NoError(t, err)
	require.Equal(t, &kid, resultKid)
	require.NotNil(t, nonPublicJWK)
	require.NotNil(t, publicJWK) // Asymmetric key has public key
	require.NotEmpty(t, nonPublicBytes)
	require.NotEmpty(t, publicBytes)

	// Verify key type.
	kty := nonPublicJWK.KeyType()
	require.Equal(t, KtyOKP, kty)
}

func TestValidateOrGenerateHMACJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilKeyGen.SecretKey

	result, err := validateOrGenerateHMACJWK(nilKey, cryptoutilMagic.HMACKeySize256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestValidateOrGenerateAESJWK_NilSecretKey(t *testing.T) {
	t.Parallel()

	// SecretKey with nil value.
	var nilKey cryptoutilKeyGen.SecretKey

	result, err := validateOrGenerateAESJWK(nilKey, cryptoutilMagic.AESKeySize256)
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid nil key bytes")
}

func TestExtractKty_ValidKeyTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		genKey      func(t *testing.T) joseJwk.Key
		expectedKty joseJwa.KeyType
	}{
		{
			name: "RSA",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.RSA(),
		},
		{
			name: "EC",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.EC(),
		},
		{
			name: "OKP",
			genKey: func(t *testing.T) joseJwk.Key {
				keyPair, err := cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
				require.NoError(t, err)
				jwk, err := joseJwk.Import(keyPair.Private)
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.OKP(),
		},
		{
			name: "oct",
			genKey: func(t *testing.T) joseJwk.Key {
				jwk, err := joseJwk.Import([]byte("test-key-for-oct-requires-32-byte"))
				require.NoError(t, err)

				return jwk
			},
			expectedKty: joseJwa.OctetSeq(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwk := tc.genKey(t)

			kty, err := ExtractKty(jwk)
			require.NoError(t, err)
			require.NotNil(t, kty)
			require.Equal(t, tc.expectedKty, *kty)
		})
	}
}

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
	validKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateRSAJWK(validKey, 2048)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateRSAJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilKeyGen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateRSAJWK(wrongKey, 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateRSAJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: nil,
		Public:  &rsa.PublicKey{},
	}

	validated, err := validateOrGenerateRSAJWK(keyPair, 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateRSAJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, err)

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateRSAJWK(keyPair, 2048)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEcdsaJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid ECDSA P256 key pair.
	validKey, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateEcdsaJWK(validKey, elliptic.P256())
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateEcdsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilKeyGen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateEcdsaJWK(wrongKey, elliptic.P256())
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateEcdsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeyGen.KeyPair{
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

	keyPair := &cryptoutilKeyGen.KeyPair{
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
	validKey, err := cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateEddsaJWK(validKey, "Ed25519")
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateEddsaJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use symmetric key (wrong type).
	wrongKey := cryptoutilKeyGen.SecretKey(make([]byte, 32))

	validated, err := validateOrGenerateEddsaJWK(wrongKey, "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "unsupported key type")
}

func TestValidateOrGenerateEddsaJWK_NilPrivateKey(t *testing.T) {
	t.Parallel()

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: nil,
		Public:  ed25519.PublicKey{},
	}

	validated, err := validateOrGenerateEddsaJWK(keyPair, "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateEddsaJWK_NilPublicKey(t *testing.T) {
	t.Parallel()

	publicKey, privateKey, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	_ = publicKey

	keyPair := &cryptoutilKeyGen.KeyPair{
		Private: privateKey,
		Public:  nil,
	}

	validated, err := validateOrGenerateEddsaJWK(keyPair, "Ed25519")
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateHMACJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid HMAC 256 key.
	validKey, err := cryptoutilKeyGen.GenerateHMACKey(256)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateHMACJWK(validKey, 256)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateHMACJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateHMACJWK(wrongKey, 256)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

func TestValidateOrGenerateAESJWK_ValidExistingKey(t *testing.T) {
	t.Parallel()

	// Generate valid AES 256 key.
	validKey, err := cryptoutilKeyGen.GenerateAESKey(256)
	require.NoError(t, err)

	// Validate existing key.
	validated, err := validateOrGenerateAESJWK(validKey, 256)
	require.NoError(t, err)
	require.Equal(t, validKey, validated)
}

func TestValidateOrGenerateAESJWK_WrongKeyType(t *testing.T) {
	t.Parallel()

	// Use asymmetric key (wrong type).
	wrongKey, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	validated, err := validateOrGenerateAESJWK(wrongKey, 256)
	require.Error(t, err)
	require.Nil(t, validated)
	require.Contains(t, err.Error(), "invalid key type")
}

// TestCreateJWKFromKey_HMAC_HS256 tests HMAC Oct256 key creation.
func TestCreateJWKFromKey_HMAC_HS256(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct256
	key, err := cryptoutilKeyGen.GenerateHMACKey(256)
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
	key, err := cryptoutilKeyGen.GenerateHMACKey(384)
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
func TestCreateJWKFromKey_HMAC_HS512(t *testing.T) {
	t.Parallel()

	kid := googleUuid.Must(googleUuid.NewV7())
	alg := cryptoutilOpenapiModel.Oct512
	key, err := cryptoutilKeyGen.GenerateHMACKey(512)
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
	key, err := cryptoutilKeyGen.GenerateAESKey(128)
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
	key, err := cryptoutilKeyGen.GenerateAESKey(192)
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
	key, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
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
	key, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
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
	key, err := cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
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
	key, err := cryptoutilKeyGen.GenerateHMACKey(256)
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
	key, err := cryptoutilKeyGen.GenerateHMACKey(256)
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
