// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// SecretHasher provides bcrypt hashing for client secrets.
type SecretHasher interface {
	HashSecret(plaintext string) (string, error)
	CompareSecret(hashed, plaintext string) error
}

// BcryptHasher implements SecretHasher using bcrypt algorithm.
type BcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new bcrypt hasher with default cost.
func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{
		cost: bcrypt.DefaultCost, // Cost 10 (2^10 = 1024 iterations).
	}
}

// HashSecret hashes a plaintext client secret using bcrypt.
func (h *BcryptHasher) HashSecret(plaintext string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash client secret: %w", err)
	}

	return string(hashed), nil
}

// CompareSecret compares a hashed secret with a plaintext secret.
func (h *BcryptHasher) CompareSecret(hashed, plaintext string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)); err != nil {
		return fmt.Errorf("client secret mismatch: %w", err)
	}

	return nil
}

// MigrateClientSecrets migrates all client secrets from plaintext to bcrypt hashes.
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

		// Check if secret is already hashed (bcrypt hashes start with "$2a$" or "$2b$").
		if isBcryptHash(client.ClientSecret) {
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

// isBcryptHash checks if a string is a bcrypt hash.
func isBcryptHash(secret string) bool {
	// Bcrypt hashes have format: $2a$cost$salt+hash (60 characters).
	// Example: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy.
	return len(secret) == 60 && (secret[:4] == "$2a$" || secret[:4] == "$2b$" || secret[:4] == "$2y$")
}

// SecretBasedAuthenticator provides client authentication with bcrypt secret hashing.
type SecretBasedAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
	hasher     SecretHasher
}

// NewSecretBasedAuthenticator creates a new client authenticator with bcrypt hashing.
func NewSecretBasedAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *SecretBasedAuthenticator {
	return &SecretBasedAuthenticator{
		clientRepo: clientRepo,
		hasher:     NewBcryptHasher(),
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

	// Compare client secret using bcrypt.
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
