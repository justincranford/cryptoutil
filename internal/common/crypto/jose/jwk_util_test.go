package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"errors"
	"sync"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

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
			{wantErr: cryptoutilAppErr.ErrCantBeNil, name: "nil JWK"},
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
