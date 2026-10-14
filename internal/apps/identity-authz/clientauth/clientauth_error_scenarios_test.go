// Copyright (c) 2025-2026 Justin Cranford.
package clientauth

import (
	"context"
	"errors"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// makeValidatorClient creates a minimal client for validateClaims tests.
func makeValidatorClient() *cryptoutilIdentityDomain.Client {
	return &cryptoutilIdentityDomain.Client{ClientID: testTestClient}
}

func TestPrivateKeyJWTValidator_ValidateClaims_InvalidSubject(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, "different-client")) // wrong subject

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid subject")
}

func TestPrivateKeyJWTValidator_ValidateClaims_MissingAudience(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, client.ClientID))
	// Intentionally no audience

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "missing audience claim")
}

func TestPrivateKeyJWTValidator_ValidateClaims_MissingExpiration(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	// Intentionally no exp

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "missing expiration claim")
}

func TestPrivateKeyJWTValidator_ValidateClaims_ExpiredToken(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, tok.Set(joseJwt.ExpirationKey, time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute)))

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT expired at")
}

func TestPrivateKeyJWTValidator_ValidateClaims_MissingIssuedAt(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, tok.Set(joseJwt.ExpirationKey, time.Now().UTC().Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	// Intentionally no iat

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "missing issued at claim")
}

func TestPrivateKeyJWTValidator_ValidateClaims_FutureIssuedAt(t *testing.T) {
	t.Parallel()

	validator := &PrivateKeyJWTValidator{
		expectedAudience: testTokenEndpointURL,
	}
	client := makeValidatorClient()

	now := time.Now().UTC()
	tok := joseJwt.New()
	require.NoError(t, tok.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, tok.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, tok.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials*time.Minute)))
	require.NoError(t, tok.Set(joseJwt.IssuedAtKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))

	err := validator.validateClaims(context.Background(), tok, client)
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT issued in the future at")
}

// --- AuthenticateBasic coverage ---

// mockSecretHasher is a mock SecretHasher for testing AuthenticateBasic.
type mockSecretHasher struct {
	hashFn    func(plaintext string) (string, error)
	compareFn func(hashed, plaintext string) error
}

func (m *mockSecretHasher) HashLowEntropyNonDeterministic(plaintext string) (string, error) {
	return m.hashFn(plaintext)
}

func (m *mockSecretHasher) CompareSecret(hashed, plaintext string) error {
	return m.compareFn(hashed, plaintext)
}

func TestAuthenticateBasic_ClientRepoError(t *testing.T) {
	t.Parallel()

	auth := &SecretBasedAuthenticator{
		clientRepo: &mockErrorClientRepo{},
	}

	client, err := auth.AuthenticateBasic(context.Background(), testTestClient, "secret")
	require.Error(t, err)
	require.ErrorContains(t, err, "client authentication failed")
	require.Nil(t, client)
}

func TestAuthenticateBasic_ClientDisabled(t *testing.T) {
	t.Parallel()

	disabled := false
	auth := &SecretBasedAuthenticator{
		clientRepo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				testTestClient: {ClientID: testTestClient, Enabled: &disabled},
			},
		},
	}

	client, err := auth.AuthenticateBasic(context.Background(), testTestClient, "secret")
	require.Error(t, err)
	require.ErrorContains(t, err, "client is disabled")
	require.Nil(t, client)
}

func TestAuthenticateBasic_NilEnabled(t *testing.T) {
	t.Parallel()

	auth := &SecretBasedAuthenticator{
		clientRepo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				testTestClient: {ClientID: testTestClient, Enabled: nil},
			},
		},
	}

	client, err := auth.AuthenticateBasic(context.Background(), testTestClient, "secret")
	require.Error(t, err)
	require.ErrorContains(t, err, "client is disabled")
	require.Nil(t, client)
}

func TestAuthenticateBasic_InvalidHashedSecret(t *testing.T) {
	t.Parallel()

	enabled := true
	auth := &SecretBasedAuthenticator{
		clientRepo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				testTestClient: {
					ClientID:     testTestClient,
					Enabled:      &enabled,
					ClientSecret: "wrong-hashed-secret",
				},
			},
		},
		hasher: &mockSecretHasher{
			compareFn: func(_, _ string) error {
				return errors.New("secret mismatch")
			},
		},
	}

	client, err := auth.AuthenticateBasic(context.Background(), testTestClient, testWrongSecret)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid client credentials")
	require.Nil(t, client)
}

func TestAuthenticateBasic_ValidSecret(t *testing.T) {
	t.Parallel()

	enabled := true
	auth := &SecretBasedAuthenticator{
		clientRepo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				testTestClient: {
					ClientID:     testTestClient,
					Enabled:      &enabled,
					ClientSecret: "hashed-secret",
				},
			},
		},
		hasher: &mockSecretHasher{
			compareFn: func(_, _ string) error { return nil },
		},
	}

	client, err := auth.AuthenticateBasic(context.Background(), testTestClient, testCorrectSecret)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, testTestClient, client.ClientID)
}
