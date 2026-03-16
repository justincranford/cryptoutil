// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestNewJWSIssuerLegacy_Success validates successful JWS issuer creation with valid parameters.
func TestNewJWSIssuerLegacy_Success(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	issuer, err := cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		"https://auth.example.com",
		privateKey,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)

	require.NoError(t, err)
	require.NotNil(t, issuer)
}

// TestNewJWSIssuerLegacy_EmptyIssuer validates error when issuer is empty.
func TestNewJWSIssuerLegacy_EmptyIssuer(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	_, err = cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		"", // Empty issuer
		privateKey,
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidConfiguration)
}

// TestNewJWSIssuerLegacy_EmptySigningAlgorithm validates error when signing algorithm is empty.
func TestNewJWSIssuerLegacy_EmptySigningAlgorithm(t *testing.T) {
	t.Parallel()

	privateKey, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	_, err = cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		"https://auth.example.com",
		privateKey,
		"", // Empty algorithm
		time.Hour,
		time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidConfiguration)
}

// TestNewJWSIssuerLegacy_NilSigningKey validates error when signing key is nil.
func TestNewJWSIssuerLegacy_NilSigningKey(t *testing.T) {
	t.Parallel()

	_, err := cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		"https://auth.example.com",
		nil, // Nil key
		cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
		time.Hour,
		time.Hour,
	)

	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidConfiguration)
}

// TestNewJWSIssuerLegacy_ECDSAKey validates successful creation with ECDSA key.
func TestNewJWSIssuerLegacy_ECDSAKey(t *testing.T) {
	t.Parallel()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	issuer, err := cryptoutilIdentityIssuer.NewJWSIssuerLegacy(
		"https://auth.example.com",
		privateKey,
		cryptoutilSharedMagic.JoseAlgES256,
		time.Hour,
		time.Hour,
	)

	require.NoError(t, err)
	require.NotNil(t, issuer)
}

// TestNewJWEIssuerLegacy_Success validates successful JWE issuer creation with valid key.
func TestNewJWEIssuerLegacy_Success(t *testing.T) {
	t.Parallel()

	// AES-256 requires 32 bytes.
	encryptionKey := make([]byte, cryptoutilSharedMagic.AES256KeySize)
	_, err := crand.Read(encryptionKey)
	require.NoError(t, err)

	issuer, err := cryptoutilIdentityIssuer.NewJWEIssuerLegacy(encryptionKey)

	require.NoError(t, err)
	require.NotNil(t, issuer)
}

// TestNewJWEIssuerLegacy_InvalidKeySize validates error when encryption key has wrong size.
func TestNewJWEIssuerLegacy_InvalidKeySize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		keySize   int
		wantError bool
	}{
		{
			name:      "too short (16 bytes)",
			keySize:   cryptoutilSharedMagic.RealmMinTokenLengthBytes,
			wantError: true,
		},
		{
			name:      "too short (24 bytes)",
			keySize:   cryptoutilSharedMagic.HoursPerDay,
			wantError: true,
		},
		{
			name:      "valid (32 bytes)",
			keySize:   cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
			wantError: false,
		},
		{
			name:      "too long (48 bytes)",
			keySize:   cryptoutilSharedMagic.HMACSHA384KeySize,
			wantError: true,
		},
		{
			name:      "empty (0 bytes)",
			keySize:   0,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			encryptionKey := make([]byte, tc.keySize)
			if tc.keySize > 0 {
				_, err := crand.Read(encryptionKey)
				require.NoError(t, err)
			}

			issuer, err := cryptoutilIdentityIssuer.NewJWEIssuerLegacy(encryptionKey)

			if tc.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidConfiguration)
				require.Nil(t, issuer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, issuer)
			}
		})
	}
}
