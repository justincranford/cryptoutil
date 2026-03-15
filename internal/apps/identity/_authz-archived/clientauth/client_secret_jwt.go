// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// ClientSecretJWTAuthenticator authenticates clients using JWT signed with client secret.
type ClientSecretJWTAuthenticator struct {
	validator JWTValidator
	repo      cryptoutilIdentityRepository.ClientRepository
}

// NewClientSecretJWTAuthenticator creates a new client secret JWT authenticator.
func NewClientSecretJWTAuthenticator(tokenEndpointURL string, repo cryptoutilIdentityRepository.ClientRepository, jtiRepo cryptoutilIdentityRepository.JTIReplayCacheRepository) *ClientSecretJWTAuthenticator {
	return &ClientSecretJWTAuthenticator{
		validator: NewClientSecretJWTValidator(tokenEndpointURL, jtiRepo),
		repo:      repo,
	}
}

// Method returns the authentication method name.
func (a *ClientSecretJWTAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodSecretJWT)
}

// Authenticate authenticates a client using client_secret_jwt method.
// The clientID parameter contains the client assertion JWT, and credential is ignored.
func (a *ClientSecretJWTAuthenticator) Authenticate(ctx context.Context, clientID, _ string) (*cryptoutilIdentityDomain.Client, error) {
	// clientID parameter actually contains the JWT assertion for this auth method.
	jwtAssertion := clientID

	if jwtAssertion == "" {
		return nil, fmt.Errorf("missing client_assertion parameter")
	}

	// Parse JWT to extract client ID claim before full validation.
	// We need the client ID to fetch the client's secret for HMAC verification.
	token, err := a.validator.ValidateJWT(ctx, jwtAssertion, &cryptoutilIdentityDomain.Client{
		ClientID: "", // Will be extracted from JWT.
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT assertion: %w", err)
	}

	claims, err := a.validator.ExtractClaims(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Fetch client by ID from claims.
	client, err := a.repo.GetByClientID(ctx, claims.Issuer)
	if err != nil {
		return nil, fmt.Errorf("client not found: %w", err)
	}

	// Now validate JWT with client's actual secret.
	_, err = a.validator.ValidateJWT(ctx, jwtAssertion, client)
	if err != nil {
		return nil, fmt.Errorf("JWT validation failed: %w", err)
	}

	return client, nil
}
