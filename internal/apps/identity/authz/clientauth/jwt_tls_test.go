// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"crypto/x509"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestClientSecretJWTAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewClientSecretJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository(), nil)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodSecretJWT), auth.Method())
}

func TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewClientSecretJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository(), nil)

	_, err := auth.Authenticate(ctx, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing client_assertion parameter")
}

// NOTE: Additional testing of Authenticate method is currently limited due to a design issue.
// The method calls validator.ValidateJWT with an empty client secret on first validation attempt (line 47),
// which always fails because the validator requires a configured client secret (jwt_validator.go line 173).
// This prevents testing of the normal flow (parse JWT → extract claims → fetch client → validate signature).
// The method needs refactoring to support insecure JWT parsing before client lookup.

func TestPrivateKeyJWTAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewPrivateKeyJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository(), nil)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT), auth.Method())
}

func TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewPrivateKeyJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository(), nil)

	_, err := auth.Authenticate(ctx, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing client_assertion parameter")
}

func TestTLSClientAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewTLSClientAuthenticator(repoFactory.ClientRepository(), validator)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth), auth.Method())
}

func TestTLSClientAuthenticator_Authenticate_MissingCredential(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewTLSClientAuthenticator(repoFactory.ClientRepository(), validator)

	clientID := googleUuid.NewString()
	_, err := auth.Authenticate(ctx, clientID, "")
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidClientAuth)
}

func TestSelfSignedAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewSelfSignedAuthenticator(repoFactory.ClientRepository(), validator)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth), auth.Method())
}

func TestSelfSignedAuthenticator_Authenticate_MissingCredential(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := getTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewSelfSignedAuthenticator(repoFactory.ClientRepository(), validator)

	clientID := googleUuid.NewString()
	_, err := auth.Authenticate(ctx, clientID, "")
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidClientAuth)
}
