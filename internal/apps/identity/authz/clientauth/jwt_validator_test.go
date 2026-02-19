// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	rsa "crypto/rsa"
	json "encoding/json"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
)

// Test constants.
const (
	testTokenEndpointURL = "https://auth.example.com/token"
	testClientID         = "test-client-id"
	testClientSecret     = "test-client-secret-very-long-for-hmac-sha256"
)

func TestPrivateKeyJWTValidator_ValidateJWT_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Generate RSA key pair for testing.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	// Extract *rsa.PrivateKey from KeyPair.
	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	// Create JWK from private key.
	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	// Create public key set for client.
	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     string(publicKeySetBytes),
	}

	// Create valid JWT.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign JWT with private key.
	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Validate JWT.
	validatedToken, err := validator.ValidateJWT(ctx, string(signedToken), client)
	require.NoError(t, err)
	require.NotNil(t, validatedToken)

	// Extract and verify claims.
	claims, err := validator.ExtractClaims(ctx, validatedToken)
	require.NoError(t, err)
	require.Equal(t, testClientID, claims.Issuer)
	require.Equal(t, testClientID, claims.Subject)
	require.Contains(t, claims.Audience, testTokenEndpointURL)
}

func TestPrivateKeyJWTValidator_ValidateJWT_NoJWKSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     "", // Empty JWK set.
	}

	_, err := validator.ValidateJWT(ctx, "fake-jwt-token", client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "client has no JWK set configured")
}

func TestPrivateKeyJWTValidator_ValidateJWT_InvalidJWKSet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     "invalid-json-not-a-jwk-set",
	}

	_, err := validator.ValidateJWT(ctx, "fake-jwt-token", client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse client JWK set")
}

func TestPrivateKeyJWTValidator_ValidateJWT_InvalidSignature(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Generate RSA key pair for signing.
	keyID := googleUuid.NewString()
	rsaKeyPair1, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey1, ok := rsaKeyPair1.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey1)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	// Generate DIFFERENT RSA key pair for client's public key.
	rsaKeyPair2, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey2, ok := rsaKeyPair2.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	differentPrivateJWK, err := joseJwk.Import(rsaPrivateKey2)
	require.NoError(t, err)

	differentPublicJWK, err := joseJwk.PublicKeyOf(differentPrivateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(differentPublicJWK))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     string(publicKeySetBytes),
	}

	// Create valid JWT but sign with DIFFERENT private key.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign with the FIRST private key (not matching client's public key).
	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Validation should fail due to signature mismatch.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse and verify JWT")
}

func TestPrivateKeyJWTValidator_ValidateJWT_ExpiredToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     string(publicKeySetBytes),
	}

	// Create EXPIRED JWT.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(-time.Hour))) // Expired 1 hour ago.
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now.Add(-2*time.Hour)))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Validation should fail due to expiration.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token is expired")
}

func TestPrivateKeyJWTValidator_ValidateJWT_InvalidIssuer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     string(publicKeySetBytes),
	}

	// Create valid JWT with WRONG issuer.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, "wrong-client-id")) // Wrong issuer.
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Validation should fail due to issuer mismatch.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid issuer")
}

func TestPrivateKeyJWTValidator_ValidateJWT_InvalidAudience(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	client := &cryptoutilIdentityDomain.Client{
		ClientID: testClientID,
		JWKs:     string(publicKeySetBytes),
	}

	// Create JWT with WRONG audience.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{"https://wrong.example.com/token"})) // Wrong audience.
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Validation should fail due to audience mismatch.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid audience")
}
