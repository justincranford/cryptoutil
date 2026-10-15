// Copyright (c) 2025-2026 Justin Cranford.
// Package revocation provides certificate revocation services including CRL and OCSP.
package revocation

import (
	"context"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"sync"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RevocationReason represents X.509 revocation reasons.
type RevocationReason int

// X.509 revocation reasons per RFC 5280.
const (
	ReasonUnspecified          RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonUnspecified
	ReasonKeyCompromise        RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonKeyCompromise
	ReasonCACompromise         RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCACompromise
	ReasonAffiliationChanged   RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonAffiliationChanged
	ReasonSuperseded           RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonSuperseded
	ReasonCessationOfOperation RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCessationOfOperation
	ReasonCertificateHold      RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonCertificateHold
	ReasonRemoveFromCRL        RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonRemoveFromCRL
	ReasonPrivilegeWithdrawn   RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonPrivilegeWithdrawn
	ReasonAACompromise         RevocationReason = cryptoutilSharedMagic.RFC5280RevocationReasonAACompromise
)

// String returns the string representation of a revocation reason.
func (r RevocationReason) String() string {
	reasonStrings := map[RevocationReason]string{
		ReasonUnspecified:          cryptoutilSharedMagic.RFC5280RevocationReasonNameUnspecified,
		ReasonKeyCompromise:        cryptoutilSharedMagic.RFC5280RevocationReasonNameKeyCompromise,
		ReasonCACompromise:         cryptoutilSharedMagic.RFC5280RevocationReasonNameCACompromise,
		ReasonAffiliationChanged:   cryptoutilSharedMagic.RFC5280RevocationReasonNameAffiliationChanged,
		ReasonSuperseded:           cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded,
		ReasonCessationOfOperation: cryptoutilSharedMagic.RFC5280RevocationReasonNameCessationOfOperation,
		ReasonCertificateHold:      cryptoutilSharedMagic.RFC5280RevocationReasonNameCertificateHold,
		ReasonRemoveFromCRL:        cryptoutilSharedMagic.RFC5280RevocationReasonNameRemoveFromCRL,
		ReasonPrivilegeWithdrawn:   cryptoutilSharedMagic.RFC5280RevocationReasonNamePrivilegeWithdrawn,
		ReasonAACompromise:         cryptoutilSharedMagic.RFC5280RevocationReasonNameAACompromise,
	}

	if str, ok := reasonStrings[r]; ok {
		return str
	}

	return cryptoutilSharedMagic.UNKNOWN
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

	// AllocateCRLNumber durably allocates the next number for the issuer key ID.
	// Production deployments use this to coordinate CRL generation across instances.
	AllocateCRLNumber func(context.Context, string) (*big.Int, error)
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
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorConfigRequired)
	}

	if config.Issuer == nil {
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorIssuerCertificateRequired)
	}

	if config.PrivateKey == nil {
		return nil, fmt.Errorf("private key is required")
	}

	if config.Provider == nil {
		return nil, fmt.Errorf(cryptoutilSharedMagic.PKIErrorCryptoProviderRequired)
	}

	crlNumber := config.CRLNumber
	if crlNumber == nil {
		crlNumber = big.NewInt(1)
	}

	return &CRLService{
		config:           config,
		revokedCerts:     make([]RevokedCertificate, 0),
		currentCRLNumber: new(big.Int).Set(crlNumber),
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
	return s.GenerateCRLContext(context.Background())
}

// GenerateCRLContext creates a CRL and allocates its number using ctx when a durable
// allocator is configured.
func (s *CRLService) GenerateCRLContext(ctx context.Context) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.generateCRL(ctx, s.revokedCerts)
}

// GenerateCRLWithRevokedCertificates creates a CRL from persisted revocation records.
func (s *CRLService) GenerateCRLWithRevokedCertificates(revokedCerts []RevokedCertificate) ([]byte, error) {
	return s.GenerateCRLWithRevokedCertificatesContext(context.Background(), revokedCerts)
}

// GenerateCRLWithRevokedCertificatesContext creates a CRL from persisted revocation
// records, allocating its number from the configured durable issuer sequence when present.
func (s *CRLService) GenerateCRLWithRevokedCertificatesContext(ctx context.Context, revokedCerts []RevokedCertificate) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.generateCRL(ctx, revokedCerts)
}

