package clientauth

import (
	"crypto/x509"

	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// Registry manages client authentication methods.
type Registry struct {
	authenticators map[string]ClientAuthenticator
}

// NewRegistry creates a new authentication method registry.
func NewRegistry(repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *Registry {
	clientRepo := repoFactory.ClientRepository()

	// Create certificate validators
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		// Fallback to empty pool if system certs can't be loaded
		systemCertPool = x509.NewCertPool()
	}

	caValidator := NewCACertificateValidator(systemCertPool)
	// For self-signed, start with empty pinned certificates (would be configured per deployment)
	selfSignedValidator := NewSelfSignedCertificateValidator(make(map[string]*x509.Certificate))

	return &Registry{
		authenticators: map[string]ClientAuthenticator{
			"client_secret_basic":         NewBasicAuthenticator(clientRepo),
			"client_secret_post":          NewPostAuthenticator(clientRepo),
			"tls_client_auth":             NewTLSClientAuthenticator(clientRepo, caValidator),
			"self_signed_tls_client_auth": NewSelfSignedAuthenticator(clientRepo, selfSignedValidator),
		},
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
