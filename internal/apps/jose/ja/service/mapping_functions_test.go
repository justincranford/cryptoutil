// Copyright (c) 2025 Justin Cranford
//

package service

import (
	json "encoding/json"
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
func TestMapToGenerateAlgorithmForRotation_AllBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		expectNil bool
	}{
		// RSA signing.
		{name: "RS256", algorithm: cryptoutilSharedMagic.JoseAlgRS256, expectNil: false},
		{name: "RS384", algorithm: cryptoutilSharedMagic.JoseAlgRS384, expectNil: false},
		{name: "RS512", algorithm: cryptoutilSharedMagic.JoseAlgRS512, expectNil: false},
		{name: "RSA/2048", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA2048, expectNil: false},
		// RSA-PSS signing.
		{name: "PS256", algorithm: cryptoutilSharedMagic.JoseAlgPS256, expectNil: false},
		{name: "PS384", algorithm: cryptoutilSharedMagic.JoseAlgPS384, expectNil: false},
		{name: "PS512", algorithm: cryptoutilSharedMagic.JoseAlgPS512, expectNil: false},
		{name: "RSA/3072", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA3072, expectNil: false},
		// RSA/4096.
		{name: "RSA/4096", algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA4096, expectNil: false},
		// ECDSA signing.
		{name: "ES256", algorithm: cryptoutilSharedMagic.JoseAlgES256, expectNil: false},
		{name: "EC/P256", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP256, expectNil: false},
		{name: "ES384", algorithm: cryptoutilSharedMagic.JoseAlgES384, expectNil: false},
		{name: "EC/P384", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP384, expectNil: false},
		{name: "ES512", algorithm: cryptoutilSharedMagic.JoseAlgES512, expectNil: false},
		{name: "EC/P521", algorithm: cryptoutilSharedMagic.JoseKeyTypeECP521, expectNil: false},
		// EdDSA.
		{name: "EdDSA", algorithm: cryptoutilSharedMagic.JoseAlgEdDSA, expectNil: false},
		{name: "OKP/Ed25519", algorithm: cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, expectNil: false},
		// Symmetric keys.
		{name: "oct/128", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct128, expectNil: false},
		{name: "A128GCM", algorithm: cryptoutilSharedMagic.JoseEncA128GCM, expectNil: false},
		{name: "oct/192", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct192, expectNil: false},
		{name: "A192GCM", algorithm: cryptoutilSharedMagic.JoseEncA192GCM, expectNil: false},
		{name: "oct/256", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct256, expectNil: false},
		{name: "A256GCM", algorithm: cryptoutilSharedMagic.JoseEncA256GCM, expectNil: false},
		{name: "oct/384", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct384, expectNil: false},
		{name: "A128CBC-HS256", algorithm: cryptoutilSharedMagic.JoseEncA128CBCHS256, expectNil: false},
		{name: "oct/512", algorithm: cryptoutilSharedMagic.JoseKeyTypeOct512, expectNil: false},
		{name: "A256CBC-HS512", algorithm: cryptoutilSharedMagic.JoseEncA256CBCHS512, expectNil: false},
		// Invalid/unknown.
		{name: "invalid", algorithm: "INVALID", expectNil: true},
		{name: "empty", algorithm: "", expectNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mapToGenerateAlgorithmForRotation(tt.algorithm)

			if tt.expectNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

// TestParseClaimsMap_AllBranches tests all type assertion branches in parseClaimsMap.
func TestParseClaimsMap_AllBranches(t *testing.T) {
	t.Parallel()

	// Create a service instance to call parseClaimsMap directly.
	svc := &jwtServiceImpl{}

	tests := []struct {
		name      string
		claimsMap map[string]any
		verify    func(t *testing.T, claims *JWTClaims)
	}{
		{
			name: "iss as non-string (int) - type assertion fails",
			claimsMap: map[string]any{
				"iss": 12345, // Not a string.
				"sub": "test-subject",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.Issuer)
				require.Equal(t, "test-subject", claims.Subject)
			},
		},
		{
			name: "sub as non-string (bool) - type assertion fails",
			claimsMap: map[string]any{
				"iss": "test-issuer",
				"sub": true, // Not a string.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "test-issuer", claims.Issuer)
				require.Empty(t, claims.Subject)
			},
		},
		{
			name: "aud as string",
			claimsMap: map[string]any{
				"aud": "single-audience",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"single-audience"}, claims.Audience)
			},
		},
		{
			name: "aud as array of strings",
			claimsMap: map[string]any{
				"aud": []any{"aud1", "aud2", "aud3"},
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"aud1", "aud2", "aud3"}, claims.Audience)
			},
		},
		{
			name: "aud as neither string nor array",
			claimsMap: map[string]any{
				"aud": 12345, // Neither string nor array.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.Audience)
			},
		},
		{
			name: "aud as array with non-string items (filters out)",
			claimsMap: map[string]any{
				"aud": []any{123, "valid-aud", true}, // Mixed types.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"valid-aud"}, claims.Audience)
			},
		},
		{
			name: "exp as float64",
			claimsMap: map[string]any{
				"exp": float64(1700000000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.ExpiresAt)
				require.Equal(t, int64(1700000000), claims.ExpiresAt.Unix())
			},
		},
		{
			name: "exp as json.Number",
			claimsMap: map[string]any{
				"exp": json.Number("1700000000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.ExpiresAt)
				require.Equal(t, int64(1700000000), claims.ExpiresAt.Unix())
			},
		},
		{
			name: "exp as non-numeric value",
			claimsMap: map[string]any{
				"exp": "not-a-number", // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.ExpiresAt)
			},
		},
		{
			name: "nbf as float64",
			claimsMap: map[string]any{
				"nbf": float64(1699990000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.NotBefore)
				require.Equal(t, int64(1699990000), claims.NotBefore.Unix())
			},
		},
		{
			name: "nbf as json.Number",
			claimsMap: map[string]any{
				"nbf": json.Number("1699990000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.NotBefore)
				require.Equal(t, int64(1699990000), claims.NotBefore.Unix())
			},
		},
		{
			name: "nbf as non-numeric value",
			claimsMap: map[string]any{
				"nbf": true, // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.NotBefore)
			},
		},
		{
			name: "iat as float64",
			claimsMap: map[string]any{
				"iat": float64(1699980000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.IssuedAt)
				require.Equal(t, int64(1699980000), claims.IssuedAt.Unix())
			},
		},
		{
			name: "iat as json.Number",
			claimsMap: map[string]any{
				"iat": json.Number("1699980000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.IssuedAt)
				require.Equal(t, int64(1699980000), claims.IssuedAt.Unix())
			},
		},
		{
			name: "iat as non-numeric value",
			claimsMap: map[string]any{
				"iat": []string{"array"}, // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.IssuedAt)
			},
		},
		{
			name: "jti as string",
			claimsMap: map[string]any{
				"jti": "test-jti",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "test-jti", claims.JTI)
			},
		},
		{
			name: "jti as non-string",
			claimsMap: map[string]any{
				"jti": 12345, // Not a string.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.JTI)
			},
		},
		{
			name: "custom claims are preserved",
			claimsMap: map[string]any{
				"custom_string": "value",
				"custom_int":    42,
				"custom_bool":   true,
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "value", claims.Custom["custom_string"])
				require.Equal(t, 42, claims.Custom["custom_int"])
				require.Equal(t, true, claims.Custom["custom_bool"])
			},
		},
		{
			name: "all claims combined",
			claimsMap: map[string]any{
				"iss":    "combined-issuer",
				"sub":    "combined-subject",
				"aud":    []any{"aud1", "aud2"},
				"exp":    json.Number("1700000000"),
				"nbf":    float64(1699990000),
				"iat":    float64(1699980000),
				"jti":    "combined-jti",
				"custom": "custom-value",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "combined-issuer", claims.Issuer)
				require.Equal(t, "combined-subject", claims.Subject)
				require.Equal(t, []string{"aud1", "aud2"}, claims.Audience)
				require.NotNil(t, claims.ExpiresAt)
				require.NotNil(t, claims.NotBefore)
				require.NotNil(t, claims.IssuedAt)
				require.Equal(t, "combined-jti", claims.JTI)
				require.Equal(t, "custom-value", claims.Custom["custom"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			claims := svc.parseClaimsMap(tt.claimsMap)
			tt.verify(t, claims)
		})
	}
}