func (s *CRLService) generateCRL(ctx context.Context, revokedCerts []RevokedCertificate) ([]byte, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	crlNumber := new(big.Int).Set(s.currentCRLNumber)
	if s.config.AllocateCRLNumber != nil {
		if len(s.config.Issuer.SubjectKeyId) == 0 {
			return nil, fmt.Errorf("issuer subject key identifier is required for durable CRL numbering")
		}

		allocatedNumber, err := s.config.AllocateCRLNumber(ctx, fmt.Sprintf("%x", s.config.Issuer.SubjectKeyId))
		if err != nil {
			return nil, fmt.Errorf("failed to allocate durable CRL number: %w", err)
		}

		if allocatedNumber == nil || allocatedNumber.Sign() <= 0 {
			return nil, fmt.Errorf("durable CRL allocator returned an invalid sequence number")
		}

		crlNumber = new(big.Int).Set(allocatedNumber)
	}

	now := time.Now().UTC()
	nextUpdate := now.Add(s.config.Validity)

	if s.config.NextUpdateBuffer > 0 {
		nextUpdate = nextUpdate.Add(-s.config.NextUpdateBuffer)
	}

	// Build revoked certificate list.
	revokedList := make([]x509.RevocationListEntry, len(revokedCerts))
	for i, cert := range revokedCerts {
		revokedList[i] = x509.RevocationListEntry{
			SerialNumber:   cert.SerialNumber,
			RevocationTime: cert.RevocationTime,
			ReasonCode:     int(cert.Reason),
		}
	}

	// Get signature algorithm.
	sigAlg, err := s.config.Provider.GetSignatureAlgorithm(s.config.Issuer.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get signature algorithm: %w", err)
	}

	// Build CRL template.
	template := &x509.RevocationList{
		Number:                    crlNumber,
		ThisUpdate:                now,
		NextUpdate:                nextUpdate,
		RevokedCertificateEntries: revokedList,
		SignatureAlgorithm:        sigAlg,
	}

	// Sign the CRL.
	crlDER, err := x509.CreateRevocationList(crand.Reader, template, s.config.Issuer, s.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRL: %w", err)
	}

	if s.config.AllocateCRLNumber == nil {
		s.currentCRLNumber = new(big.Int).Add(crlNumber, big.NewInt(1))
	}

	return crlDER, nil
}

// GenerateCRLPEM creates a PEM-encoded CRL.
func (s *CRLService) GenerateCRLPEM() ([]byte, error) {
	crlDER, err := s.GenerateCRL()
	if err != nil {
		return nil, err
	}

	return encodeCRLPEM(crlDER), nil
}

// GenerateCRLPEMContext creates a PEM CRL using the context-aware durable allocator.
func (s *CRLService) GenerateCRLPEMContext(ctx context.Context) ([]byte, error) {
	crlDER, err := s.GenerateCRLContext(ctx)
	if err != nil {
		return nil, err
	}

	return encodeCRLPEM(crlDER), nil
}

// GenerateCRLPEMWithRevokedCertificates creates a PEM CRL from persisted revocation records.
func (s *CRLService) GenerateCRLPEMWithRevokedCertificates(revokedCerts []RevokedCertificate) ([]byte, error) {
	crlDER, err := s.GenerateCRLWithRevokedCertificates(revokedCerts)
	if err != nil {
		return nil, err
	}

	return encodeCRLPEM(crlDER), nil
}

// GenerateCRLPEMWithRevokedCertificatesContext creates a PEM CRL from persisted
// revocation records using a context-aware durable allocator.
func (s *CRLService) GenerateCRLPEMWithRevokedCertificatesContext(
	ctx context.Context,
	revokedCerts []RevokedCertificate,
) ([]byte, error) {
	crlDER, err := s.GenerateCRLWithRevokedCertificatesContext(ctx, revokedCerts)
	if err != nil {
		return nil, err
	}

	return encodeCRLPEM(crlDER), nil
}

func encodeCRLPEM(crlDER []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "X509 CRL",
		Bytes: crlDER,
	})
}

// GetRevokedCertificates returns a copy of all revoked certificates.
func (s *CRLService) GetRevokedCertificates() []RevokedCertificate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]RevokedCertificate, len(s.revokedCerts))
	copy(result, s.revokedCerts)

	return result
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
		SerialNumber:   rc.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		RevocationTime: rc.RevocationTime,
		Reason:         rc.Reason.String(),
		ReasonCode:     int(rc.Reason),
		IssuerDN:       issuerDN,
	}
}
