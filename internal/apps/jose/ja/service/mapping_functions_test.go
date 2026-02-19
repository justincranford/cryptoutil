// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"testing"

	jose "github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestMapToJWEAlgorithms_AllBranches tests all branches in mapToJWEAlgorithms.
func TestMapToJWEAlgorithms_AllBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		algorithm      string
		expectedKeyAlg jose.KeyAlgorithm
		expectedEnc    jose.ContentEncryption
	}{
		// RSA variants.
		{
			name:           "RSA/2048",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA2048,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "RSA/3072",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA3072,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "RSA/4096",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA4096,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "RSA-OAEP",
			algorithm:      cryptoutilSharedMagic.JoseAlgRSAOAEP,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "RSA-OAEP-256",
			algorithm:      cryptoutilSharedMagic.JoseAlgRSAOAEP256,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		// EC variants.
		{
			name:           "EC/P256",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP256,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "EC/P384",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP384,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "EC/P521",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP521,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           "ECDH-ES",
			algorithm:      cryptoutilSharedMagic.JoseAlgECDHES,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		// AES key wrapping variants.
		{
			name:           "A128KW",
			algorithm:      "A128KW",
			expectedKeyAlg: jose.A128KW,
			expectedEnc:    jose.A128GCM,
		},
		{
			name:           "A192KW",
			algorithm:      "A192KW",
			expectedKeyAlg: jose.A192KW,
			expectedEnc:    jose.A192GCM,
		},
		{
			name:           "A256KW",
			algorithm:      "A256KW",
			expectedKeyAlg: jose.A256KW,
			expectedEnc:    jose.A256GCM,
		},
		// AES GCM key wrapping variants.
		{
			name:           "A128GCMKW",
			algorithm:      "A128GCMKW",
			expectedKeyAlg: jose.A128GCMKW,
			expectedEnc:    jose.A128GCM,
		},
		{
			name:           "A192GCMKW",
			algorithm:      "A192GCMKW",
			expectedKeyAlg: jose.A192GCMKW,
			expectedEnc:    jose.A192GCM,
		},
		{
			name:           "A256GCMKW",
			algorithm:      "A256GCMKW",
			expectedKeyAlg: jose.A256GCMKW,
			expectedEnc:    jose.A256GCM,
		},
		// Direct encryption variants.
		{
			name:           "dir (A128GCM)",
			algorithm:      cryptoutilSharedMagic.JoseAlgDir,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A128GCM,
		},
		{
			name:           "oct/128",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeOct128,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A128GCM,
		},
		{
			name:           "oct/192",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeOct192,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A192GCM,
		},
		{
			name:           "oct/256",
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeOct256,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A256GCM,
		},
		// Invalid/unknown.
		{
			name:           "invalid algorithm",
			algorithm:      "INVALID",
			expectedKeyAlg: "",
			expectedEnc:    "",
		},
		{
			name:           "empty algorithm",
			algorithm:      "",
			expectedKeyAlg: "",
			expectedEnc:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			keyAlg, enc := mapToJWEAlgorithms(tt.algorithm)
			require.Equal(t, tt.expectedKeyAlg, keyAlg)
			require.Equal(t, tt.expectedEnc, enc)
		})
	}
}

// TestMapToSignatureAlgorithm_AllBranches tests all branches in mapToSignatureAlgorithm.
func TestMapToSignatureAlgorithm_AllBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		algorithm   string
		expectedAlg jose.SignatureAlgorithm
	}{
		// RSA PKCS#1 variants.
		{
			name:        "RS256",
			algorithm:   "RS256",
			expectedAlg: jose.RS256,
		},
		{
			name:        "RS384",
			algorithm:   "RS384",
			expectedAlg: jose.RS384,
		},
		{
			name:        "RS512",
			algorithm:   "RS512",
			expectedAlg: jose.RS512,
		},
		{
			name:        "RSA/2048",
			algorithm:   "RSA/2048",
			expectedAlg: jose.RS256,
		},
		{
			name:        "RSA/3072",
			algorithm:   "RSA/3072",
			expectedAlg: jose.RS384,
		},
		{
			name:        "RSA/4096",
			algorithm:   "RSA/4096",
			expectedAlg: jose.RS512,
		},
		// RSA-PSS variants.
		{
			name:        "PS256",
			algorithm:   "PS256",
			expectedAlg: jose.PS256,
		},
		{
			name:        "PS384",
			algorithm:   "PS384",
			expectedAlg: jose.PS384,
		},
		{
			name:        "PS512",
			algorithm:   "PS512",
			expectedAlg: jose.PS512,
		},
		// ECDSA variants.
		{
			name:        "ES256",
			algorithm:   "ES256",
			expectedAlg: jose.ES256,
		},
		{
			name:        "ES384",
			algorithm:   "ES384",
			expectedAlg: jose.ES384,
		},
		{
			name:        "ES512",
			algorithm:   "ES512",
			expectedAlg: jose.ES512,
		},
		{
			name:        "EC/P256",
			algorithm:   "EC/P256",
			expectedAlg: jose.ES256,
		},
		{
			name:        "EC/P384",
			algorithm:   "EC/P384",
			expectedAlg: jose.ES384,
		},
		{
			name:        "EC/P521",
			algorithm:   "EC/P521",
			expectedAlg: jose.ES512,
		},
		// EdDSA variants.
		{
			name:        "EdDSA",
			algorithm:   "EdDSA",
			expectedAlg: jose.EdDSA,
		},
		{
			name:        "OKP/Ed25519",
			algorithm:   "OKP/Ed25519",
			expectedAlg: jose.EdDSA,
		},
		// HMAC variants.
		{
			name:        "HS256",
			algorithm:   "HS256",
			expectedAlg: jose.HS256,
		},
		{
			name:        "HS384",
			algorithm:   "HS384",
			expectedAlg: jose.HS384,
		},
		{
			name:        "HS512",
			algorithm:   "HS512",
			expectedAlg: jose.HS512,
		},
		{
			name:        "oct/256",
			algorithm:   "oct/256",
			expectedAlg: jose.HS256,
		},
		{
			name:        "oct/384",
			algorithm:   "oct/384",
			expectedAlg: jose.HS384,
		},
		{
			name:        "oct/512",
			algorithm:   "oct/512",
			expectedAlg: jose.HS512,
		},
		// Invalid/unknown.
		{
			name:        "invalid algorithm",
			algorithm:   "INVALID",
			expectedAlg: "",
		},
		{
			name:        "empty algorithm",
			algorithm:   "",
			expectedAlg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			alg := mapToSignatureAlgorithm(tt.algorithm)
			require.Equal(t, tt.expectedAlg, alg)
		})
	}
}

// TestMapToGenerateAlgorithmForRotation_AllBranches tests all branches in mapToGenerateAlgorithmForRotation.
