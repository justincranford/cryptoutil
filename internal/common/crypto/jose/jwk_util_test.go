package jose

import (
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
	// JWK keys only - we don't need to store references to the raw keys
	rsaPrivJwk joseJwk.Key
	rsaPubJwk  joseJwk.Key
	ecPrivJwk  joseJwk.Key
	ecPubJwk   joseJwk.Key
	okpPrivJwk joseJwk.Key
	okpPubJwk  joseJwk.Key
	symJwk     joseJwk.Key
}

var (
	// Global test keys
	testKeys     *jwkTestKeys
	testKeysOnce sync.Once
)

// getTestKeys generates or returns cached test keys
func getTestKeys(t *testing.T) *jwkTestKeys {
	t.Helper()

	testKeysOnce.Do(func() {
		testKeys = &jwkTestKeys{}
		var err error
		var keyPair *cryptoutilKeyGen.KeyPair

		// Generate RSA keys using cryptoutil keygen
		keyPair, err = cryptoutilKeyGen.GenerateRSAKeyPair(2048)
		require.NoError(t, err, "Failed to generate RSA key")
		rsaPrivKey := keyPair.Private.(*rsa.PrivateKey)
		rsaPubKey := keyPair.Public.(*rsa.PublicKey)

		// Generate ECDSA keys using cryptoutil keygen
		keyPair, err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P256())
		require.NoError(t, err, "Failed to generate ECDSA key")
		ecPrivKey := keyPair.Private.(*ecdsa.PrivateKey)
		ecPubKey := keyPair.Public.(*ecdsa.PublicKey)

		// Generate EdDSA keys using cryptoutil keygen
		keyPair, err = cryptoutilKeyGen.GenerateEDDSAKeyPair("Ed25519")
		require.NoError(t, err, "Failed to generate Ed25519 key")
		edPrivKey := keyPair.Private.(ed25519.PrivateKey)
		edPubKey := keyPair.Public.(ed25519.PublicKey)

		// Generate symmetric key
		symKey, err := cryptoutilKeyGen.GenerateAESKey(256)
		require.NoError(t, err, "Failed to generate AES key")

		// Convert to JWK format
		testKeys.rsaPrivJwk, err = joseJwk.Import(rsaPrivKey)
		require.NoError(t, err, "Failed to import RSA private key to JWK")
		testKeys.rsaPubJwk, err = joseJwk.Import(rsaPubKey)
		require.NoError(t, err, "Failed to import RSA public key to JWK")

		testKeys.ecPrivJwk, err = joseJwk.Import(ecPrivKey)
		require.NoError(t, err, "Failed to import ECDSA private key to JWK")
		testKeys.ecPubJwk, err = joseJwk.Import(ecPubKey)
		require.NoError(t, err, "Failed to import ECDSA public key to JWK")

		testKeys.okpPrivJwk, err = joseJwk.Import(edPrivKey)
		require.NoError(t, err, "Failed to import Ed25519 private key to JWK")
		testKeys.okpPubJwk, err = joseJwk.Import(edPubKey)
		require.NoError(t, err, "Failed to import Ed25519 public key to JWK")

		testKeys.symJwk, err = joseJwk.Import([]byte(symKey))
		require.NoError(t, err, "Failed to import symmetric key to JWK")
	})

	return testKeys
}

func TestIsPrivateJwk(t *testing.T) {
	keys := getTestKeys(t)

	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

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
			jwk:      keys.ecPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecPubJwk,
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
	keys := getTestKeys(t)

	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

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
			jwk:      keys.ecPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecPubJwk,
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
	keys := getTestKeys(t)

	type testCase struct {
		name     string
		jwk      joseJwk.Key
		expected bool
		wantErr  error
	}

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
			jwk:      keys.ecPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      keys.ecPubJwk,
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
