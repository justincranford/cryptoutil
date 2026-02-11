// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// SecretHasher provides FIPS 140-3 approved hashing for client secrets.
type SecretHasher interface {
	HashLowEntropyNonDeterministic(plaintext string) (string, error)
	CompareSecret(hashed, plaintext string) error
}

// RotationService provides client secret rotation functionality.
type RotationService interface {
	ValidateSecretDuringGracePeriod(ctx context.Context, clientID googleUuid.UUID, secretPlaintext string) (bool, int, error)
}

// Legacy comment - implementation replaced with FIPS 140-3 approved PBKDF2-HMAC-SHA256.

// MigrateClientSecrets migrates all client secrets from plaintext to PBKDF2-HMAC-SHA256 hashes.
func MigrateClientSecrets(ctx context.Context, clientRepo cryptoutilIdentityRepository.ClientRepository, hasher SecretHasher) (int, error) {
	// Fetch all clients.
	clients, err := clientRepo.GetAll(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch clients for migration: %w", err)
	}

	migrated := 0

	for _, client := range clients {
		// Skip clients with no secret (public clients, mTLS clients).
		if client.ClientSecret == "" {
			continue
		}

		// Check if secret is already hashed (PBKDF2 hashes start with "$pbkdf2-sha256$").
		if isPBKDF2Hash(client.ClientSecret) {
			continue
		}

		// Hash plaintext secret.
		hashedSecret, err := hasher.HashLowEntropyNonDeterministic(client.ClientSecret)
		if err != nil {
			return migrated, fmt.Errorf("failed to hash secret for client %s: %w", client.ClientID, err)
		}

		// Update client with hashed secret.
		client.ClientSecret = hashedSecret

		if err := clientRepo.Update(ctx, client); err != nil {
			return migrated, fmt.Errorf("failed to update client %s with hashed secret: %w", client.ClientID, err)
		}

		migrated++
	}

	return migrated, nil
}

// isPBKDF2Hash checks if a string is a PBKDF2-HMAC-SHA256 hash.
func isPBKDF2Hash(secret string) bool {
	// PBKDF2 hashes have format: $pbkdf2-sha256$iterations$salt$hash.
	return len(secret) > len("$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$") && secret[:len("$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$")] == "$"+cryptoutilSharedMagic.PBKDF2DefaultHashName+"$"
}

// SecretBasedAuthenticator provides client authentication with FIPS 140-3 approved PBKDF2-HMAC-SHA256 hashing.
type SecretBasedAuthenticator struct {
	clientRepo      cryptoutilIdentityRepository.ClientRepository
	hasher          SecretHasher
	rotationService RotationService
}

// NewSecretBasedAuthenticator creates a new client authenticator with PBKDF2-HMAC-SHA256 hashing.
func NewSecretBasedAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository, rotationService RotationService) *SecretBasedAuthenticator {
	return &SecretBasedAuthenticator{
		clientRepo:      clientRepo,
		hasher:          NewPBKDF2Hasher(),
		rotationService: rotationService,
	}
}

// AuthenticateBasic authenticates a client using HTTP Basic authentication.
func (a *SecretBasedAuthenticator) AuthenticateBasic(ctx context.Context, clientID, clientSecret string) (*cryptoutilIdentityDomain.Client, error) {
	// Fetch client from repository.
	client, err := a.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("client authentication failed: %w", err)
	}

	// Validate client is enabled.
	if client.Enabled == nil || !*client.Enabled {
		return nil, fmt.Errorf("client is disabled")
	}

	// Multi-secret validation: check all active secrets during grace period.
	if a.rotationService != nil {
		valid, version, err := a.rotationService.ValidateSecretDuringGracePeriod(ctx, client.ID, clientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to validate secret: %w", err)
		}

		if valid {
			// Log which version was used for audit purposes.
			_ = version // TODO: Add audit logging for secret version usage

			return client, nil
		}

		return nil, fmt.Errorf("invalid client credentials")
	}

	// Fallback: single-secret validation for backward compatibility.
	if err := a.hasher.CompareSecret(client.ClientSecret, clientSecret); err != nil {
		return nil, fmt.Errorf("invalid client credentials: %w", err)
	}

	return client, nil
}

// AuthenticatePost authenticates a client using POST body credentials.
func (a *SecretBasedAuthenticator) AuthenticatePost(ctx context.Context, clientID, clientSecret string) (*cryptoutilIdentityDomain.Client, error) {
	// Same logic as Basic auth (different HTTP transport, same validation).
	return a.AuthenticateBasic(ctx, clientID, clientSecret)
}

// Authenticate implements ClientAuthenticator interface - delegates to AuthenticateBasic.
func (a *SecretBasedAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	return a.AuthenticateBasic(ctx, clientID, credential)
}

// Method implements ClientAuthenticator interface - returns basic auth method (registry handles mapping).
func (a *SecretBasedAuthenticator) Method() string {
	return "client_secret_basic"
}

// MigrateSecrets wraps MigrateClientSecrets for use through registry.
func (a *SecretBasedAuthenticator) MigrateSecrets(ctx context.Context, clientRepo cryptoutilIdentityRepository.ClientRepository) (int, error) {
	return MigrateClientSecrets(ctx, clientRepo, a.hasher)
}
