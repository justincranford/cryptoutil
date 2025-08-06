package jose

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
)

func TestIsPrivateJwk(t *testing.T) {
	// Generate test keys
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	rsaPubKey := &rsaPrivKey.PublicKey

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	ecPubKey := &ecPrivKey.PublicKey

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	symKey, err := cryptoutilKeyGen.GenerateAESKey(256)
	assert.NoError(t, err)

	// Create JWK keys
	rsaPrivJwk, err := joseJwk.Import(rsaPrivKey)
	assert.NoError(t, err)
	rsaPubJwk, err := joseJwk.Import(rsaPubKey)
	assert.NoError(t, err)

	ecPrivJwk, err := joseJwk.Import(ecPrivKey)
	assert.NoError(t, err)
	ecPubJwk, err := joseJwk.Import(ecPubKey)
	assert.NoError(t, err)

	okpPrivJwk, err := joseJwk.Import(privKey)
	assert.NoError(t, err)
	okpPubJwk, err := joseJwk.Import(pubKey)
	assert.NoError(t, err)

	symJwk, err := joseJwk.Import([]byte(symKey))
	assert.NoError(t, err)

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
			jwk:      rsaPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      ecPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      ecPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      okpPrivJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      symJwk,
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
	// Generate test keys
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	rsaPubKey := &rsaPrivKey.PublicKey

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	ecPubKey := &ecPrivKey.PublicKey

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	symKey, err := cryptoutilKeyGen.GenerateAESKey(256)
	assert.NoError(t, err)

	// Create JWK keys
	rsaPrivJwk, err := joseJwk.Import(rsaPrivKey)
	assert.NoError(t, err)
	rsaPubJwk, err := joseJwk.Import(rsaPubKey)
	assert.NoError(t, err)

	ecPrivJwk, err := joseJwk.Import(ecPrivKey)
	assert.NoError(t, err)
	ecPubJwk, err := joseJwk.Import(ecPubKey)
	assert.NoError(t, err)

	okpPrivJwk, err := joseJwk.Import(privKey)
	assert.NoError(t, err)
	okpPubJwk, err := joseJwk.Import(pubKey)
	assert.NoError(t, err)

	symJwk, err := joseJwk.Import([]byte(symKey))
	assert.NoError(t, err)

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
			jwk:      rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      rsaPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      ecPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      ecPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      okpPubJwk,
			expected: true,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      symJwk,
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
	// Generate test keys
	rsaPrivKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	rsaPubKey := &rsaPrivKey.PublicKey

	ecPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	ecPubKey := &ecPrivKey.PublicKey

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	symKey, err := cryptoutilKeyGen.GenerateAESKey(256)
	assert.NoError(t, err)

	// Create JWK keys
	rsaPrivJwk, err := joseJwk.Import(rsaPrivKey)
	assert.NoError(t, err)
	rsaPubJwk, err := joseJwk.Import(rsaPubKey)
	assert.NoError(t, err)

	ecPrivJwk, err := joseJwk.Import(ecPrivKey)
	assert.NoError(t, err)
	ecPubJwk, err := joseJwk.Import(ecPubKey)
	assert.NoError(t, err)

	okpPrivJwk, err := joseJwk.Import(privKey)
	assert.NoError(t, err)
	okpPubJwk, err := joseJwk.Import(pubKey)
	assert.NoError(t, err)

	symJwk, err := joseJwk.Import([]byte(symKey))
	assert.NoError(t, err)

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
			jwk:      rsaPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "RSA public key",
			jwk:      rsaPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA private key",
			jwk:      ecPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "ECDSA public key",
			jwk:      ecPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 private key",
			jwk:      okpPrivJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "OKP Ed25519 public key",
			jwk:      okpPubJwk,
			expected: false,
			wantErr:  nil,
		},
		{
			name:     "Symmetric key",
			jwk:      symJwk,
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
