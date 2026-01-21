// Copyright (c) 2025 Justin Cranford
//
//

// Package clientauth provides client authentication methods for OAuth2/OIDC.
package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"
)

// BasicAuthenticator implements HTTP Basic authentication for OAuth 2.1 clients.
type BasicAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
}

// NewBasicAuthenticator creates a new BasicAuthenticator.
func NewBasicAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *BasicAuthenticator {
	return &BasicAuthenticator{
		clientRepo: clientRepo,
	}
}

// Method returns the authentication method name.
func (b *BasicAuthenticator) Method() string {
	return cryptoutilIdentityMagic.ClientAuthMethodSecretBasic
}

// Authenticate authenticates a client using HTTP Basic authentication.
// The credential parameter should be the plaintext client_secret (not base64-encoded).
func (b *BasicAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential is the plaintext client_secret.
	clientSecret := credential

	// Fetch client from database.
	client, err := b.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	if client == nil {
		return nil, cryptoutilIdentityAppErr.ErrClientNotFound
	}

	// Validate client secret using PBKDF2-HMAC-SHA256 hash comparison.
	// Use cryptoutilCrypto.VerifySecret (format: pbkdf2$iter$salt$hash) instead of
	// clientauth.CompareSecret (format: salt:hash).
	match, err := cryptoutilDigests.VerifySecret(client.ClientSecret, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to compare client secret: %w", err)
	}

	if !match {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientSecret
	}

	// Validate client authentication method.
	if !b.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (b *BasicAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSecretBasic
}
