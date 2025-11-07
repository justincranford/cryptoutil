package clientauth

import (
	"context"
	"crypto/x509"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TLSClientAuthenticator implements TLS client certificate authentication.
type TLSClientAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
}

// NewTLSClientAuthenticator creates a new TLSClientAuthenticator.
func NewTLSClientAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository) *TLSClientAuthenticator {
	return &TLSClientAuthenticator{
		clientRepo: clientRepo,
	}
}

// Method returns the authentication method name.
func (t *TLSClientAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth)
}

// Authenticate authenticates a client using TLS client certificate.
func (t *TLSClientAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential contains the PEM-encoded client certificate.
	// TODO: Parse PEM certificate and validate.
	_ = credential

	// Fetch client from database.
	client, err := t.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client authentication method.
	if !t.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// TODO: Validate certificate chain against stored certificates.
	// TODO: Check certificate revocation status.

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (t *TLSClientAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth
}

// ValidateCertificate validates a client certificate against stored certificates.
func (t *TLSClientAuthenticator) ValidateCertificate(clientCert *x509.Certificate, storedCerts []*x509.Certificate) error {
	// TODO: Implement certificate validation logic.
	// - Verify certificate is not expired
	// - Verify certificate signature
	// - Check certificate against stored certificates (pinning)
	// - Validate certificate extensions
	_ = clientCert
	_ = storedCerts

	return fmt.Errorf("certificate validation not yet implemented")
}
