// Copyright (c) 2025 Justin Cranford

package clientauth

import (
	"context"
	"errors"
	"testing"

	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// mockJWTValidator is a test mock for the JWTValidator interface.
type mockJWTValidator struct {
	validateJWTFn   func(ctx context.Context, jwtString string, client *cryptoutilIdentityDomain.Client) (joseJwt.Token, error)
	extractClaimsFn func(ctx context.Context, token joseJwt.Token) (*ClientClaims, error)
}

func (m *mockJWTValidator) ValidateJWT(ctx context.Context, jwtString string, client *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
	return m.validateJWTFn(ctx, jwtString, client)
}

func (m *mockJWTValidator) ExtractClaims(ctx context.Context, token joseJwt.Token) (*ClientClaims, error) {
	return m.extractClaimsFn(ctx, token)
}

// mockErrorClientRepo returns errors on GetByClientID.
type mockErrorClientRepo struct {
	mockClientRepo
}

func (m *mockErrorClientRepo) GetByClientID(_ context.Context, _ string) (*cryptoutilIdentityDomain.Client, error) {
	return nil, errors.New("database error")
}

// TestClientSecretJWTAuthenticator_EmptyAssertion tests that empty assertion returns error.
func TestClientSecretJWTAuthenticator_EmptyAssertion(t *testing.T) {
	t.Parallel()

	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{},
		repo:      &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "missing client_assertion parameter")
	require.Nil(t, client)
}

// TestClientSecretJWTAuthenticator_ValidateJWTFirstPassFails tests error on first JWT validation.
func TestClientSecretJWTAuthenticator_ValidateJWTFirstPassFails(t *testing.T) {
	t.Parallel()

	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return nil, errors.New("jwt parse error")
			},
		},
		repo: &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to parse JWT assertion")
	require.Nil(t, client)
}

// TestClientSecretJWTAuthenticator_ExtractClaimsFails tests error on claims extraction.
func TestClientSecretJWTAuthenticator_ExtractClaimsFails(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return nil, errors.New("claims extraction error")
			},
		},
		repo: &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to extract claims")
	require.Nil(t, client)
}

// TestClientSecretJWTAuthenticator_ClientNotFound tests error when client is not in repo.
func TestClientSecretJWTAuthenticator_ClientNotFound(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockErrorClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "client not found")
	require.Nil(t, client)
}

// TestClientSecretJWTAuthenticator_ValidateJWTSecondPassFails tests error on second JWT validation.
func TestClientSecretJWTAuthenticator_ValidateJWTSecondPassFails(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	expectedClient := &cryptoutilIdentityDomain.Client{ClientID: "test-client-id"}
	callCount := 0
	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				callCount++
				if callCount == 1 {
					return fakeToken, nil // first pass succeeds
				}

				return nil, errors.New("jwt validation with secret failed") // second pass fails
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				"test-client-id": expectedClient,
			},
		},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT validation failed")
	require.Nil(t, client)
}

// TestClientSecretJWTAuthenticator_Success tests successful authentication.
func TestClientSecretJWTAuthenticator_Success(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	expectedClient := &cryptoutilIdentityDomain.Client{ClientID: "test-client-id"}
	auth := &ClientSecretJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				"test-client-id": expectedClient,
			},
		},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "test-client-id", client.ClientID)
}

// TestPrivateKeyJWTAuthenticator_EmptyAssertion tests that empty assertion returns error.
func TestPrivateKeyJWTAuthenticator_EmptyAssertion(t *testing.T) {
	t.Parallel()

	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{},
		repo:      &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "missing client_assertion parameter")
	require.Nil(t, client)
}

// TestPrivateKeyJWTAuthenticator_ValidateJWTFirstPassFails tests error on first JWT validation.
func TestPrivateKeyJWTAuthenticator_ValidateJWTFirstPassFails(t *testing.T) {
	t.Parallel()

	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return nil, errors.New("jwt parse error")
			},
		},
		repo: &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to parse JWT assertion")
	require.Nil(t, client)
}

// TestPrivateKeyJWTAuthenticator_ExtractClaimsFails tests error on claims extraction.
func TestPrivateKeyJWTAuthenticator_ExtractClaimsFails(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return nil, errors.New("claims extraction error")
			},
		},
		repo: &mockClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to extract claims")
	require.Nil(t, client)
}

// TestPrivateKeyJWTAuthenticator_ClientNotFound tests error when client not found in repo.
func TestPrivateKeyJWTAuthenticator_ClientNotFound(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockErrorClientRepo{},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "client not found")
	require.Nil(t, client)
}

// TestPrivateKeyJWTAuthenticator_ValidateJWTSecondPassFails tests error on second JWT validation.
func TestPrivateKeyJWTAuthenticator_ValidateJWTSecondPassFails(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	expectedClient := &cryptoutilIdentityDomain.Client{ClientID: "test-client-id"}
	callCount := 0
	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				callCount++
				if callCount == 1 {
					return fakeToken, nil
				}

				return nil, errors.New("jwt validation with key failed")
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				"test-client-id": expectedClient,
			},
		},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.Error(t, err)
	require.ErrorContains(t, err, "JWT validation failed")
	require.Nil(t, client)
}

// TestPrivateKeyJWTAuthenticator_Success tests successful authentication.
func TestPrivateKeyJWTAuthenticator_Success(t *testing.T) {
	t.Parallel()

	fakeToken := joseJwt.New()
	expectedClient := &cryptoutilIdentityDomain.Client{ClientID: "test-client-id"}
	auth := &PrivateKeyJWTAuthenticator{
		validator: &mockJWTValidator{
			validateJWTFn: func(_ context.Context, _ string, _ *cryptoutilIdentityDomain.Client) (joseJwt.Token, error) {
				return fakeToken, nil
			},
			extractClaimsFn: func(_ context.Context, _ joseJwt.Token) (*ClientClaims, error) {
				return &ClientClaims{Issuer: "test-client-id"}, nil
			},
		},
		repo: &mockClientRepo{
			clients: map[string]*cryptoutilIdentityDomain.Client{
				"test-client-id": expectedClient,
			},
		},
	}

	client, err := auth.Authenticate(context.Background(), "some.jwt.token", "")
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "test-client-id", client.ClientID)
}
