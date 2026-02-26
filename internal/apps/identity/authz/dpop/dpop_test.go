// Copyright (c) 2025 Justin Cranford

package dpop

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	sha256 "crypto/sha256"
	"encoding/base64"
	json "encoding/json"
	"strings"
	"testing"
	"time"

	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
)

func TestValidateProof(t *testing.T) {
	t.Parallel()

	// Generate test key (ES256).
	alg := joseJwa.ES256()
	_, privateKey, publicKey, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	// Compute JWK thumbprint (same way as dpop.go does).
	jwkJSON, err := json.Marshal(publicKey)
	require.NoError(t, err)

	hash := sha256.Sum256(jwkJSON)
	expectedThumbprint := base64.RawURLEncoding.EncodeToString(hash[:])

	tests := []struct {
		name          string
		dpopBuilder   func() string
		httpMethod    string
		httpURI       string
		accessToken   string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid DPoP proof",
			dpopBuilder: func() string {
				return buildValidProof(t, privateKey, publicKey, "POST", "https://server.example.com/token", "")
			},
			httpMethod:  "POST",
			httpURI:     "https://server.example.com/token",
			accessToken: "",
			expectError: false,
		},
		{
			name: "valid DPoP proof with access token",
			dpopBuilder: func() string {
				accessToken := "test-access-token"

				return buildValidProof(t, privateKey, publicKey, "GET", "https://server.example.com/resource", accessToken)
			},
			httpMethod:  "GET",
			httpURI:     "https://server.example.com/resource",
			accessToken: "test-access-token",
			expectError: false,
		},
		{
			name: "empty DPoP header",
			dpopBuilder: func() string {
				return ""
			},
			httpMethod:    "POST",
			httpURI:       "https://server.example.com/token",
			expectError:   true,
			errorContains: "DPoP header is required",
		},
		{
			name: "htm mismatch",
			dpopBuilder: func() string {
				return buildValidProof(t, privateKey, publicKey, "GET", "https://server.example.com/token", "")
			},
			httpMethod:    "POST",
			httpURI:       "https://server.example.com/token",
			expectError:   true,
			errorContains: "htm claim must match HTTP method",
		},
		{
			name: "htu mismatch",
			dpopBuilder: func() string {
				return buildValidProof(t, privateKey, publicKey, "POST", "https://other.example.com/token", "")
			},
			httpMethod:    "POST",
			httpURI:       "https://server.example.com/token",
			expectError:   true,
			errorContains: "htu claim must match HTTP URI",
		},
		{
			name: "ath claim mismatch",
			dpopBuilder: func() string {
				return buildValidProof(t, privateKey, publicKey, "GET", "https://server.example.com/resource", "wrong-token")
			},
			httpMethod:    "GET",
			httpURI:       "https://server.example.com/resource",
			accessToken:   "correct-token",
			expectError:   true,
			errorContains: "ath claim does not match access token",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			dpopHeader := testCase.dpopBuilder()
			proof, err := ValidateProof(dpopHeader, testCase.httpMethod, testCase.httpURI, testCase.accessToken)

			if testCase.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.errorContains)
				require.Nil(t, proof)
			} else {
				require.NoError(t, err)
				require.NotNil(t, proof)
				require.NotEmpty(t, proof.JTI)
				require.Equal(t, strings.ToUpper(testCase.httpMethod), strings.ToUpper(proof.HTM))
				require.Equal(t, testCase.httpURI, proof.HTU)
				require.Equal(t, expectedThumbprint, proof.JWKThumbprint)
				require.WithinDuration(t, time.Now().UTC(), proof.IAT, cryptoutilSharedMagic.IdentityDefaultIdleTimeoutSeconds*time.Second)
			}
		})
	}
}

func TestComputeAccessTokenHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		accessToken string
		expected    string
	}{
		{
			name:        "empty token",
			accessToken: "",
			expected:    "47DEQpj8HBSa-_TImW-5JCeuQeRkm5NMpJWZG3hSuFU",
		},
		{
			name:        "simple token",
			accessToken: "test-token",
			expected:    computeExpectedHash("test-token"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := ComputeAccessTokenHash(testCase.accessToken)
			require.Equal(t, testCase.expected, result)
		})
	}
}

func TestIsDPoPBound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tokenBuilder func() string
		expectBound  bool
		expectedJKT  string
		expectError  bool
	}{
		{
			name: "DPoP-bound token",
			tokenBuilder: func() string {
				return buildAccessTokenWithCnf(t, "test-thumbprint")
			},
			expectBound: true,
			expectedJKT: "test-thumbprint",
			expectError: false,
		},
		{
			name: "non-DPoP-bound token",
			tokenBuilder: func() string {
				return buildAccessTokenWithoutCnf(t)
			},
			expectBound: false,
			expectedJKT: "",
			expectError: false,
		},
		{
			name: "invalid token",
			tokenBuilder: func() string {
				return "invalid-jwt"
			},
			expectBound: false,
			expectedJKT: "",
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			token := testCase.tokenBuilder()
			bound, jkt, err := IsDPoPBound(token)

			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectBound, bound)
				require.Equal(t, testCase.expectedJKT, jkt)
			}
		})
	}
}

// Helper functions for building test DPoP proofs and access tokens.

func buildValidProof(t *testing.T, privateKey, publicKey joseJwk.Key, httpMethod, httpURI, accessToken string) string {
	t.Helper()

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimJti, "unique-jti-value"))
	require.NoError(t, token.Set("htm", httpMethod))
	require.NoError(t, token.Set("htu", httpURI))
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimIat, time.Now().UTC().Unix()))

	if accessToken != "" {
		ath := ComputeAccessTokenHash(accessToken)
		require.NoError(t, token.Set("ath", ath))
	}

	headers := joseJws.NewHeaders()
	require.NoError(t, headers.Set("typ", "dpop+jwt"))
	require.NoError(t, headers.Set("alg", joseJwa.ES256()))
	require.NoError(t, headers.Set("jwk", publicKey))

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey, joseJws.WithProtectedHeaders(headers)))
	require.NoError(t, err)

	return string(signed)
}

func buildAccessTokenWithCnf(t *testing.T, jkt string) string {
	t.Helper()

	alg := joseJwa.ES256()
	_, privateKey, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimSub, "test-subject"))
	require.NoError(t, token.Set("cnf", map[string]any{"jkt": jkt}))

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey))
	require.NoError(t, err)

	return string(signed)
}

func buildAccessTokenWithoutCnf(t *testing.T) string {
	t.Helper()

	alg := joseJwa.ES256()
	_, privateKey, _, _, _, err := cryptoutilSharedCryptoJose.GenerateJWSJWKForAlg(&alg)
	require.NoError(t, err)

	token := joseJwt.New()
	require.NoError(t, token.Set(cryptoutilSharedMagic.ClaimSub, "test-subject"))

	signed, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.ES256(), privateKey))
	require.NoError(t, err)

	return string(signed)
}

func computeExpectedHash(input string) string {
	hash := sha256.Sum256([]byte(input))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}
