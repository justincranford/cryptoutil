// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// SecretHasher provides FIPS 140-3 approved hashing for client secrets.
type SecretHasher interface {
	HashSecret(plaintext string) (string, error)
	CompareSecret(hashed, plaintext string) error
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
		hashedSecret, err := hasher.HashSecret(client.ClientSecret)
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
	return len(secret) > 16 && secret[:16] == "$pbkdf2-sha256$"
}

// SecretBasedAuthenticator provides client authentication with FIPS 140-3 approved PBKDF2-HMAC-SHA256 hashing.
type SecretBasedAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
	hasher     SecretHasher
}

// NewSecretBasedAuthenticator creates a new client authenticator with PBKDF2-HMAC-SHA256 hashing.
func NewSecretBasedAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *SecretBasedAuthenticator {
	return &SecretBasedAuthenticator{
		clientRepo: clientRepo,
		hasher:     NewPBKDF2Hasher(),
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
	if !client.Enabled {
		return nil, fmt.Errorf("client is disabled")
	}

	// Compare client secret using PBKDF2-HMAC-SHA256.
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

// MigrateSecrets wraps MigrateClientSecrets for use through registry.
func (a *SecretBasedAuthenticator) MigrateSecrets(ctx context.Context, clientRepo cryptoutilIdentityRepository.ClientRepository) (int, error) {
	return MigrateClientSecrets(ctx, clientRepo, a.hasher)
}
