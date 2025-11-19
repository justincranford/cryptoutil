// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"fmt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// TLSClientAuthenticator implements TLS client certificate authentication.
type TLSClientAuthenticator struct {
	clientRepo cryptoutilIdentityRepository.ClientRepository
	validator  CertificateValidator
	parser     *CertificateParser
}

// NewTLSClientAuthenticator creates a new TLSClientAuthenticator.
func NewTLSClientAuthenticator(clientRepo cryptoutilIdentityRepository.ClientRepository, validator CertificateValidator) *TLSClientAuthenticator {
	return &TLSClientAuthenticator{
		clientRepo: clientRepo,
		validator:  validator,
		parser:     &CertificateParser{},
	}
}

// Method returns the authentication method name.
func (t *TLSClientAuthenticator) Method() string {
	return string(cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth)
}

// Authenticate authenticates a client using TLS client certificate.
func (t *TLSClientAuthenticator) Authenticate(ctx context.Context, clientID, credential string) (*cryptoutilIdentityDomain.Client, error) {
	// credential contains the PEM-encoded client certificate chain.
	if credential == "" {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// Parse the certificate chain from PEM
	certs, err := t.parser.ParsePEMCertificateChain([]byte(credential))
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate chain: %w", err)
	}

	if len(certs) == 0 {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// The first certificate is the client certificate
	clientCert := certs[0]

	// Validate the certificate
	rawCerts := make([][]byte, len(certs))
	for i, cert := range certs {
		rawCerts[i] = cert.Raw
	}

	if err := t.validator.ValidateCertificate(clientCert, rawCerts); err != nil {
		return nil, fmt.Errorf("certificate validation failed: %w", err)
	}

	// Fetch client from database.
	client, err := t.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Validate client authentication method.
	if !t.validateAuthMethod(client) {
		return nil, cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// TODO: Optionally validate that the certificate subject matches the client
	// This could be done by storing certificate fingerprints or subject info in the client record

	return client, nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (t *TLSClientAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth
}
