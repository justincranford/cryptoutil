// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
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

	if client == nil {
		return nil, cryptoutilIdentityAppErr.ErrClientNotFound
	}

	// Validate client secret using PBKDF2-HMAC-SHA256 hash comparison.
	// Use cryptoutilCrypto.VerifySecret (format: pbkdf2$iter$salt$hash) instead of
	// clientauth.CompareSecret (format: salt:hash).
	match, err := cryptoutilSharedCryptoDigests.VerifySecret(client.ClientSecret, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to compare client secret: %w", err)
	}

	if !match {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientSecret
	}

	// Validate client authentication method.
	if !p.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (p *PostAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSecretPost
}
