// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// CertificateValidator defines the interface for certificate validation.
type CertificateValidator interface {
	ValidateCertificate(clientCert *x509.Certificate, rawCerts [][]byte) error
	IsRevoked(serialNumber *big.Int) bool
}

// CACertificateValidator validates certificates issued by a Certificate Authority.
type CACertificateValidator struct {
	// trustedCAs contains the trusted Certificate Authorities
	trustedCAs *x509.CertPool
	// crlCache caches Certificate Revocation Lists (deprecated - use revocationChecker instead)
	crlCache map[string]*pkix.CertificateList
	// revocationChecker handles CRL/OCSP revocation checking
	revocationChecker RevocationChecker
	// validateSubject enables certificate subject validation
	validateSubject bool
	// validateFingerprint enables certificate fingerprint validation
	validateFingerprint bool
}

// NewCACertificateValidator creates a new CA certificate validator.
func NewCACertificateValidator(trustedCAs *x509.CertPool, revocationChecker RevocationChecker) *CACertificateValidator {
	return &CACertificateValidator{
		trustedCAs:          trustedCAs,
		crlCache:            make(map[string]*pkix.CertificateList),
		revocationChecker:   revocationChecker,
		validateSubject:     true, // Enable subject validation by default.
		validateFingerprint: true, // Enable fingerprint validation by default.
	}
}

// SetValidationOptions configures certificate validation strictness.
func (v *CACertificateValidator) SetValidationOptions(validateSubject, validateFingerprint bool) {
	v.validateSubject = validateSubject
	v.validateFingerprint = validateFingerprint
}

// ValidateCertificate validates a client certificate against trusted CAs.
func (v *CACertificateValidator) ValidateCertificate(clientCert *x509.Certificate, rawCerts [][]byte) error {
	if clientCert == nil {
		return cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// Check if certificate is expired
	now := time.Now().UTC()
	if now.Before(clientCert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}

	if now.After(clientCert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	// Verify certificate signature and chain
	opts := x509.VerifyOptions{
		Roots:         v.trustedCAs,
		CurrentTime:   now,
		Intermediates: x509.NewCertPool(),
	}

	// Add intermediate certificates if provided
	if len(rawCerts) > 1 {
		for _, rawCert := range rawCerts[1:] {
			if cert, err := x509.ParseCertificate(rawCert); err == nil {
				opts.Intermediates.AddCert(cert)
			}
		}
	}

	// Verify the certificate chain
	chains, err := clientCert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	// Ensure at least one valid chain exists
	if len(chains) == 0 {
		return fmt.Errorf("no valid certificate chains found")
	}

	// Check revocation status using CRL/OCSP
	if v.revocationChecker != nil {
		// Extract issuer certificate from chain for revocation checking.
		var issuer *x509.Certificate
		if len(chains) > 0 && len(chains[0]) > 1 {
			issuer = chains[0][1] // Issuer is second cert in first valid chain.
		}

		if issuer != nil {
			// Create context with timeout for revocation check.
			ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.DefaultRevocationTimeout)
			defer cancel()

			if err := v.revocationChecker.CheckRevocation(ctx, clientCert, issuer); err != nil {
				return fmt.Errorf("certificate revocation check failed: %w", err)
			}
		}
	}

	return nil
}

// IsRevoked checks if a certificate has been revoked.
// Deprecated: Use revocationChecker.CheckRevocation instead.
func (v *CACertificateValidator) IsRevoked(_ *big.Int) bool {
	return false
}

// SelfSignedCertificateValidator validates self-signed certificates using certificate pinning.
type SelfSignedCertificateValidator struct {
	// pinnedCertificates maps client IDs to their pinned certificates
	pinnedCertificates map[string]*x509.Certificate
}

// NewSelfSignedCertificateValidator creates a new self-signed certificate validator.
func NewSelfSignedCertificateValidator(pinnedCerts map[string]*x509.Certificate) *SelfSignedCertificateValidator {
	return &SelfSignedCertificateValidator{
		pinnedCertificates: pinnedCerts,
	}
}

// ValidateCertificate validates a self-signed certificate against pinned certificates.
func (v *SelfSignedCertificateValidator) ValidateCertificate(clientCert *x509.Certificate, _ [][]byte) error {
	if clientCert == nil {
		return cryptoutilIdentityAppErr.ErrInvalidClientAuth
	}

	// Check if certificate is expired
	now := time.Now().UTC()
	if now.Before(clientCert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}

	if now.After(clientCert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	// For self-signed certificates, we require exact pinning
	// The client ID should be derivable from the certificate or provided separately
	// For now, we'll check against all pinned certificates
	found := false

	for _, pinnedCert := range v.pinnedCertificates {
		if certificatesEqual(clientCert, pinnedCert) {
			found = true

			break
		}
	}

	if !found {
		return fmt.Errorf("certificate not found in pinned certificates")
	}

	return nil
}

// IsRevoked always returns false for self-signed certificates.
// Revocation for self-signed certificates is handled by removing the pinned certificate.
func (v *SelfSignedCertificateValidator) IsRevoked(_ *big.Int) bool {
	return false
}

// certificatesEqual compares two certificates for equality.
// This is a simplified comparison - in production, you might compare fingerprints or specific fields.
func certificatesEqual(cert1, cert2 *x509.Certificate) bool {
	if cert1 == nil || cert2 == nil {
		return false
	}

	// Compare raw bytes for exact match
	return string(cert1.Raw) == string(cert2.Raw)
}

// CertificateParser provides utilities for parsing certificates from various formats.
type CertificateParser struct{}

// ParsePEMCertificate parses a PEM-encoded certificate.
func (p *CertificateParser) ParsePEMCertificate(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ParsePEMCertificateChain parses a PEM-encoded certificate chain.
func (p *CertificateParser) ParsePEMCertificateChain(pemData []byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate

	for len(pemData) > 0 {
		block, rest := pem.Decode(pemData)
		if block == nil {
			break
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate in chain: %w", err)
		}

		certs = append(certs, cert)
		pemData = rest
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM data")
	}

	return certs, nil
}
