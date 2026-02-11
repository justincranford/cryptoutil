// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"crypto/x509"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// Registry manages client authentication methods.
type Registry struct {
	authenticators map[string]ClientAuthenticator
	hasher         *SecretBasedAuthenticator
}

// NewRegistry creates a new client authentication registry.
func NewRegistry(repoFactory *cryptoutilIdentityRepository.RepositoryFactory, config *cryptoutilIdentityConfig.Config, rotationService RotationService) *Registry {
	clientRepo := repoFactory.ClientRepository()
	jtiRepoCache := repoFactory.JTIReplayCacheRepository()

	// Create certificate validators
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		// Fallback to empty pool if system certs can't be loaded
		systemCertPool = x509.NewCertPool()
	}

	// Create combined CRL/OCSP revocation checker
	// OCSP timeout: 5s, CRL timeout: 10s, CRL cache: 1 hour
	revocationChecker := NewCombinedRevocationChecker(
		cryptoutilIdentityMagic.DefaultOCSPTimeout,
		cryptoutilIdentityMagic.DefaultCRLTimeout,
		cryptoutilIdentityMagic.DefaultCRLCacheMaxAge,
	)

	caValidator := NewCACertificateValidator(systemCertPool, revocationChecker)
	// For self-signed, start with empty pinned certificates (would be configured per deployment)
	selfSignedValidator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))

	// Get token issuer URL from config for JWT-based authenticators
	tokenEndpointURL := config.Tokens.Issuer + "/token"

	// Create secret-based authenticator with rotation service support
	secretAuth := NewSecretBasedAuthenticator(clientRepo, rotationService)

	return &Registry{
		authenticators: map[string]ClientAuthenticator{
			"client_secret_basic":         secretAuth,
			"client_secret_post":          secretAuth,
			"tls_client_auth":             NewTLSClientAuthenticator(clientRepo, caValidator),
			"self_signed_tls_client_auth": NewSelfSignedAuthenticator(clientRepo, selfSignedValidator),
			"private_key_jwt":             NewPrivateKeyJWTAuthenticator(tokenEndpointURL, clientRepo, jtiRepoCache),
			"client_secret_jwt":           NewClientSecretJWTAuthenticator(tokenEndpointURL, clientRepo, jtiRepoCache),
		},
		hasher: secretAuth,
	}
}

// GetAuthenticator returns the authenticator for the specified method.
func (r *Registry) GetAuthenticator(method string) (ClientAuthenticator, bool) {
	auth, ok := r.authenticators[method]

	return auth, ok
}

// RegisterAuthenticator registers a new authentication method.
func (r *Registry) RegisterAuthenticator(authenticator ClientAuthenticator) {
	r.authenticators[authenticator.Method()] = authenticator
}

// GetHasher returns the secret-based authenticator for migration operations.
func (r *Registry) GetHasher() *SecretBasedAuthenticator {
	return r.hasher
}
