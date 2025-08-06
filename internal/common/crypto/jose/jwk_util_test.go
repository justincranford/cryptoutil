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
	rsaPrivateJWK     joseJwk.Key
	rsaPublicJWK      joseJwk.Key
	ecdsaPrivateJWK   joseJwk.Key
	ecdsaPublicJWK    joseJwk.Key
	ecdhPrivateJWK    joseJwk.Key
	ecdhPublicJWK     joseJwk.Key
	ed25519PrivateJWK joseJwk.Key
	ed25519PublicJWK  joseJwk.Key
	aesSecretJWK      joseJwk.Key
}

type testCase struct {
	name                string
	jwk                 joseJwk.Key
	expectedIsPrivate   bool
	expectedIsPublic    bool
	expectedIsSymmetric bool
	wantErr             error
}

var (
	testKeys             *jwkTestKeys
	testKeysGenerateOnce sync.Once
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

func getTestCases(t *testing.T) []testCase {
	t.Helper()
	keys := getTestKeys(t)
	return []testCase{
		{
			name:                "nil JWK",
			jwk:                 nil,
			expectedIsPrivate:   false,
			expectedIsPublic:    false,
			expectedIsSymmetric: false,
			wantErr:             cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:                "RSA private key",
			jwk:                 keys.rsaPrivateJWK,
			expectedIsPrivate:   true,
			expectedIsPublic:    false,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "RSA public key",
			jwk:                 keys.rsaPublicJWK,
			expectedIsPrivate:   false,
			expectedIsPublic:    true,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "ECDSA private key",
			jwk:                 keys.ecdsaPrivateJWK,
			expectedIsPrivate:   true,
			expectedIsPublic:    false,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "ECDSA public key",
			jwk:                 keys.ecdsaPublicJWK,
			expectedIsPrivate:   false,
			expectedIsPublic:    true,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "ECDH private key",
			jwk:                 keys.ecdhPrivateJWK,
			expectedIsPrivate:   true,
			expectedIsPublic:    false,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "ECDH public key",
			jwk:                 keys.ecdhPublicJWK,
			expectedIsPrivate:   false,
			expectedIsPublic:    true,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "OKP Ed25519 private key",
			jwk:                 keys.ed25519PrivateJWK,
			expectedIsPrivate:   true,
			expectedIsPublic:    false,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "OKP Ed25519 public key",
			jwk:                 keys.ed25519PublicJWK,
			expectedIsPrivate:   false,
			expectedIsPublic:    true,
			expectedIsSymmetric: false,
			wantErr:             nil,
		},
		{
			name:                "Symmetric key",
			jwk:                 keys.aesSecretJWK,
			expectedIsPrivate:   false,
			expectedIsPublic:    false,
			expectedIsSymmetric: true,
			wantErr:             nil,
		},
	}
}

func getTestKeys(t *testing.T) *jwkTestKeys {
	t.Helper()
	testKeysGenerateOnce.Do(func() {
		testKeys = &jwkTestKeys{}
		var rsaErr, ecdsaErr, ecdhErr, ed25519Err, aesErr error

		var wg sync.WaitGroup
		wg.Add(5)
		go func() {
			defer wg.Done()
			var rsaKeyPair *cryptoutilKeyGen.KeyPair
			rsaKeyPair, rsaErr = cryptoutilKeyGen.GenerateRSAKeyPair(2048)
			if rsaErr == nil {
				testKeys.rsaPrivateJWK, rsaErr = joseJwk.Import(rsaKeyPair.Private.(*rsa.PrivateKey))
				if rsaErr == nil {
					testKeys.rsaPublicJWK, rsaErr = joseJwk.Import(rsaKeyPair.Public.(*rsa.PublicKey))
				}
			}
		}()
		go func() {
			defer wg.Done()
			var ecdsaKeyPair *cryptoutilKeyGen.KeyPair
			ecdsaKeyPair, ecdsaErr = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
			if ecdsaErr == nil {
				testKeys.ecdsaPrivateJWK, ecdsaErr = joseJwk.Import(ecdsaKeyPair.Private.(*ecdsa.PrivateKey))
				if ecdsaErr == nil {
					testKeys.ecdsaPublicJWK, ecdsaErr = joseJwk.Import(ecdsaKeyPair.Public.(*ecdsa.PublicKey))
				}
			}
		}()
		go func() {
			defer wg.Done()
			var ecdhKeyPair *cryptoutilKeyGen.KeyPair
			ecdhKeyPair, ecdhErr = cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
			if ecdhErr == nil {
				testKeys.ecdhPrivateJWK, ecdhErr = joseJwk.Import(ecdhKeyPair.Private.(*ecdh.PrivateKey))
				if ecdhErr == nil {
					testKeys.ecdhPublicJWK, ecdhErr = joseJwk.Import(ecdhKeyPair.Public.(*ecdh.PublicKey))
				}
			}
		}()
		go func() {
			defer wg.Done()
			var ed25519KeyPair *cryptoutilKeyGen.KeyPair
			ed25519KeyPair, ed25519Err = cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
			if ed25519Err == nil {
				testKeys.ed25519PrivateJWK, ed25519Err = joseJwk.Import(ed25519KeyPair.Private.(ed25519.PrivateKey))
				if ed25519Err == nil {
					testKeys.ed25519PublicJWK, ed25519Err = joseJwk.Import(ed25519KeyPair.Public.(ed25519.PublicKey))
				}
			}
		}()
		go func() {
			defer wg.Done()
			var aesSecretKey []byte
			aesSecretKey, aesErr = cryptoutilKeyGen.GenerateAESKey(256)
			if aesErr == nil {
				testKeys.aesSecretJWK, aesErr = joseJwk.Import(aesSecretKey)
			}
		}()
		wg.Wait()
		if rsaErr != nil || ecdsaErr != nil || ecdhErr != nil || ed25519Err != nil || aesErr != nil {
			t.Fatalf("failed to generate keys: %v", errors.Join(rsaErr, ecdsaErr, ecdhErr, ed25519Err, aesErr))
		}
	})

	return testKeys
}
