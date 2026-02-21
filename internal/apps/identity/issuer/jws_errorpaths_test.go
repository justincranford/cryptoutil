// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"encoding/base64"
	json "encoding/json"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// TestValidateToken_MalformedInputs tests ValidateToken with various malformed JWT inputs.
func TestValidateToken_MalformedInputs(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilIdentityMagic.RSA2048KeySize)
	testify.NoError(t, err)

	jwsIssuer, err := NewJWSIssuerLegacy("test-issuer", rsaKey, "RS256", time.Hour, time.Hour)
	testify.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "wrong number of parts",
			token:   "only.two",
			wantErr: "invalid_token",
		},
		{
			name:    "invalid base64 header",
			token:   "!!!invalid-base64!!!.Y2xhaW1z.c2ln",
			wantErr: "failed to decode header",
		},
		{
			name:    "invalid base64 claims",
			token:   base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`)) + ".!!!invalid-base64!!!.c2ln",
			wantErr: "failed to decode claims",
		},
		{
			name: "invalid JSON claims",
			token: base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`)) +
				"." + base64.RawURLEncoding.EncodeToString([]byte(`not-json`)) +
				"." + base64.RawURLEncoding.EncodeToString([]byte(`sig`)),
			wantErr: "failed to parse claims",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			claims, err := jwsIssuer.ValidateToken(ctx, tc.token)
			testify.Error(t, err)
			testify.Nil(t, claims)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// mustMarshalJSON marshals to JSON or fails the test.
func mustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()

	data, err := json.Marshal(v)
	testify.NoError(t, err)

	return data
}

// TestVerifySignature_LegacyRSA tests signature verification with legacy RSA key.
func TestVerifySignature_LegacyRSA(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilIdentityMagic.RSA2048KeySize)
	testify.NoError(t, err)

	jwsIssuer, err := NewJWSIssuerLegacy("test-issuer", rsaKey, "RS256", time.Hour, time.Hour)
	testify.NoError(t, err)

	ctx := context.Background()

	token, err := jwsIssuer.IssueAccessToken(ctx, map[string]any{
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
		cryptoutilIdentityMagic.ClaimAud: "test-audience",
	})
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.NoError(t, err)
	testify.NotNil(t, claims)
}

// TestVerifySignature_LegacyECDSA tests signature verification with legacy ECDSA key.
func TestVerifySignature_LegacyECDSA(t *testing.T) {
	t.Parallel()

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	testify.NoError(t, err)

	jwsIssuer, err := NewJWSIssuerLegacy("test-issuer", ecKey, "ES256", time.Hour, time.Hour)
	testify.NoError(t, err)

	ctx := context.Background()

	token, err := jwsIssuer.IssueAccessToken(ctx, map[string]any{
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
		cryptoutilIdentityMagic.ClaimAud: "test-audience",
	})
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.NoError(t, err)
	testify.NotNil(t, claims)
}

// TestVerifySignature_NoSigningKey tests verification when no key is configured.
func TestVerifySignature_NoSigningKey(t *testing.T) {
	t.Parallel()

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claimsJSON := base64.RawURLEncoding.EncodeToString(mustMarshalJSON(t, map[string]interface{}{
		"exp": time.Now().UTC().Add(time.Hour).Unix(),
		"iss": "test-issuer",
	}))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	token := header + "." + claimsJSON + "." + sig

	result, err := jwsIssuer.ValidateToken(ctx, token)
	testify.Error(t, err)
	testify.Nil(t, result)
	testify.Contains(t, err.Error(), "no signing key available")
}

// TestBuildJWS_NoSigningKey tests buildJWS when no signing key is available.
func TestBuildJWS_NoSigningKey(t *testing.T) {
	t.Parallel()

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	result, err := jwsIssuer.buildJWS(map[string]interface{}{
		"sub": "test",
		"iss": "test-issuer",
	})
	testify.Error(t, err)
	testify.Empty(t, result)
	testify.Contains(t, err.Error(), "no signing key available")
}

// TestSignJWT_UnsupportedAlgorithm tests signJWT with an unsupported algorithm.
func TestSignJWT_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilIdentityMagic.RSA2048KeySize)
	testify.NoError(t, err)

	sig, err := signJWT("test-signing-input", "UNSUPPORTED", rsaKey)
	testify.Error(t, err)
	testify.Empty(t, sig)
	testify.Contains(t, err.Error(), "unsupported signing algorithm")
}

