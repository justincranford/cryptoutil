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
		{name: cryptoutilSharedMagic.DefaultBrowserSessionJWSAlgorithm, algorithm: cryptoutilSharedMagic.JoseAlgRS256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgRS384, algorithm: cryptoutilSharedMagic.JoseAlgRS384, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgRS512, algorithm: cryptoutilSharedMagic.JoseAlgRS512, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeRSA2048, algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA2048, expectNil: false},
		// RSA-PSS signing.
		{name: cryptoutilSharedMagic.JoseAlgPS256, algorithm: cryptoutilSharedMagic.JoseAlgPS256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgPS384, algorithm: cryptoutilSharedMagic.JoseAlgPS384, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgPS512, algorithm: cryptoutilSharedMagic.JoseAlgPS512, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeRSA3072, algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA3072, expectNil: false},
		// RSA/4096.
		{name: cryptoutilSharedMagic.JoseKeyTypeRSA4096, algorithm: cryptoutilSharedMagic.JoseKeyTypeRSA4096, expectNil: false},
		// ECDSA signing.
		{name: cryptoutilSharedMagic.JoseAlgES256, algorithm: cryptoutilSharedMagic.JoseAlgES256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeECP256, algorithm: cryptoutilSharedMagic.JoseKeyTypeECP256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgES384, algorithm: cryptoutilSharedMagic.JoseAlgES384, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeECP384, algorithm: cryptoutilSharedMagic.JoseKeyTypeECP384, expectNil: false},
		{name: cryptoutilSharedMagic.JoseAlgES512, algorithm: cryptoutilSharedMagic.JoseAlgES512, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeECP521, algorithm: cryptoutilSharedMagic.JoseKeyTypeECP521, expectNil: false},
		// EdDSA.
		{name: cryptoutilSharedMagic.JoseAlgEdDSA, algorithm: cryptoutilSharedMagic.JoseAlgEdDSA, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, algorithm: cryptoutilSharedMagic.JoseKeyTypeOKPEd25519, expectNil: false},
		// Symmetric keys.
		{name: cryptoutilSharedMagic.JoseKeyTypeOct128, algorithm: cryptoutilSharedMagic.JoseKeyTypeOct128, expectNil: false},
		{name: cryptoutilSharedMagic.JoseEncA128GCM, algorithm: cryptoutilSharedMagic.JoseEncA128GCM, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeOct192, algorithm: cryptoutilSharedMagic.JoseKeyTypeOct192, expectNil: false},
		{name: cryptoutilSharedMagic.JoseEncA192GCM, algorithm: cryptoutilSharedMagic.JoseEncA192GCM, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeOct256, algorithm: cryptoutilSharedMagic.JoseKeyTypeOct256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseEncA256GCM, algorithm: cryptoutilSharedMagic.JoseEncA256GCM, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeOct384, algorithm: cryptoutilSharedMagic.JoseKeyTypeOct384, expectNil: false},
		{name: cryptoutilSharedMagic.JoseEncA128CBCHS256, algorithm: cryptoutilSharedMagic.JoseEncA128CBCHS256, expectNil: false},
		{name: cryptoutilSharedMagic.JoseKeyTypeOct512, algorithm: cryptoutilSharedMagic.JoseKeyTypeOct512, expectNil: false},
		{name: cryptoutilSharedMagic.JoseEncA256CBCHS512, algorithm: cryptoutilSharedMagic.JoseEncA256CBCHS512, expectNil: false},
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
				cryptoutilSharedMagic.ClaimIss: 12345, // Not a string.
				cryptoutilSharedMagic.ClaimSub: "test-subject",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.Issuer)
				require.Equal(t, "test-subject", claims.Subject)
			},
		},
		{
			name: "sub as non-string (bool) - type assertion fails",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimIss: "test-issuer",
				cryptoutilSharedMagic.ClaimSub: true, // Not a string.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "test-issuer", claims.Issuer)
				require.Empty(t, claims.Subject)
			},
		},
		{
			name: "aud as string",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimAud: "single-audience",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"single-audience"}, claims.Audience)
			},
		},
		{
			name: "aud as array of strings",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimAud: []any{"aud1", "aud2", "aud3"},
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"aud1", "aud2", "aud3"}, claims.Audience)
			},
		},
		{
			name: "aud as neither string nor array",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimAud: 12345, // Neither string nor array.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.Audience)
			},
		},
		{
			name: "aud as array with non-string items (filters out)",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimAud: []any{123, "valid-aud", true}, // Mixed types.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, []string{"valid-aud"}, claims.Audience)
			},
		},
		{
			name: "exp as float64",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimExp: float64(1700000000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.ExpiresAt)
				require.Equal(t, int64(1700000000), claims.ExpiresAt.Unix())
			},
		},
		{
			name: "exp as json.Number",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimExp: json.Number("1700000000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.ExpiresAt)
				require.Equal(t, int64(1700000000), claims.ExpiresAt.Unix())
			},
		},
		{
			name: "exp as non-numeric value",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimExp: "not-a-number", // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.ExpiresAt)
			},
		},
		{
			name: "nbf as float64",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimNbf: float64(1699990000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.NotBefore)
				require.Equal(t, int64(1699990000), claims.NotBefore.Unix())
			},
		},
		{
			name: "nbf as json.Number",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimNbf: json.Number("1699990000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.NotBefore)
				require.Equal(t, int64(1699990000), claims.NotBefore.Unix())
			},
		},
		{
			name: "nbf as non-numeric value",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimNbf: true, // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.NotBefore)
			},
		},
		{
			name: "iat as float64",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimIat: float64(1699980000),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.IssuedAt)
				require.Equal(t, int64(1699980000), claims.IssuedAt.Unix())
			},
		},
		{
			name: "iat as json.Number",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimIat: json.Number("1699980000"),
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.NotNil(t, claims.IssuedAt)
				require.Equal(t, int64(1699980000), claims.IssuedAt.Unix())
			},
		},
		{
			name: "iat as non-numeric value",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimIat: []string{"array"}, // Should be float64 or json.Number.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Nil(t, claims.IssuedAt)
			},
		},
		{
			name: "jti as string",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimJti: "test-jti",
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "test-jti", claims.JTI)
			},
		},
		{
			name: "jti as non-string",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimJti: 12345, // Not a string.
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Empty(t, claims.JTI)
			},
		},
		{
			name: "custom claims are preserved",
			claimsMap: map[string]any{
				"custom_string": "value",
				"custom_int":    cryptoutilSharedMagic.AnswerToLifeUniverseEverything,
				"custom_bool":   true,
			},
			verify: func(t *testing.T, claims *JWTClaims) {
				require.Equal(t, "value", claims.Custom["custom_string"])
				require.Equal(t, cryptoutilSharedMagic.AnswerToLifeUniverseEverything, claims.Custom["custom_int"])
				require.Equal(t, true, claims.Custom["custom_bool"])
			},
		},
		{
			name: "all claims combined",
			claimsMap: map[string]any{
				cryptoutilSharedMagic.ClaimIss:    "combined-issuer",
				cryptoutilSharedMagic.ClaimSub:    "combined-subject",
				cryptoutilSharedMagic.ClaimAud:    []any{"aud1", "aud2"},
				cryptoutilSharedMagic.ClaimExp:    json.Number("1700000000"),
				cryptoutilSharedMagic.ClaimNbf:    float64(1699990000),
				cryptoutilSharedMagic.ClaimIat:    float64(1699980000),
				cryptoutilSharedMagic.ClaimJti:    "combined-jti",
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
