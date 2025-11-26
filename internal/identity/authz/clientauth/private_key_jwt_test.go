package clientauth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

func TestPrivateKeyJWTAuthenticator_Method_Custom(t *testing.T) {
	t.Parallel()

	mockRepo := &mockClientRepository{}
	authenticator := NewPrivateKeyJWTAuthenticator("https://example.com/token", mockRepo)

	method := authenticator.Method()
	require.Equal(t, string(cryptoutilIdentityDomain.ClientAuthMethodPrivateKeyJWT), method)
}

func TestPrivateKeyJWTAuthenticator_Authenticate_MissingAssertion_Custom(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockRepo := &mockClientRepository{}
	authenticator := NewPrivateKeyJWTAuthenticator("https://example.com/token", mockRepo)

	client, err := authenticator.Authenticate(ctx, "", "")

	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "missing client_assertion")
}
