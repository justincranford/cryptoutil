// Copyright (c) 2025 Justin Cranford

// Package revocation provides certificate revocation services including CRL and OCSP.
package revocation

import (
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"

	"golang.org/x/crypto/ocsp"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAMagic "cryptoutil/internal/apps/pki/ca/magic"
)

// RevocationReason represents X.509 revocation reasons.
type RevocationReason int

// X.509 revocation reasons per RFC 5280.
const (
	ReasonUnspecified          RevocationReason = 0
	ReasonKeyCompromise        RevocationReason = 1
	ReasonCACompromise         RevocationReason = 2
	ReasonAffiliationChanged   RevocationReason = 3
	ReasonSuperseded           RevocationReason = 4
	ReasonCessationOfOperation RevocationReason = 5
	ReasonCertificateHold      RevocationReason = 6
	ReasonRemoveFromCRL        RevocationReason = 8
	ReasonPrivilegeWithdrawn   RevocationReason = 9
	ReasonAACompromise         RevocationReason = 10
)

// String returns the string representation of a revocation reason.
func (r RevocationReason) String() string {
	reasonStrings := map[RevocationReason]string{
		ReasonUnspecified:          "unspecified",
		ReasonKeyCompromise:        "keyCompromise",
		ReasonCACompromise:         "caCompromise",
		ReasonAffiliationChanged:   "affiliationChanged",
		ReasonSuperseded:           "superseded",
		ReasonCessationOfOperation: "cessationOfOperation",
		ReasonCertificateHold:      "certificateHold",
		ReasonRemoveFromCRL:        "removeFromCRL",
		ReasonPrivilegeWithdrawn:   "privilegeWithdrawn",
		ReasonAACompromise:         "aaCompromise",
	}

	if str, ok := reasonStrings[r]; ok {
		return str
	}

	return "unknown"
}

// RevokedCertificate represents a revoked certificate entry.
type RevokedCertificate struct {
	SerialNumber   *big.Int
	RevocationTime time.Time
	Reason         RevocationReason
}

// CRLConfig configures CRL generation.
type CRLConfig struct {
	// Issuer is the CA certificate that signs the CRL.
	Issuer *x509.Certificate

	// PrivateKey is the CA's signing key.
	PrivateKey crypto.Signer

	// Provider handles cryptographic operations.
	Provider cryptoutilCACrypto.Provider

	// Validity is how long the CRL is valid.
	Validity time.Duration

	// NextUpdateBuffer is subtracted from expiry for NextUpdate.
	NextUpdateBuffer time.Duration

	// CRLNumber tracks the CRL sequence.
	CRLNumber *big.Int
}

// CRLService generates Certificate Revocation Lists.
type CRLService struct {
	config           *CRLConfig
	revokedCerts     []RevokedCertificate
	currentCRLNumber *big.Int
	mu               sync.RWMutex
}

// NewCRLService creates a new CRL service.
func NewCRLService(config *CRLConfig) (*CRLService, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.Issuer == nil {
		return nil, fmt.Errorf("issuer certificate is required")
	}

	if config.PrivateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}

	if config.Provider == nil {
		return nil, fmt.Errorf("crypto provider is required")
	}

	crlNumber := config.CRLNumber
	if crlNumber == nil {
		crlNumber = big.NewInt(1)
	}

	return &CRLService{
		config:           config,
		revokedCerts:     make([]RevokedCertificate, 0),
		currentCRLNumber: crlNumber,
	}, nil
}

// Revoke adds a certificate to the revocation list.
func (s *CRLService) Revoke(serialNumber *big.Int, reason RevocationReason) error {
	if serialNumber == nil {
		return fmt.Errorf("serial number is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already revoked.
	for _, cert := range s.revokedCerts {
		if cert.SerialNumber.Cmp(serialNumber) == 0 {
			return fmt.Errorf("certificate already revoked")
		}
	}

	s.revokedCerts = append(s.revokedCerts, RevokedCertificate{
		SerialNumber:   serialNumber,
		RevocationTime: time.Now().UTC(),
		Reason:         reason,
	})

	return nil
}

// IsRevoked checks if a certificate is revoked.
func (s *CRLService) IsRevoked(serialNumber *big.Int) (bool, *RevokedCertificate) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, cert := range s.revokedCerts {
		if cert.SerialNumber.Cmp(serialNumber) == 0 {
			return true, &cert
		}
	}

	return false, nil
}

// GenerateCRL creates a new CRL.
func (s *CRLService) GenerateCRL() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	nextUpdate := now.Add(s.config.Validity)

	if s.config.NextUpdateBuffer > 0 {
		nextUpdate = nextUpdate.Add(-s.config.NextUpdateBuffer)
	}

	// Build revoked certificate list.
	revokedList := make([]pkix.RevokedCertificate, len(s.revokedCerts))
	for i, cert := range s.revokedCerts {
		revokedList[i] = pkix.RevokedCertificate{
			SerialNumber:   cert.SerialNumber,
			RevocationTime: cert.RevocationTime,
		}
	}

	// Get signature algorithm.
	sigAlg, err := s.config.Provider.GetSignatureAlgorithm(s.config.Issuer.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature algorithm: %w", err)
	}

	// Build CRL template.
	template := &x509.RevocationList{
		Number:              s.currentCRLNumber,
		ThisUpdate:          now,
		NextUpdate:          nextUpdate,
		RevokedCertificates: revokedList,
		SignatureAlgorithm:  sigAlg,
	}

	// Sign the CRL.
	crlDER, err := x509.CreateRevocationList(crand.Reader, template, s.config.Issuer, s.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL: %w", err)
	}

	// Increment CRL number for next generation.
	s.currentCRLNumber = new(big.Int).Add(s.currentCRLNumber, big.NewInt(1))

	return crlDER, nil
}

// GenerateCRLPEM creates a PEM-encoded CRL.
func (s *CRLService) GenerateCRLPEM() ([]byte, error) {
	crlDER, err := s.GenerateCRL()
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "X509 CRL",
		Bytes: crlDER,
	}), nil
}

// GetRevokedCertificates returns a copy of all revoked certificates.
func (s *CRLService) GetRevokedCertificates() []RevokedCertificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]RevokedCertificate, len(s.revokedCerts))
	copy(result, s.revokedCerts)

	return result
}

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
		return nil, fmt.Errorf("config is required")
	}

	if config.Issuer == nil {
		return nil, fmt.Errorf("issuer certificate is required")
	}

	if config.Responder == nil {
		return nil, fmt.Errorf("responder certificate is required")
	}

	if config.ResponderKey == nil {
		return nil, fmt.Errorf("responder private key is required")
	}

	if config.Provider == nil {
		return nil, fmt.Errorf("crypto provider is required")
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

	revoked, revokedCert := s.crlService.IsRevoked(cert.SerialNumber)
	if revoked {
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

// RevocationEntry represents a revocation record for storage.
type RevocationEntry struct {
	SerialNumber   string    `json:"serial_number"`
	RevocationTime time.Time `json:"revocation_time"`
	Reason         string    `json:"reason"`
	ReasonCode     int       `json:"reason_code"`
	IssuerDN       string    `json:"issuer_dn"`
}

// ToEntry converts a RevokedCertificate to a RevocationEntry.
func (rc *RevokedCertificate) ToEntry(issuerDN string) *RevocationEntry {
	return &RevocationEntry{
		SerialNumber:   rc.SerialNumber.Text(cryptoutilCAMagic.HexBase),
		RevocationTime: rc.RevocationTime,
		Reason:         rc.Reason.String(),
		ReasonCode:     int(rc.Reason),
		IssuerDN:       issuerDN,
	}
}
