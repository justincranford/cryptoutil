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

	// Validate certificate subject if configured.
	if err := t.validateCertificateSubject(client, clientCert); err != nil {
		return nil, err
	}

	// Validate certificate fingerprint if configured.
	if err := t.validateCertificateFingerprint(client, clientCert); err != nil {
		return nil, err
	}

	return client, nil
}

// validateCertificateSubject checks if certificate subject matches client registration.
func (t *TLSClientAuthenticator) validateCertificateSubject(client *cryptoutilIdentityDomain.Client, clientCert *x509.Certificate) error {
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
func (t *TLSClientAuthenticator) validateCertificateFingerprint(client *cryptoutilIdentityDomain.Client, clientCert *x509.Certificate) error {
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
func (t *TLSClientAuthenticator) validateAuthMethod(client *cryptoutilIdentityDomain.Client) bool {
	return client.TokenEndpointAuthMethod == cryptoutilIdentityDomain.ClientAuthMethodTLSClientAuth
}
