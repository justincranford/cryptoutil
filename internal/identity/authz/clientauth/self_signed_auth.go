package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// SelfSignedAuthenticator implements self-signed TLS client certificate authentication.
type SelfSignedAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
	validator  CertificateValidator
	parser     *CertificateParser
}

// NewSelfSignedAuthenticator creates a new SelfSignedAuthenticator.
func NewSelfSignedAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository, validator CertificateValidator) *SelfSignedAuthenticator {
	return &SelfSignedAuthenticator{
		clientRepo: clientRepo,
		validator:  validator,
		parser:     &CertificateParser{},
	}
}

// Method returns the authentication method name.
func (s *SelfSignedAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth)
}

// Authenticate authenticates a client using self-signed TLS client certificate.
func (s *SelfSignedAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential contains the PEM-encoded self-signed client certificate.
	if credential == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// Parse the certificate from PEM
	certs, err := s.parser.ParsePEMCertificateChain([]byte(credential))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	if len(certs) == 0 {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// For self-signed auth, we only validate the client certificate
	clientCert := certs[0]

	// Validate the self-signed certificate
	rawCerts := make([][]byte, len(certs))
	for i, cert := range certs {
		rawCerts[i] = cert.Raw
	}

	if err := s.validator.ValidateCertificate(clientCert, rawCerts); err != nil {
		return nil, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Fetch client from database.
	client, err := s.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client authentication method.
	if !s.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// TODO: Optionally validate that the certificate fingerprint matches stored client certificate info
	// This could be done by storing certificate fingerprints in the client record

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (s *SelfSignedAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth
}
