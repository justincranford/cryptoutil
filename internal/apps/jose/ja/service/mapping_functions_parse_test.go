// Copyright (c) 2025 Justin Cranford
//

package service

import (
	json "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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
