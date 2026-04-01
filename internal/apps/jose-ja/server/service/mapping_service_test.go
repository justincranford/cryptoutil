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
			name:           cryptoutilSharedMagic.JoseKeyTypeRSA2048,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA2048,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeRSA3072,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA3072,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeRSA4096,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeRSA4096,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseAlgRSAOAEP,
			algorithm:      cryptoutilSharedMagic.JoseAlgRSAOAEP,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseAlgRSAOAEP256,
			algorithm:      cryptoutilSharedMagic.JoseAlgRSAOAEP256,
			expectedKeyAlg: jose.RSA_OAEP_256,
			expectedEnc:    jose.A256GCM,
		},
		// EC variants.
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeECP256,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP256,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeECP384,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP384,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeECP521,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeECP521,
			expectedKeyAlg: jose.ECDH_ES_A256KW,
			expectedEnc:    jose.A256GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseAlgECDHES,
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
			name:           cryptoutilSharedMagic.JoseKeyTypeOct128,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeOct128,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A128GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeOct192,
			algorithm:      cryptoutilSharedMagic.JoseKeyTypeOct192,
			expectedKeyAlg: jose.DIRECT,
			expectedEnc:    jose.A192GCM,
		},
		{
			name:           cryptoutilSharedMagic.JoseKeyTypeOct256,
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
			name:        cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			algorithm:   cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm,
			expectedAlg: jose.RS256,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgRS384,
			algorithm:   cryptoutilSharedMagic.JoseAlgRS384,
			expectedAlg: jose.RS384,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgRS512,
			algorithm:   cryptoutilSharedMagic.JoseAlgRS512,
			expectedAlg: jose.RS512,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeRSA2048,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeRSA2048,
			expectedAlg: jose.RS256,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeRSA3072,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeRSA3072,
			expectedAlg: jose.RS384,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeRSA4096,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeRSA4096,
			expectedAlg: jose.RS512,
		},
		// RSA-PSS variants.
		{
			name:        cryptoutilSharedMagic.JoseAlgPS256,
			algorithm:   cryptoutilSharedMagic.JoseAlgPS256,
			expectedAlg: jose.PS256,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgPS384,
			algorithm:   cryptoutilSharedMagic.JoseAlgPS384,
			expectedAlg: jose.PS384,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgPS512,
			algorithm:   cryptoutilSharedMagic.JoseAlgPS512,
			expectedAlg: jose.PS512,
		},
		// ECDSA variants.
		{
			name:        cryptoutilSharedMagic.JoseAlgES256,
			algorithm:   cryptoutilSharedMagic.JoseAlgES256,
			expectedAlg: jose.ES256,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgES384,
			algorithm:   cryptoutilSharedMagic.JoseAlgES384,
			expectedAlg: jose.ES384,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgES512,
			algorithm:   cryptoutilSharedMagic.JoseAlgES512,
			expectedAlg: jose.ES512,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeECP256,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeECP256,
			expectedAlg: jose.ES256,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeECP384,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeECP384,
			expectedAlg: jose.ES384,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeECP521,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeECP521,
			expectedAlg: jose.ES512,
		},
		// EdDSA variants.
		{
			name:        cryptoutilSharedMagic.JoseAlgEdDSA,
			algorithm:   cryptoutilSharedMagic.JoseAlgEdDSA,
			expectedAlg: jose.EdDSA,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeOKPEd25519,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeOKPEd25519,
			expectedAlg: jose.EdDSA,
		},
		// HMAC variants.
		{
			name:        cryptoutilSharedMagic.JoseAlgHS256,
			algorithm:   cryptoutilSharedMagic.JoseAlgHS256,
			expectedAlg: jose.HS256,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgHS384,
			algorithm:   cryptoutilSharedMagic.JoseAlgHS384,
			expectedAlg: jose.HS384,
		},
		{
			name:        cryptoutilSharedMagic.JoseAlgHS512,
			algorithm:   cryptoutilSharedMagic.JoseAlgHS512,
			expectedAlg: jose.HS512,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeOct256,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeOct256,
			expectedAlg: jose.HS256,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeOct384,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeOct384,
			expectedAlg: jose.HS384,
		},
		{
			name:        cryptoutilSharedMagic.JoseKeyTypeOct512,
			algorithm:   cryptoutilSharedMagic.JoseKeyTypeOct512,
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
