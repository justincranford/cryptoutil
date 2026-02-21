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
	"fmt"
	"testing"
	"time"

	testify "github.com/stretchr/testify/require"

	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// failingKeyGenerator implements KeyGenerator and always returns errors.
type failingKeyGenerator struct {
	signingErr    error
	encryptionErr error
}

func (f *failingKeyGenerator) GenerateSigningKey(_ context.Context, _ string) (*SigningKey, error) {
	return nil, f.signingErr
}

func (f *failingKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	return nil, f.encryptionErr
}

// TestGenerateRSASigningKey_InvalidAlgorithm tests invalid RSA algorithm.
func TestGenerateRSASigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateRSASigningKey(ctx, "RS999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid RSA algorithm")
}

// TestGenerateECDSASigningKey_InvalidAlgorithm tests invalid ECDSA algorithm.
func TestGenerateECDSASigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateECDSASigningKey(ctx, "ES999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid ECDSA algorithm")
}

// TestGenerateHMACSigningKey_InvalidAlgorithm tests invalid HMAC algorithm.
func TestGenerateHMACSigningKey_InvalidAlgorithm(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()
	ctx := context.Background()

	key, err := gen.generateHMACSigningKey(ctx, "HS999")
	testify.Error(t, err)
	testify.Nil(t, key)
	testify.Contains(t, err.Error(), "invalid HMAC algorithm")
}

// TestEcdsaCurveName_UnknownCurve tests ecdsaCurveName with unsupported curve.
func TestEcdsaCurveName_UnknownCurve(t *testing.T) {
	t.Parallel()

	// P-224 is a valid curve but not in our switch.
	result := ecdsaCurveName(elliptic.P224())
	testify.Equal(t, "", result)
}

// TestGetPublicKeys_ExpiredAndInvalidKeys tests filtering of expired/invalid keys.
func TestGetPublicKeys_ExpiredAndInvalidKeys(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	// Add an expired key (should be skipped by continue path).
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	testify.NoError(t, err)

	mgr.signingKeys = append(mgr.signingKeys, &SigningKey{
		KeyID:         "expired-key",
		Key:           ecKey,
		Algorithm:     "ES256",
		CreatedAt:     time.Now().UTC().Add(-48 * time.Hour),
		ExpiresAt:     time.Now().UTC().Add(-24 * time.Hour), // Already expired.
		Active:        false,
		ValidForVerif: true,
	})

	// Add a key with ValidForVerif=false (should also be skipped).
	mgr.signingKeys = append(mgr.signingKeys, &SigningKey{
		KeyID:         "not-valid-for-verif",
		Key:           ecKey,
		Algorithm:     "ES256",
		CreatedAt:     time.Now().UTC(),
		ExpiresAt:     time.Now().UTC().Add(24 * time.Hour),
		Active:        false,
		ValidForVerif: false,
	})

	keys := mgr.GetPublicKeys()
	testify.Empty(t, keys)
}

// TestRotateSigningKey_GeneratorFailure tests signing key rotation with failing generator.
func TestRotateSigningKey_GeneratorFailure(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing key generation error"),
		encryptionErr: fmt.Errorf("mock encryption key generation error"),
	}

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), failGen, nil)
	testify.NoError(t, err)

	err = mgr.RotateSigningKey(context.Background(), "RS256")
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate signing key")
}

// TestRotateEncryptionKey_GeneratorFailure tests encryption key rotation with failing generator.
func TestRotateEncryptionKey_GeneratorFailure(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing key generation error"),
		encryptionErr: fmt.Errorf("mock encryption key generation error"),
	}

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), failGen, nil)
	testify.NoError(t, err)

	err = mgr.RotateEncryptionKey(context.Background())
	testify.Error(t, err)
	testify.Contains(t, err.Error(), "failed to generate encryption key")
}

// TestRotateEncryptionKey_MaxKeysExceeded tests encryption key pruning when max keys exceeded.
func TestRotateEncryptionKey_MaxKeysExceeded(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	policy := &KeyRotationPolicy{
		RotationInterval:    time.Hour,
		GracePeriod:         time.Minute,
		MaxActiveKeys:       2,
		AutoRotationEnabled: false,
	}

	mgr, err := NewKeyRotationManager(policy, gen, nil)
	testify.NoError(t, err)

	ctx := context.Background()

	// Rotate 3 times to exceed MaxActiveKeys of 2.
	for i := 0; i < 3; i++ {
		err = mgr.RotateEncryptionKey(ctx)
		testify.NoError(t, err)
	}

	// Should have been pruned to MaxActiveKeys.
	mgr.mu.RLock()
	keyCount := len(mgr.encryptionKeys)
	mgr.mu.RUnlock()

	testify.LessOrEqual(t, keyCount, policy.MaxActiveKeys)
}

// TestStartAutoRotation_WithErrors tests auto rotation continues after errors.
func TestStartAutoRotation_WithErrors(t *testing.T) {
	t.Parallel()

	failGen := &failingKeyGenerator{
		signingErr:    fmt.Errorf("mock signing error"),
		encryptionErr: fmt.Errorf("mock encryption error"),
	}

	policy := &KeyRotationPolicy{
		RotationInterval:    10 * time.Millisecond,
		GracePeriod:         time.Millisecond,
		MaxActiveKeys:       5,
		AutoRotationEnabled: true,
	}

	mgr, err := NewKeyRotationManager(policy, failGen, nil)
	testify.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run auto rotation — it should encounter errors and continue.
	mgr.StartAutoRotation(ctx, "RS256")

	// StartAutoRotation blocks until context is done. If we get here, it worked.
	testify.Error(t, ctx.Err())
}