// TestSignJWT_WrongKeyType tests signJWT with wrong key type for algorithm.
func TestSignJWT_WrongKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		key       any
		wantErr   string
	}{
		{
			name:      "RSA algorithm with ECDSA key",
			algorithm: cryptoutilIdentityMagic.AlgorithmRS256,
			key:       "not-an-rsa-key",
			wantErr:   "expected RSA private key",
		},
		{
			name:      "ECDSA algorithm with RSA key",
			algorithm: cryptoutilIdentityMagic.AlgorithmES256,
			key:       "not-an-ecdsa-key",
			wantErr:   "expected ECDSA private key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sig, err := signJWT("test-signing-input", tc.algorithm, tc.key)
			testify.Error(t, err)
			testify.Empty(t, sig)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestIssueAccessToken_NoSigningKey tests IssueAccessToken when no signing key is available.
func TestIssueAccessToken_NoSigningKey(t *testing.T) {
	t.Parallel()

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	token, err := jwsIssuer.IssueAccessToken(ctx, map[string]any{
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
	})
	testify.Error(t, err)
	testify.Empty(t, token)
}

// TestIssueIDToken_MissingRequiredClaims tests IssueIDToken with missing required claims.
func TestIssueIDToken_MissingRequiredClaims(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilIdentityMagic.RSA2048KeySize)
	testify.NoError(t, err)

	jwsIssuer, err := NewJWSIssuerLegacy("test-issuer", rsaKey, "RS256", time.Hour, time.Hour)
	testify.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		claims  map[string]any
		wantErr string
	}{
		{
			name:    "missing subject",
			claims:  map[string]any{cryptoutilIdentityMagic.ClaimAud: "test-audience"},
			wantErr: cryptoutilIdentityMagic.ClaimSub,
		},
		{
			name:    "missing audience",
			claims:  map[string]any{cryptoutilIdentityMagic.ClaimSub: "test-subject"},
			wantErr: cryptoutilIdentityMagic.ClaimAud,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			token, err := jwsIssuer.IssueIDToken(ctx, tc.claims)
			testify.Error(t, err)
			testify.Empty(t, token)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestVerifySignature_RotationManagerNoKeys tests verification via rotation manager with no valid keys.
func TestVerifySignature_RotationManagerNoKeys(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	mgr.signingKeys = nil

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		keyRotationMgr:   mgr,
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT","kid":"nonexistent"}`))
	claimsJSON := base64.RawURLEncoding.EncodeToString(mustMarshalJSON(t, map[string]interface{}{
		"exp": time.Now().UTC().Add(time.Hour).Unix(),
		"iss": "test-issuer",
	}))
	sig := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	token := header + "." + claimsJSON + "." + sig

	result, err := jwsIssuer.ValidateToken(ctx, token)
	testify.Error(t, err)
	testify.Nil(t, result)
	testify.Contains(t, err.Error(), "no valid verification keys")
}

// TestVerifySignature_RotationManagerWithValidKey tests verification via rotation manager with a valid key.
func TestVerifySignature_RotationManagerWithValidKey(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	err = mgr.RotateSigningKey(context.Background(), cryptoutilIdentityMagic.AlgorithmRS256)
	testify.NoError(t, err)

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		keyRotationMgr:   mgr,
		defaultAlgorithm: cryptoutilIdentityMagic.AlgorithmRS256,
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	token, err := jwsIssuer.IssueAccessToken(ctx, map[string]any{
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
	})
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.NoError(t, err)
	testify.NotNil(t, claims)
}

// TestVerifySignature_RotationManagerECDSA tests verification via rotation manager with ECDSA key.
func TestVerifySignature_RotationManagerECDSA(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	err = mgr.RotateSigningKey(context.Background(), cryptoutilIdentityMagic.AlgorithmES256)
	testify.NoError(t, err)

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		keyRotationMgr:   mgr,
		defaultAlgorithm: cryptoutilIdentityMagic.AlgorithmES256,
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	token, err := jwsIssuer.IssueAccessToken(ctx, map[string]any{
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
	})
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.NoError(t, err)
	testify.NotNil(t, claims)
}

// TestVerifyJWTSignature_WrongKeyType tests verifyJWTSignature with wrong key types.
func TestVerifyJWTSignature_WrongKeyType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		algorithm string
		key       any
		wantErr   string
	}{
		{
			name:      "RS256 with non-RSA key",
			algorithm: cryptoutilIdentityMagic.AlgorithmRS256,
			key:       "not-an-rsa-key",
			wantErr:   "expected RSA public key",
		},
		{
			name:      "ES256 with non-ECDSA key",
			algorithm: cryptoutilIdentityMagic.AlgorithmES256,
			key:       "not-an-ecdsa-key",
			wantErr:   "expected ECDSA public key",
		},
		{
			name:      "unsupported algorithm",
			algorithm: "PS256",
			key:       "any-key",
			wantErr:   "unsupported verification algorithm",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := verifyJWTSignature("test-input", []byte("test-sig"), tc.algorithm, tc.key)
			testify.Error(t, err)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestVerifyJWTSignature_InvalidECDSASignatureLength tests ECDSA signature with wrong length.
func TestVerifyJWTSignature_InvalidECDSASignatureLength(t *testing.T) {
	t.Parallel()

	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	testify.NoError(t, err)

	err = verifyJWTSignature("test-input", []byte("short"), cryptoutilIdentityMagic.AlgorithmES256, &ecKey.PublicKey)
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "invalid ECDSA signature length")
}

// TestValidateToken_ExpiredToken tests expiration checking with a real signed token.
func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Parallel()

	rsaKey, err := rsa.GenerateKey(crand.Reader, cryptoutilIdentityMagic.RSA2048KeySize)
	testify.NoError(t, err)

	jwsIssuer, err := NewJWSIssuerLegacy("test-issuer", rsaKey, "RS256", time.Hour, time.Hour)
	testify.NoError(t, err)

	ctx := context.Background()

	expiredClaims := map[string]any{
		cryptoutilIdentityMagic.ClaimIss: "test-issuer",
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
		cryptoutilIdentityMagic.ClaimExp: float64(time.Now().UTC().Add(-1 * time.Hour).Unix()),
		cryptoutilIdentityMagic.ClaimIat: float64(time.Now().UTC().Add(-2 * time.Hour).Unix()),
	}

	token, err := jwsIssuer.buildJWS(expiredClaims)
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.Error(t, err)
	testify.Nil(t, claims)
	testify.Contains(t, err.Error(), "expired")
}
