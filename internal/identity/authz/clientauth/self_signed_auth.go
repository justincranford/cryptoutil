// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	sha256 "crypto/sha256"
	"crypto/x509"
	"encoding/hex"
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

	// Validate certificate subject if configured.
	if err := s.validateCertificateSubject(client, clientCert); err != nil {
		return nil, err
	}

	// Validate certificate fingerprint if configured.
	if err := s.validateCertificateFingerprint(client, clientCert); err != nil {
		return nil, err
	}

	return client, nil
}

// validateCertificateSubject checks if certificate subject matches client registration.
func (s *SelfSignedAuthenticator) validateCertificateSubject(client *cryptoutilIdentityDomain.Client, clientCert *x509.Certificate) error {
	if client.CertificateSubject == "" {
		// No subject validation required.
		return nil
	}

	if clientCert.Subject.CommonName != client.CertificateSubject {
		return fmt.Errorf("certificate subject mismatch: expected %s, got %s", client.CertificateSubject, clientCert.Subject.CommonName)
	}

	return nil
}

// validateCertificateFingerprint checks if certificate fingerprint matches stored value.
func (s *SelfSignedAuthenticator) validateCertificateFingerprint(client *cryptoutilIdentityDomain.Client, clientCert *x509.Certificate) error {
	if client.CertificateFingerprint == "" {
		// No fingerprint validation required.
		return nil
	}

	// Compute SHA-256 fingerprint of certificate.
	hash := sha256.Sum256(clientCert.Raw)
	fingerprint := hex.EncodeToString(hash[:])

	if fingerprint != client.CertificateFingerprint {
		return fmt.Errorf("certificate fingerprint mismatch: expected %s, got %s", client.CertificateFingerprint, fingerprint)
	}

	return nil
}

// validateAuthMethod checks if the client supports this authentication method.
func (s *SelfSignedAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodSelfSignedTLSAuth
}
