// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"crypto/x509"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestClientSecretJWTAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewClientSecretJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository())

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodSecretJWT), auth.Method())
}

func TestClientSecretJWTAuthenticator_Authenticate_MissingAssertion(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewClientSecretJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository())

	_, err := auth.Authenticate(ctx, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing client_assertion parameter")
}

func TestPrivateKeyJWTAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewPrivateKeyJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository())

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT), auth.Method())
}

func TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	auth := NewPrivateKeyJWTAuthenticator("https://example.com/token", repoFactory.ClientRepository())

	_, err := auth.Authenticate(ctx, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing client_assertion parameter")
}

func TestTLSClientAuthenticator_Method(t *testing.T) {
	t.Parallel()

	repoFactory, _ := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewTLSClientAuthenticator(repoFactory.ClientRepository(), validator)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth), auth.Method())
}

func TestTLSClientAuthenticator_Authenticate_MissingCredential(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := setupTestRepository(t)

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

	repoFactory, _ := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewSelfSignedAuthenticator(repoFactory.ClientRepository(), validator)

	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth), auth.Method())
}

func TestSelfSignedAuthenticator_Authenticate_MissingCredential(t *testing.T) {
	t.Parallel()

	repoFactory, ctx := setupTestRepository(t)

	defer func() { _ = repoFactory.Close() }() //nolint:errcheck // Test cleanup

	validator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))
	auth := NewSelfSignedAuthenticator(repoFactory.ClientRepository(), validator)

	clientID := googleUuid.NewString()
	_, err := auth.Authenticate(ctx, clientID, "")
	require.Error(t, err)
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrInvalidClientAuth)
}
