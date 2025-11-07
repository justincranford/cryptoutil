package clientauth

import (
	"context"
	"crypto/x509"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// SelfSignedAuthenticator implements self-signed TLS client certificate authentication.
type SelfSignedAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
}

// NewSelfSignedAuthenticator creates a new SelfSignedAuthenticator.
func NewSelfSignedAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *SelfSignedAuthenticator {
	return &SelfSignedAuthenticator{
		clientRepo: clientRepo,
	}
}

// Method returns the authentication method name.
func (s *SelfSignedAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth)
}

// Authenticate authenticates a client using self-signed TLS client certificate.
func (s *SelfSignedAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential contains the PEM-encoded self-signed client certificate.
	// TODO: Parse PEM certificate and validate.
	_ = credential

	// Fetch client from database.
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client authentication method.
	if !s.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// TODO: Validate self-signed certificate against stored certificate (pinning required).

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (s *SelfSignedAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth
}

// ValidateSelfSignedCertificate validates a self-signed certificate against stored certificate.
func (s *SelfSignedAuthenticator) ValidateSelfSignedCertificate(clientCert, pinnedCert *x509.Certificate) error {
	// TODO: Implement self-signed certificate validation logic.
	// - Verify certificate is not expired
	// - Verify certificate matches pinned certificate (exact match or fingerprint)
	// - Validate certificate extensions and subject
	_ = clientCert
	_ = pinnedCert

	return fmt.Errorf("self-signed certificate validation not yet implemented")
}
