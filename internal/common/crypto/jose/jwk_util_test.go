package jose

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"sync"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwkTestKeys struct {
	rsaPrivJwk   joseJwk.Key
	rsaPubJwk    joseJwk.Key
	ecdsaPrivJwk joseJwk.Key
	ecdsaPubJwk  joseJwk.Key
	ecdhPrivJwk  joseJwk.Key
	ecdhPubJwk   joseJwk.Key
	okpPrivJwk   joseJwk.Key
	okpPubJwk    joseJwk.Key
	symJwk       joseJwk.Key
}

var (
	testKeys     *jwkTestKeys
	testKeysOnce sync.Once
)

func getTestKeys(t *testing.T) *jwkTestKeys {
	t.Helper()

	testKeysOnce.Do(func() {
		rsaKeyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
		require.NoError(t, err, "failed to generate RSA key")
		ecdsaKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
		require.NoError(t, err, "failed to generate ECDSA key")
		ecdhKeyPair, err := cryptoutilKeyGen.GenerateECDHKeyPair(ecdh.P256())
		require.NoError(t, err, "failed to generate ECDH key")
		ed25519KeyPair, err := cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
		require.NoError(t, err, "failed to generate Ed25519 key")
		aesKey, err := cryptoutilKeyGen.GenerateAESKey(256)
		require.NoError(t, err, "failed to generate AES key")

		testKeys = &jwkTestKeys{}
		testKeys.rsaPrivJwk, err = joseJwk.Import(rsaKeyPair.Private.(*rsa.PrivateKey))
		require.NoError(t, err, "failed to import RSA private key to JWK")
		testKeys.rsaPubJwk, err = joseJwk.Import(rsaKeyPair.Public.(*rsa.PublicKey))
		require.NoError(t, err, "failed to import RSA public key to JWK")
		testKeys.ecdsaPrivJwk, err = joseJwk.Import(ecdsaKeyPair.Private.(*ecdsa.PrivateKey))
		require.NoError(t, err, "failed to import ECDSA private key to JWK")
		testKeys.ecdsaPubJwk, err = joseJwk.Import(ecdsaKeyPair.Public.(*ecdsa.PublicKey))
		require.NoError(t, err, "failed to import ECDSA public key to JWK")
		testKeys.ecdhPrivJwk, err = joseJwk.Import(ecdhKeyPair.Private.(*ecdh.PrivateKey))
		require.NoError(t, err, "failed to import ECDH private key to JWK")
		testKeys.ecdhPubJwk, err = joseJwk.Import(ecdhKeyPair.Public.(*ecdh.PublicKey))
		require.NoError(t, err, "failed to import ECDH public key to JWK")
		testKeys.okpPrivJwk, err = joseJwk.Import(ed25519KeyPair.Private.(ed25519.PrivateKey))
		require.NoError(t, err, "failed to import Ed25519 private key to JWK")
		testKeys.okpPubJwk, err = joseJwk.Import(ed25519KeyPair.Public.(ed25519.PublicKey))
		require.NoError(t, err, "failed to import Ed25519 public key to JWK")
		testKeys.symJwk, err = joseJwk.Import([]byte(aesKey))
		require.NoError(t, err, "failed to import AES secret key to JWK")
	})

	return testKeys
}

func TestIsPrivateJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: false,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPrivate, err := IsPrivateJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isPrivate)
			}
		})
	}
}

func TestIsPublicJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: false,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isPublic, err := IsPublicJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isPublic)
			}
		})
	}
}

func TestIsSymmetricJwk(t *testing.T) {
	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

	keys := getTestKeys(t)
	tests := []testCase{
		{
			name:     "nil JWK",
			jwk:      nil,
			expected: false,
			wantErr:  cryptoutilAppErr.ErrCantBeNil,
		},
		{
			name:     "RSA private key",
			jwk:      keys.rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      keys.rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      keys.ecdsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecdsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH private key",
			jwk:      keys.ecdhPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDH public key",
			jwk:      keys.ecdhPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      keys.okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      keys.okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      keys.symJwk,
			expected: true,
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isSymmetric, err := IsSymmetricJwk(tc.jwk)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, isSymmetric)
			}
		})
	}
}
