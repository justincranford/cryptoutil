// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/ocsp"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// OCSPConfig configures OCSP response generation.
type OCSPConfig struct {
	// Issuer is the CA certificate.
	Issuer *x509.Certificate

	// Responder is the OCSP responder certificate.
	Responder *x509.Certificate

	// ResponderKey is the OCSP responder's private key.
	ResponderKey crypto.Signer

	// Provider handles cryptographic operations.
	Provider cryptoutilCACrypto.Provider

	// Validity is how long the OCSP response is valid.
	Validity time.Duration
}

// OCSPService generates OCSP responses.
type OCSPService struct {
	config     *OCSPConfig
	crlService *CRLService
	mu         sync.RWMutex
}

// NewOCSPService creates a new OCSP service.
func NewOCSPService(config *OCSPConfig, crlService *CRLService) (*OCSPService, error) {
	if config == nil {
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorConfigRequired)
	}

	if config.Issuer == nil {
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorIssuerCertificateRequired)
	}

	if config.Responder == nil {
		return nil, fmt.Errorf("responder certificate is required")
	}

	if config.ResponderKey == nil {
		return nil, fmt.Errorf("responder private key is required")
	}

	if config.Provider == nil {
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorCryptoProviderRequired)
	}

	if crlService == nil {
		return nil, fmt.Errorf("CRL service is required")
	}

	return &OCSPService{
		config:     config,
		crlService: crlService,
	}, nil
}

// CreateResponse generates an OCSP response for a certificate.
func (s *OCSPService) CreateResponse(cert *x509.Certificate) ([]byte, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is required")
	}

	revoked, revokedCert := s.crlService.IsRevoked(cert.SerialNumber)
	if !revoked {
		revokedCert = nil
	}

	return s.CreateResponseWithRevocation(cert, revokedCert)
}

// CreateResponseWithRevocation generates an OCSP response using the supplied persisted status.
func (s *OCSPService) CreateResponseWithRevocation(cert *x509.Certificate, revokedCert *RevokedCertificate) ([]byte, error) {
	if cert == nil {
		return nil, fmt.Errorf("certificate is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	nextUpdate := now.Add(s.config.Validity)

	// Check revocation status.
	status := ocsp.Good

	var (
		revokedAt        time.Time
		revocationReason int
	)

	if revokedCert != nil {
		status = ocsp.Revoked
		revokedAt = revokedCert.RevocationTime
		revocationReason = int(revokedCert.Reason)
	}

	// Build OCSP response template.
	template := ocsp.Response{
		Status:           status,
		SerialNumber:     cert.SerialNumber,
		Certificate:      s.config.Responder,
		ThisUpdate:       now,
		NextUpdate:       nextUpdate,
		RevokedAt:        revokedAt,
		RevocationReason: revocationReason,
	}

	// Sign the response.
	responseBytes, err := ocsp.CreateResponse(
		s.config.Issuer,
		s.config.Responder,
		template,
		s.config.ResponderKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCSP response: %w", err)
	}

	return responseBytes, nil
}

// ParseRequest parses an OCSP request.
func (s *OCSPService) ParseRequest(requestBytes []byte) (*ocsp.Request, error) {
	if len(requestBytes) == 0 {
		return nil, fmt.Errorf("empty OCSP request")
	}

	request, err := ocsp.ParseRequest(requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OCSP request: %w", err)
	}

	return request, nil
}

// RespondToRequest processes an OCSP request and returns a response.
func (s *OCSPService) RespondToRequest(requestBytes []byte, certLookup func(serialNumber *big.Int) *x509.Certificate) ([]byte, error) {
	request, err := s.ParseRequest(requestBytes)
	if err != nil {
		return nil, err
	}

	// Look up the certificate.
	cert := certLookup(request.SerialNumber)
	if cert == nil {
		// Return unknown status for certificate not found.
		return s.createUnknownResponse(request.SerialNumber)
	}

	return s.CreateResponse(cert)
}

// RespondToRequestWithStatus processes an OCSP request using persisted certificate status.
func (s *OCSPService) RespondToRequestWithStatus(
	requestBytes []byte,
	certificateLookup func(serialNumber *big.Int) (*x509.Certificate, *RevokedCertificate, error),
) ([]byte, error) {
	if certificateLookup == nil {
		return nil, fmt.Errorf("certificate status lookup is required")
	}

	request, err := s.ParseRequest(requestBytes)
	if err != nil {
		return nil, err
	}

	if err := s.validateRequestIssuer(request); err != nil {
		return nil, err
	}

	certificate, revokedCert, err := certificateLookup(request.SerialNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to look up certificate status: %w", err)
	}

	if certificate == nil {
		return s.createUnknownResponse(request.SerialNumber)
	}

	return s.CreateResponseWithRevocation(certificate, revokedCert)
}

func (s *OCSPService) validateRequestIssuer(request *ocsp.Request) error {
	expectedRequestDER, err := ocsp.CreateRequest(
		s.config.Issuer,
		s.config.Issuer,
		&ocsp.RequestOptions{Hash: request.HashAlgorithm},
	)
	if err != nil {
		return fmt.Errorf("failed to compute OCSP issuer identity: %w", err)
	}

	expectedRequest, err := ocsp.ParseRequest(expectedRequestDER)
	if err != nil {
		return fmt.Errorf("failed to parse computed OCSP issuer identity: %w", err)
	}

	if !bytes.Equal(request.IssuerNameHash, expectedRequest.IssuerNameHash) ||
		!bytes.Equal(request.IssuerKeyHash, expectedRequest.IssuerKeyHash) {
		return fmt.Errorf("OCSP request issuer does not match the selected authority")
	}

	return nil
}

// createUnknownResponse creates an OCSP response for an unknown certificate.
func (s *OCSPService) createUnknownResponse(serialNumber *big.Int) ([]byte, error) {
	now := time.Now().UTC()

	template := ocsp.Response{
		Status:       ocsp.Unknown,
		SerialNumber: serialNumber,
		Certificate:  s.config.Responder,
		ThisUpdate:   now,
		NextUpdate:   now.Add(s.config.Validity),
	}

	responseBytes, err := ocsp.CreateResponse(
		s.config.Issuer,
		s.config.Responder,
		template,
		s.config.ResponderKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create unknown OCSP response: %w", err)
	}

	return responseBytes, nil
}
