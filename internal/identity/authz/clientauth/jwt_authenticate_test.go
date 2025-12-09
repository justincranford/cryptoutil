// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// createTestClientWithSecret creates a client with client_secret for testing.
func createTestClientWithSecret(t *testing.T, clientRepo cryptoutilIdentityRepository.ClientRepository, clientID, secret string) *cryptoutilIdentityDomain.Client {
	t.Helper()

	client := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.New(),
		ClientID:     clientID,
		ClientSecret: secret, // Store raw secret for HMAC validation.
		Name:         "Test Client " + clientID,
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	return client
}

// createTestClientWithJWKs creates a client with JWKs for testing.
func createTestClientWithJWKs(t *testing.T, clientRepo cryptoutilIdentityRepository.ClientRepository, clientID, jwks string) *cryptoutilIdentityDomain.Client {
	t.Helper()

	client := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.New(),
		ClientID:     clientID,
		JWKs:         jwks,
		Name:         "Test Client " + clientID,
		RedirectURIs: []string{"https://example.com/callback"},
	}

	err := clientRepo.Create(context.Background(), client)
	require.NoError(t, err)

	return client
}

// TestClientSecretJWTAuthenticator_Authenticate_Success tests successful client_secret_jwt authentication.
func TestClientSecretJWTAuthenticator_Authenticate_Success(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }()

	clientRepo := repoFactory.ClientRepository()

	// Create test client with client secret.
	testClient := createTestClientWithSecret(t, clientRepo, "test-secret-jwt-client", testClientSecret)

	// Create authenticator.
	auth := NewClientSecretJWTAuthenticator(testTokenEndpointURL, clientRepo)

	// Create valid JWT assertion signed with client secret (HMAC).
	now := time.Now()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Hour)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Create symmetric key from client secret.
	hmacKey := []byte(testClientSecret)
	jwtAssertion, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), hmacKey))
	require.NoError(t, err)

	// Authenticate using JWT assertion.
	authenticatedClient, err := auth.Authenticate(ctx, string(jwtAssertion), "")
	require.NoError(t, err)
	require.NotNil(t, authenticatedClient)
	require.Equal(t, testClient.ClientID, authenticatedClient.ClientID)
}

// TestClientSecretJWTAuthenticator_Authenticate_InvalidSignature tests failed authentication with wrong secret.
func TestClientSecretJWTAuthenticator_Authenticate_InvalidSignature(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }()

	clientRepo := repoFactory.ClientRepository()

	// Create test client with one secret.
	testClient := createTestClientWithSecret(t, clientRepo, "test-secret-jwt-invalid", testClientSecret)

	// Create authenticator.
	auth := NewClientSecretJWTAuthenticator(testTokenEndpointURL, clientRepo)

	// Create JWT assertion signed with WRONG secret.
	now := time.Now()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Hour)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign with WRONG secret.
	wrongSecret := []byte("wrong-secret-for-hmac-sha256-validation")
	jwtAssertion, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), wrongSecret))
	require.NoError(t, err)

	// Authenticate should fail.
	_, err = auth.Authenticate(ctx, string(jwtAssertion), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT validation failed")
}

// TestPrivateKeyJWTAuthenticator_Authenticate_Success tests successful private_key_jwt authentication.
func TestPrivateKeyJWTAuthenticator_Authenticate_Success(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }()

	clientRepo := repoFactory.ClientRepository()

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

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

	// Create test client with JWK set.
	testClient := createTestClientWithJWKs(t, clientRepo, "test-private-key-jwt", string(publicKeySetBytes))

	// Create authenticator.
	auth := NewPrivateKeyJWTAuthenticator(testTokenEndpointURL, clientRepo)

	// Create valid JWT assertion signed with private key.
	now := time.Now()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Hour)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign JWT with private key.
	jwtAssertion, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	// Authenticate using JWT assertion.
	authenticatedClient, err := auth.Authenticate(ctx, string(jwtAssertion), "")
	require.NoError(t, err)
	require.NotNil(t, authenticatedClient)
	require.Equal(t, testClient.ClientID, authenticatedClient.ClientID)
}

// TestPrivateKeyJWTAuthenticator_Authenticate_InvalidSignature tests failed authentication with wrong key.
func TestPrivateKeyJWTAuthenticator_Authenticate_InvalidSignature(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }()

	clientRepo := repoFactory.ClientRepository()

	// Generate RSA key pair for signing.
	keyID := googleUuid.NewString()
	rsaKeyPair1, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey1, ok := rsaKeyPair1.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK1, err := joseJwk.Import(rsaPrivateKey1)
	require.NoError(t, err)
	require.NoError(t, privateJWK1.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK1.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	// Generate DIFFERENT RSA key pair for client's public key set.
	rsaKeyPair2, err := cryptoutilKeyGen.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	rsaPrivateKey2, ok := rsaKeyPair2.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK2, err := joseJwk.Import(rsaPrivateKey2)
	require.NoError(t, err)

	publicJWK2, err := joseJwk.PublicKeyOf(privateJWK2)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK2))

	publicKeySetBytes, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	// Create test client with public key from KEY PAIR 2.
	testClient := createTestClientWithJWKs(t, clientRepo, "test-private-key-invalid", string(publicKeySetBytes))

	// Create authenticator.
	auth := NewPrivateKeyJWTAuthenticator(testTokenEndpointURL, clientRepo)

	// Create JWT assertion signed with DIFFERENT KEY PAIR 1.
	now := time.Now()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, testClient.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Hour)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign with WRONG private key.
	jwtAssertion, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK1))
	require.NoError(t, err)

	// Authenticate should fail due to signature mismatch.
	_, err = auth.Authenticate(ctx, string(jwtAssertion), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT validation failed")
}

// TestPrivateKeyJWTAuthenticator_Authenticate_ClientNotFound tests authentication failure for unknown client.
func TestPrivateKeyJWTAuthenticator_Authenticate_ClientNotFound(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }()

	clientRepo := repoFactory.ClientRepository()

	// Create authenticator.
	auth := NewPrivateKeyJWTAuthenticator(testTokenEndpointURL, clientRepo)

	// Create JWT assertion for non-existent client.
	now := time.Now()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, "non-existent-client-id"))
	require.NoError(t, token.Set(joseJwt.SubjectKey, "non-existent-client-id"))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(time.Hour)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

	// Sign with any key (doesn't matter, client lookup will fail first).
	jwtAssertion, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.HS256(), []byte("any-secret")))
	require.NoError(t, err)

	// Authenticate should fail with client not found.
	_, err = auth.Authenticate(ctx, string(jwtAssertion), "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "client not found")
}
