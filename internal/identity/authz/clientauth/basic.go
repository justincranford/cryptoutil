// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
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
func (b *BasicAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential should be the base64-encoded client_id:client_secret.
	decoded, err := base64.StdEncoding.DecodeString(credential)
	if err != nil {
		return nil, fmt.Errorf("failed to decode basic auth credentials: %w", err)
	}

	// Split into client_id:client_secret.
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	decodedClientID := parts[0]
	clientSecret := parts[1]

	// Validate client_id matches.
	if decodedClientID != clientID {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// Fetch client from database.
	client, err := b.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client secret (TODO: implement proper hash comparison).
	if client.ClientSecret != clientSecret {
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
