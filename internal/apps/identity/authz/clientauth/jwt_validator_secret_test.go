// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestClientSecretJWTValidator_ValidateJWT_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	}

	// Create valid JWT.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Create HMAC key from client secret.
	keyData := []byte(client.ClientSecret)
	key, err := joseJwk.Import(keyData)
	require.NoError(t, err)
	require.NoError(t, key.Set(joseJwk.KeyIDKey, "test-hmac-key"))
	require.NoError(t, key.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))

	// Sign JWT with HMAC key.
	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), key))
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

func TestClientSecretJWTValidator_ValidateJWT_NoClientSecret(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: "", // Empty client secret.
	}

	_, err := validator.ValidateJWT(ctx, "fake-jwt-token", client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "client has no secret configured")
}

func TestClientSecretJWTValidator_ValidateJWT_InvalidSignature(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	}

	// Create JWT with WRONG secret for signing.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign with DIFFERENT secret.
	wrongKeyData := []byte("wrong-client-secret-different-key")
	wrongKey, err := joseJwk.Import(wrongKeyData)
	require.NoError(t, err)
	require.NoError(t, wrongKey.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), wrongKey))
	require.NoError(t, err)

	// Validation should fail due to signature mismatch.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse and verify JWT")
}

func TestClientSecretJWTValidator_ValidateJWT_ExpiredToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
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

	keyData := []byte(client.ClientSecret)
	key, err := joseJwk.Import(keyData)
	require.NoError(t, err)
	require.NoError(t, key.Set(joseJwk.KeyIDKey, "test-hmac-key"))
	require.NoError(t, key.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), key))
	require.NoError(t, err)

	// Validation should fail due to expiration.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token is expired")
}

func TestClientSecretJWTValidator_ValidateJWT_MissingExpirationClaim(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	}

	// Create JWT WITHOUT expiration claim.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	// No expiration set.
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	keyData := []byte(client.ClientSecret)
	key, err := joseJwk.Import(keyData)
	require.NoError(t, err)
	require.NoError(t, key.Set(joseJwk.KeyIDKey, "test-hmac-key"))
	require.NoError(t, key.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), key))
	require.NoError(t, err)

	// Validation should fail due to missing expiration.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing expiration claim")
}

func TestClientSecretJWTValidator_ValidateJWT_MissingIssuedAtClaim(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	}

	// Create JWT WITHOUT issued at claim.
	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	// No issued at set.
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	keyData := []byte(client.ClientSecret)
	key, err := joseJwk.Import(keyData)
	require.NoError(t, err)
	require.NoError(t, key.Set(joseJwk.KeyIDKey, "test-hmac-key"))
	require.NoError(t, key.Set(joseJwk.AlgorithmKey, joseJwa.HS256()))

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), key))
	require.NoError(t, err)

	// Validation should fail due to missing issued at.
	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing issued at claim")
}

func TestClientSecretJWTValidator_ValidateJWT_MalformedJWT(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	client := &cryptoutilIdentityDomain.Client{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	}

	// Pass malformed JWT string.
	malformedJWT := "this.is.not.a.valid.jwt.token.at.all"

	_, err := validator.ValidateJWT(ctx, malformedJWT, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse and verify JWT")
}

func TestPrivateKeyJWTValidator_ExtractClaims_AllClaimsPresent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, nil)

	// Create token with all claims.
	now := time.Now().UTC()
	jti := googleUuid.NewString()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, jti))

	claims, err := validator.ExtractClaims(ctx, token)
	require.NoError(t, err)
	require.Equal(t, testClientID, claims.Issuer)
	require.Equal(t, testClientID, claims.Subject)
	require.Contains(t, claims.Audience, testTokenEndpointURL)
	require.Equal(t, jti, claims.JWTID)
}

func TestClientSecretJWTValidator_ExtractClaims_AllClaimsPresent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	validator := NewClientSecretJWTValidator(testTokenEndpointURL, nil)

	// Create token with all claims.
	now := time.Now().UTC()
	jti := googleUuid.NewString()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, jti))

	claims, err := validator.ExtractClaims(ctx, token)
	require.NoError(t, err)
	require.Equal(t, testClientID, claims.Issuer)
	require.Equal(t, testClientID, claims.Subject)
	require.Contains(t, claims.Audience, testTokenEndpointURL)
	require.Equal(t, jti, claims.JWTID)
}