// partialFailKeyGenerator succeeds for signing but fails for encryption.
type partialFailKeyGenerator struct {
	ProductionKeyGenerator
}

func (p *partialFailKeyGenerator) GenerateEncryptionKey(_ context.Context) (*EncryptionKey, error) {
	return nil, fmt.Errorf("encryption key generation failed")
}

// TestStartAutoRotation_SigningSucceedsEncryptionFails tests the second error continue in auto rotation.
func TestStartAutoRotation_SigningSucceedsEncryptionFails(t *testing.T) {
	t.Parallel()

	policy := &KeyRotationPolicy{
		RotationInterval:    10 * time.Millisecond,
		GracePeriod:         time.Millisecond,
		MaxActiveKeys:       5,
		AutoRotationEnabled: true,
	}

	mgr, err := NewKeyRotationManager(policy, &partialFailKeyGenerator{}, nil)
	testify.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	mgr.StartAutoRotation(ctx, "ES256")
	testify.Error(t, ctx.Err())
}

// TestValidateToken_MalformedInputs tests ValidateToken with various malformed JWT inputs.
func TestValidateToken_MalformedInputs(t *testing.T) {
	t.Parallel()

	// Create a JWS issuer with a real RSA key for verification paths.
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

	// Issue a token and validate it — covers legacy RSA PrivateKey → PublicKey extraction.
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

	// Issue a token and validate it — covers legacy ECDSA PrivateKey → PublicKey extraction.
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

	// Create issuer with no legacy key and no rotation manager.
	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	// Build a fake token that gets past parsing but fails at signature verification.
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

	// Create issuer with no legacy key and no rotation manager.
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

	// Empty the signing keys — no keys for verification.
	mgr.signingKeys = nil

	jwsIssuer := &JWSIssuer{
		issuer:           "test-issuer",
		keyRotationMgr:   mgr,
		defaultAlgorithm: "RS256",
		accessTokenTTL:   time.Hour,
		idTokenTTL:       time.Hour,
	}

	ctx := context.Background()

	// Build a fake token that gets past parsing.
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

// TestEncryptDecryptToken_RoundTrip tests encrypt/decrypt with a valid key.
func TestEncryptDecryptToken_RoundTrip(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	// Rotate to get an active encryption key.
	err = mgr.RotateEncryptionKey(context.Background())
	testify.NoError(t, err)

	jweIssuer := &JWEIssuer{
		keyRotationMgr: mgr,
	}

	ctx := context.Background()
	original := "test-token-data"

	encrypted, err := jweIssuer.EncryptToken(ctx, original)
	testify.NoError(t, err)
	testify.NotEmpty(t, encrypted)
	testify.NotEqual(t, original, encrypted)

	decrypted, err := jweIssuer.DecryptToken(ctx, encrypted)
	testify.NoError(t, err)
	testify.Equal(t, original, decrypted)
}

// TestEncryptToken_NoEncryptionKey tests EncryptToken with no active encryption key.
func TestEncryptToken_NoEncryptionKey(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	// No encryption keys rotated yet.
	jweIssuer := &JWEIssuer{
		keyRotationMgr: mgr,
	}

	ctx := context.Background()

	result, err := jweIssuer.EncryptToken(ctx, "test")
	testify.Error(t, err)
	testify.Empty(t, result)
	testify.Contains(t, err.Error(), "no active encryption key")
}

// TestDecryptToken_InvalidFormat tests DecryptToken with invalid ciphertext format.
func TestDecryptToken_InvalidFormat(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	err = mgr.RotateEncryptionKey(context.Background())
	testify.NoError(t, err)

	jweIssuer := &JWEIssuer{
		keyRotationMgr: mgr,
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "invalid base64 input",
			input:   "!!!not-valid-base64!!!",
			wantErr: "failed to decode base64",
		},
		{
			name:    "ciphertext too short",
			input:   base64.RawURLEncoding.EncodeToString([]byte("ab")),
			wantErr: "failed to create AES cipher",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := jweIssuer.DecryptToken(ctx, tc.input)
			testify.Error(t, err)
			testify.Empty(t, result)
			testify.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestVerifySignature_RotationManagerWithValidKey tests verification via rotation manager with a valid key by ID.
func TestVerifySignature_RotationManagerWithValidKey(t *testing.T) {
	t.Parallel()

	gen := NewProductionKeyGenerator()

	mgr, err := NewKeyRotationManager(DefaultKeyRotationPolicy(), gen, nil)
	testify.NoError(t, err)

	// Rotate to get a signing key.
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

	// Issue and validate — covers rotation manager active signing key + key-by-ID verification paths.
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

	// Rotate to get an ECDSA signing key.
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

	// Issue and validate — covers rotation manager ECDSA paths.
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

	// Provide a signature that's not 64 bytes (ES256 expects r||s, each 32 bytes).
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

	// Manually build a token with a past expiration.
	expiredClaims := map[string]any{
		cryptoutilIdentityMagic.ClaimIss: "test-issuer",
		cryptoutilIdentityMagic.ClaimSub: "test-subject",
		cryptoutilIdentityMagic.ClaimExp: float64(time.Now().UTC().Add(-1 * time.Hour).Unix()),
		cryptoutilIdentityMagic.ClaimIat: float64(time.Now().UTC().Add(-2 * time.Hour).Unix()),
	}

	// Use buildJWS to get a validly-signed token with expired claims.
	token, err := jwsIssuer.buildJWS(expiredClaims)
	testify.NoError(t, err)
	testify.NotEmpty(t, token)

	claims, err := jwsIssuer.ValidateToken(ctx, token)
	testify.Error(t, err)
	testify.Nil(t, claims)
	testify.Contains(t, err.Error(), "expired")
}
