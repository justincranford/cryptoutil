package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityApperr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// PostAuthenticator implements form-encoded POST authentication for OAuth 2.1 clients.
type PostAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
}

// NewPostAuthenticator creates a new PostAuthenticator.
func NewPostAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *PostAuthenticator {
	return &PostAuthenticator{
		clientRepo: clientRepo,
	}
}

// Method returns the authentication method name.
func (p *PostAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodSecretPost)
}

// Authenticate authenticates a client using form-encoded POST parameters.
func (p *PostAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential is the client_secret from the POST body.
	clientSecret := credential

	// Fetch client from database.
	client, err := p.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client secret (TODO: implement proper hash comparison).
	if client.ClientSecret != clientSecret {
		return nil, cryptoutilIdentityApperr.ErrInvalidClientSecret
	}

	// Validate client authentication method.
	if !p.validateAuthMethod(client) {
		return nil, cryptoutilIdentityApperr.ErrInvalidClientAuth
	}

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (p *PostAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSecretPost
}
