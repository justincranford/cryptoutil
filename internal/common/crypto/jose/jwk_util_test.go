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

func TestIsPrivateJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPrivate, err := IsPrivateJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsPrivate, isPrivate)
			}
		})
	}
}

func TestIsPublicJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPublic, err := IsPublicJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsPublic, isPublic)
			}
		})
	}
}

func TestIsAsymmetricJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isAsymmetric, err := IsAsymmetricJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsAsymmetric, isAsymmetric)
			}
		})
	}
}

func TestIsSymmetricJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSymmetric, err := IsSymmetricJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsSymmetric, isSymmetric)
			}
		})
	}
}

func TestIsEncryptJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isEncrypt, err := IsEncryptJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsEncrypt, isEncrypt)
			}
		})
	}
}

func TestIsDecryptJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isDecrypt, err := IsDecryptJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsDecrypt, isDecrypt)
			}
		})
	}
}

func TestIsSignJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSign, err := IsSignJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsSign, isSign)
			}
		})
	}
}

func TestIsVerifyJwk(t *testing.T) {
	tests := getTestCases(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isVerify, err := IsVerifyJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIsVerify, isVerify)
			}
		})
	}
}

func getTestCases(t *testing.T) []testCase {
	t.Helper()
	testCasesGenerateOnce.Do(func() {
		keys := getTestKeys(t)
		testCases = []testCase{
			{
				name:                 "nil JWK",
				jwk:                  nil,
				expectedIsPrivate:    false,
				expectedIsPublic:     false,
				expectedIsAsymmetric: false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              cryptoutilAppErr.ErrCantBeNil,
			},
			{
				name:                 "RSA encrypt public key",
				jwk:                  keys.rsaEncryptPublicJWK,
				expectedIsPublic:     true,
				expectedIsAsymmetric: true,
				expectedIsEncrypt:    true,
				expectedIsPrivate:    false,
				expectedIsSymmetric:  false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "RSA decrypt private key",
				jwk:                  keys.rsaDecryptPrivateJWK,
				expectedIsPrivate:    true,
				expectedIsAsymmetric: true,
				expectedIsDecrypt:    true,
				expectedIsPublic:     false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "RSA sign private key",
				jwk:                  keys.rsaSignPrivateJWK,
				expectedIsPrivate:    true,
				expectedIsAsymmetric: true,
				expectedIsSign:       true,
				expectedIsPublic:     false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "RSA verify public key",
				jwk:                  keys.rsaVerifyPublicJWK,
				expectedIsPublic:     true,
				expectedIsAsymmetric: true,
				expectedIsVerify:     true,
				expectedIsPrivate:    false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				wantErr:              nil,
			},
			{
				name:                 "ECDSA sign private key",
				jwk:                  keys.ecdsaSignPrivateJWK,
				expectedIsPrivate:    true,
				expectedIsAsymmetric: true,
				expectedIsSign:       true,
				expectedIsPublic:     false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "ECDSA verify public key",
				jwk:                  keys.ecdsaVerifyPublicJWK,
				expectedIsPublic:     true,
				expectedIsAsymmetric: true,
				expectedIsVerify:     true,
				expectedIsPrivate:    false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				wantErr:              nil,
			},
			{
				name:                 "ECDH encrypt public key",
				jwk:                  keys.ecdhEncryptPublicJWK,
				expectedIsPublic:     true,
				expectedIsAsymmetric: true,
				expectedIsEncrypt:    true,
				expectedIsPrivate:    false,
				expectedIsSymmetric:  false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "ECDH decrypt private key",
				jwk:                  keys.ecdhDecryptPrivateJWK,
				expectedIsPrivate:    true,
				expectedIsAsymmetric: true,
				expectedIsDecrypt:    true,
				expectedIsPublic:     false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "OKP Ed25519 sign private key",
				jwk:                  keys.ed25519SignPrivateJWK,
				expectedIsPrivate:    true,
				expectedIsAsymmetric: true,
				expectedIsSign:       true,
				expectedIsPublic:     false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "OKP Ed25519 verify public key",
				jwk:                  keys.ed25519VerifyPublicJWK,
				expectedIsPublic:     true,
				expectedIsAsymmetric: true,
				expectedIsVerify:     true,
				expectedIsPrivate:    false,
				expectedIsSymmetric:  false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				expectedIsSign:       false,
				wantErr:              nil,
			},
			{
				name:                 "AES encrypt/decrypt key",
				jwk:                  keys.aesEncryptDecryptSecretJWK,
				expectedIsSymmetric:  true,
				expectedIsEncrypt:    true,
				expectedIsDecrypt:    true,
				expectedIsPrivate:    false,
				expectedIsPublic:     false,
				expectedIsAsymmetric: false,
				expectedIsSign:       false,
				expectedIsVerify:     false,
				wantErr:              nil,
			},
			{
				name:                 "HMAC sign/verify key",
				jwk:                  keys.hmacSignVerifySecretJWK,
				expectedIsSymmetric:  true,
				expectedIsSign:       true,
				expectedIsVerify:     true,
				expectedIsPrivate:    false,
				expectedIsPublic:     false,
				expectedIsAsymmetric: false,
				expectedIsEncrypt:    false,
				expectedIsDecrypt:    false,
				wantErr:              nil,
			},
		}
	})
	return testCases
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
			testKeys.rsaDecryptPrivateJWK, rsaEncryptErr = joseJwk.Import(rsaEncryptKeyPair.Private.(*rsa.PrivateKey))
			if rsaEncryptErr == nil {
				testKeys.rsaEncryptPublicJWK, rsaEncryptErr = joseJwk.Import(rsaEncryptKeyPair.Public.(*rsa.PublicKey))
				if rsaEncryptErr == nil {
					rsaEncryptErr = testKeys.rsaDecryptPrivateJWK.Set("alg", "RSA-OAEP-512")
					if rsaEncryptErr == nil {
						rsaEncryptErr = testKeys.rsaDecryptPrivateJWK.Set("enc", "A256GCM")
						if rsaEncryptErr == nil {
							rsaEncryptErr = testKeys.rsaEncryptPublicJWK.Set("alg", "RSA-OAEP-512")
							if rsaEncryptErr == nil {
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
			testKeys.rsaSignPrivateJWK, rsaSignErr = joseJwk.Import(rsaSignKeyPair.Private.(*rsa.PrivateKey))
			if rsaSignErr == nil {
				testKeys.rsaVerifyPublicJWK, rsaSignErr = joseJwk.Import(rsaSignKeyPair.Public.(*rsa.PublicKey))
				if rsaSignErr == nil {
					rsaSignErr = testKeys.rsaSignPrivateJWK.Set("alg", "RS512")
					if rsaSignErr == nil {
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
			testKeys.ecdsaSignPrivateJWK, ecdsaErr = joseJwk.Import(ecdsaKeyPair.Private.(*ecdsa.PrivateKey))
			if ecdsaErr == nil {
				testKeys.ecdsaVerifyPublicJWK, ecdsaErr = joseJwk.Import(ecdsaKeyPair.Public.(*ecdsa.PublicKey))
				if ecdsaErr == nil {
					ecdsaErr = testKeys.ecdsaSignPrivateJWK.Set("alg", "ES256")
					if ecdsaErr == nil {
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
			testKeys.ecdhDecryptPrivateJWK, ecdhErr = joseJwk.Import(ecdhKeyPair.Private.(*ecdh.PrivateKey))
			if ecdhErr == nil {
				testKeys.ecdhEncryptPublicJWK, ecdhErr = joseJwk.Import(ecdhKeyPair.Public.(*ecdh.PublicKey))
				if ecdhErr == nil {
					ecdhErr = testKeys.ecdhDecryptPrivateJWK.Set("alg", "ECDH-ES+A256KW")
					if ecdhErr == nil {
						ecdhErr = testKeys.ecdhDecryptPrivateJWK.Set("enc", "A256GCM")
						if ecdhErr == nil {
							ecdhErr = testKeys.ecdhEncryptPublicJWK.Set("alg", "ECDH-ES+A256KW")
							if ecdhErr == nil {
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
			testKeys.ed25519SignPrivateJWK, ed25519Err = joseJwk.Import(ed25519KeyPair.Private.(ed25519.PrivateKey))
			if ed25519Err == nil {
				testKeys.ed25519VerifyPublicJWK, ed25519Err = joseJwk.Import(ed25519KeyPair.Public.(ed25519.PublicKey))
				if ed25519Err == nil {
					ed25519Err = testKeys.ed25519SignPrivateJWK.Set("alg", "EdDSA")
					if ed25519Err == nil {
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
				aesErr = testKeys.aesEncryptDecryptSecretJWK.Set("alg", "A256KW")
				if aesErr == nil {
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
				hmacErr = testKeys.hmacSignVerifySecretJWK.Set("alg", "HS256")
			}
		}
	}()
	wg.Wait()
	if rsaEncryptErr != nil || rsaSignErr != nil || ecdsaErr != nil || ecdhErr != nil || ed25519Err != nil || aesErr != nil || hmacErr != nil {
		t.Fatalf("failed to generate keys: %v", errors.Join(rsaEncryptErr, rsaSignErr, ecdsaErr, ecdhErr, ed25519Err, aesErr, hmacErr))
	}
	return testKeys
}
