// Copyright (c) 2025 Justin Cranford
//
//

// Package bootstrap provides client bootstrapping and initialization utilities.
package bootstrap

import (
	"context"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

// CreateDemoClient creates the demo-client for testing OAuth flows if it doesn't exist.
// Returns the client ID and plaintext secret (only on creation).
func CreateDemoClient(
	ctx context.Context,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) (clientID string, plainSecret string, created bool, err error) {
	clientRepo := repoFactory.ClientRepository()

	// Check if demo-client already exists.
	const demoClientID = "demo-client"

	existingClient, err := clientRepo.GetByClientID(ctx, demoClientID)
	if err != nil && !errors.Is(err, cryptoutilIdentityAppErr.ErrClientNotFound) {
		return "", "", false, fmt.Errorf("failed to check for existing demo-client: %w", err)
	}

	if existingClient != nil {
		// Client exists, return without secret.
		return demoClientID, "", false, nil
	}

	// Generate demo client secret.
	plainSecret = "demo-secret"

	secretHash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(plainSecret)
	if err != nil {
		return "", "", false, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrPasswordHashFailed,
			fmt.Errorf("failed to hash demo-client secret: %w", err),
		)
	}

	// Create demo client.
	requirePKCE := true
	enabled := true

	demoClient := &cryptoutilIdentityDomain.Client{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     demoClientID,
		ClientSecret: secretHash,
		ClientType:   cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:         "Demo Client",
		Description:  "Bootstrap client for testing OAuth flows",
		RedirectURIs: []string{
			"http://localhost:3000/callback",
			"https://localhost:3000/callback",
		},
		AllowedGrantTypes: []string{
			cryptoutilIdentityMagic.GrantTypeClientCredentials,
			cryptoutilIdentityMagic.GrantTypeAuthorizationCode,
			cryptoutilIdentityMagic.GrantTypeRefreshToken,
		},
		AllowedResponseTypes: []string{
			cryptoutilIdentityMagic.ResponseTypeCode,
		},
		AllowedScopes: []string{
			"openid",
			"profile",
			"email",
			"read",
			"write",
		},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		RequirePKCE:             &requirePKCE,
		PKCEChallengeMethod:     "S256",
		AccessTokenLifetime:     int(cryptoutilIdentityMagic.DefaultAccessTokenLifetime.Seconds()),
		RefreshTokenLifetime:    int(cryptoutilIdentityMagic.DefaultRefreshTokenLifetime.Seconds()),
		IDTokenLifetime:         int(cryptoutilIdentityMagic.DefaultIDTokenLifetime.Seconds()),
		Enabled:                 &enabled,
	}

	if err := clientRepo.Create(ctx, demoClient); err != nil {
		return "", "", false, cryptoutilIdentityAppErr.WrapError(
			cryptoutilIdentityAppErr.ErrDatabaseQuery,
			fmt.Errorf("failed to create demo-client: %w", err),
		)
	}

	return demoClientID, plainSecret, true, nil
}

// BootstrapClients creates all bootstrap clients for the identity server.
func BootstrapClients(
	ctx context.Context,
	_ *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
) error {
	// Create demo client.
	clientID, secret, created, err := CreateDemoClient(ctx, repoFactory)
	if err != nil {
		return fmt.Errorf("failed to bootstrap demo-client: %w", err)
	}

	if created {
		fmt.Printf("✅ Created bootstrap client: %s (secret: %s)\n", clientID, secret)
		fmt.Printf("   Redirect URIs: http://localhost:3000/callback, https://localhost:3000/callback\n")
		fmt.Printf("   Allowed Grants: client_credentials, authorization_code, refresh_token\n")
		fmt.Printf("   Allowed Scopes: openid, profile, email, read, write\n")
		fmt.Printf("   PKCE: Required (S256)\n")
		fmt.Printf("   ⚠️  SAVE THIS SECRET - it will not be shown again!\n")
	} else {
		fmt.Printf("ℹ️  Bootstrap client already exists: %s\n", clientID)
	}

	return nil
}
