// Copyright (c) 2025 Justin Cranford

// Package issuer provides end-entity certificate issuance services.
// It implements certificate request processing, validation, and signing.
package issuer

import (
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	crand "crypto/rand"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAMagic "cryptoutil/internal/apps/pki/ca/magic"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
)

// IssuingCAConfig defines an issuing CA's configuration.
type IssuingCAConfig struct {
	// Name is the unique identifier for this issuing CA.
	Name string

	// Certificate is the issuing CA's certificate.
	Certificate *x509.Certificate

	// PrivateKey is the issuing CA's signing key.
	PrivateKey crypto.PrivateKey

	// SubjectProfile defines subject constraints for issued certificates.
	SubjectProfile *cryptoutilCAProfileSubject.Profile

	// CertProfile defines certificate policy for issued certificates.
	CertProfile *cryptoutilCAProfileCertificate.Profile
}

// CertificateRequest represents a request for certificate issuance.
type CertificateRequest struct {
	// SubjectRequest contains the requested subject DN and SANs.
	SubjectRequest *cryptoutilCAProfileSubject.Request

	// PublicKey is the requestor's public key.
	PublicKey crypto.PublicKey

	// ValidityDuration is the requested certificate lifetime.
	ValidityDuration time.Duration

	// Extensions contains any additional requested extensions.
	Extensions []Extension
}

// Extension represents a custom certificate extension.
type Extension struct {
	OID      []int
	Critical bool
	Value    []byte
}

// IssuedCertificate represents a newly issued certificate.
type IssuedCertificate struct {
	// Certificate is the issued certificate.
	Certificate *x509.Certificate

	// CertificatePEM is the PEM-encoded certificate.
	CertificatePEM []byte

	// ChainPEM is the full certificate chain.
	ChainPEM []byte

	// SerialNumber is the certificate serial in hex.
	SerialNumber string

	// Fingerprint is the SHA-256 fingerprint.
	Fingerprint string

	// IssuedAt is when the certificate was issued.
	IssuedAt time.Time
}

// AuditEntry records certificate issuance.
type AuditEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	Operation      string    `json:"operation"`
	IssuerName     string    `json:"issuer_name"`
	SerialNumber   string    `json:"serial_number"`
	SubjectDN      string    `json:"subject_dn"`
	SANs           []string  `json:"sans"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	KeyAlgorithm   string    `json:"key_algorithm"`
	Fingerprint    string    `json:"fingerprint"`
	ProfileName    string    `json:"profile_name"`
	SubjectProfile string    `json:"subject_profile"`
}

// Issuer handles end-entity certificate issuance.
type Issuer struct {
	provider cryptoutilCACrypto.Provider
	caConfig *IssuingCAConfig
}

// NewIssuer creates a new certificate issuer.
func NewIssuer(provider cryptoutilCACrypto.Provider, caConfig *IssuingCAConfig) (*Issuer, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider cannot be nil")
	}

	if err := validateCAConfig(caConfig); err != nil {
		return nil, fmt.Errorf("invalid CA config: %w", err)
	}

	return &Issuer{
		provider: provider,
		caConfig: caConfig,
	}, nil
}

// GetCAConfig returns the issuer's CA configuration.
func (i *Issuer) GetCAConfig() *IssuingCAConfig {
	return i.caConfig
}

// Issue creates a new certificate based on the request.
func (i *Issuer) Issue(req *CertificateRequest) (*IssuedCertificate, *AuditEntry, error) {
	if err := i.validateRequest(req); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
	}

	// Resolve subject DN and SANs using profile.
	resolved, err := i.resolveSubject(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve subject: %w", err)
	}

	// Generate serial number.
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Calculate validity period.
	now := time.Now().UTC()
	notBefore := now.Add(-cryptoutilCAMagic.BackdateBuffer)
	notAfter := now.Add(req.ValidityDuration)

	// Ensure certificate doesn't outlive issuer.
	if notAfter.After(i.caConfig.Certificate.NotAfter) {
		notAfter = i.caConfig.Certificate.NotAfter
	}

	// Build key usage from profile.
	keyUsage := i.getKeyUsage()
	extKeyUsage := i.getExtKeyUsage()

	// Build certificate template.
	template := &x509.Certificate{
		SerialNumber:   serialNumber,
		Subject:        resolved.DN,
		NotBefore:      notBefore,
		NotAfter:       notAfter,
		KeyUsage:       keyUsage,
		ExtKeyUsage:    extKeyUsage,
		DNSNames:       resolved.DNSNames,
		IPAddresses:    resolved.IPAddresses,
		EmailAddresses: resolved.EmailAddresses,
		URIs:           resolved.URIs,
	}

	// Determine signature algorithm based on issuer key.
	sigAlg, err := i.provider.GetSignatureAlgorithm(i.caConfig.Certificate.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select signature algorithm: %w", err)
	}

	template.SignatureAlgorithm = sigAlg

	// Sign the certificate.
	certDER, err := x509.CreateCertificate(crand.Reader, template, i.caConfig.Certificate, req.PublicKey, i.caConfig.PrivateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	// Parse the created certificate.
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Encode to PEM.
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Build chain PEM.
	issuerPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: i.caConfig.Certificate.Raw,
	})
	chainPEM := append(certPEM, issuerPEM...)

	issued := &IssuedCertificate{
		Certificate:    cert,
		CertificatePEM: certPEM,
		ChainPEM:       chainPEM,
		SerialNumber:   cert.SerialNumber.Text(cryptoutilCAMagic.HexBase),
		Fingerprint:    certificateFingerprint(cert),
		IssuedAt:       now,
	}

	// Build audit entry.
	audit := &AuditEntry{
		Timestamp:    now,
		Operation:    "certificate_issuance",
		IssuerName:   i.caConfig.Certificate.Subject.CommonName,
		SerialNumber: issued.SerialNumber,
		SubjectDN:    cert.Subject.String(),
		SANs:         collectSANs(cert),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyAlgorithm: keyAlgorithmName(req.PublicKey),
		Fingerprint:  issued.Fingerprint,
	}

	if i.caConfig.CertProfile != nil {
		audit.ProfileName = i.caConfig.CertProfile.Name
	}

	if i.caConfig.SubjectProfile != nil {
		audit.SubjectProfile = i.caConfig.SubjectProfile.Name
	}

	return issued, audit, nil
}

func validateCAConfig(config *IssuingCAConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Name == "" {
		return fmt.Errorf("CA name is required")
	}

	if config.Certificate == nil {
		return fmt.Errorf("CA certificate is required")
	}

	if config.PrivateKey == nil {
		return fmt.Errorf("CA private key is required")
	}

	if !config.Certificate.IsCA {
		return fmt.Errorf("certificate is not a CA")
	}

	return nil
}

func (i *Issuer) validateRequest(req *CertificateRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.PublicKey == nil {
		return fmt.Errorf("public key is required")
	}

	if req.ValidityDuration <= 0 {
		return fmt.Errorf("validity duration must be positive")
	}

	// Validate against certificate profile if present.
	if i.caConfig.CertProfile != nil {
		if err := i.caConfig.CertProfile.Validity.ValidateDuration(req.ValidityDuration); err != nil {
			return fmt.Errorf("validity validation failed: %w", err)
		}
	}

	return nil
}

func (i *Issuer) resolveSubject(req *CertificateRequest) (*cryptoutilCAProfileSubject.ResolvedSubject, error) {
	if i.caConfig.SubjectProfile == nil {
		// No profile - build from request directly.
		return i.buildDirectSubject(req)
	}

	// Use profile to resolve and validate subject.
	resolved, err := i.caConfig.SubjectProfile.Resolve(req.SubjectRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve subject from profile: %w", err)
	}

	return resolved, nil
}

func (i *Issuer) buildDirectSubject(req *CertificateRequest) (*cryptoutilCAProfileSubject.ResolvedSubject, error) {
	if req.SubjectRequest == nil {
		return nil, fmt.Errorf("subject request is required when no subject profile is configured")
	}

	resolved := &cryptoutilCAProfileSubject.ResolvedSubject{}

	resolved.DN.CommonName = req.SubjectRequest.CommonName
	resolved.DN.Organization = req.SubjectRequest.Organization
	resolved.DN.OrganizationalUnit = req.SubjectRequest.OrganizationalUnit
	resolved.DN.Country = req.SubjectRequest.Country
	resolved.DN.Province = req.SubjectRequest.State
	resolved.DN.Locality = req.SubjectRequest.Locality

	resolved.DNSNames = req.SubjectRequest.DNSNames
	resolved.EmailAddresses = req.SubjectRequest.EmailAddresses

	// Parse IP addresses.
	for _, ipStr := range req.SubjectRequest.IPAddresses {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}

		resolved.IPAddresses = append(resolved.IPAddresses, ip)
	}

	// Parse URIs.
	for _, uriStr := range req.SubjectRequest.URIs {
		uri, err := url.Parse(uriStr)
		if err != nil {
			return nil, fmt.Errorf("invalid URI %s: %w", uriStr, err)
		}

		resolved.URIs = append(resolved.URIs, uri)
	}

	return resolved, nil
}

func (i *Issuer) getKeyUsage() x509.KeyUsage {
	if i.caConfig.CertProfile != nil {
		return i.caConfig.CertProfile.KeyUsage.ToX509KeyUsage()
	}

	// Default to TLS server key usage.
	return x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
}

func (i *Issuer) getExtKeyUsage() []x509.ExtKeyUsage {
	if i.caConfig.CertProfile != nil {
		return i.caConfig.CertProfile.ExtendedKeyUsage.ToX509ExtKeyUsage()
	}

	// Default to TLS server auth.
	return []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
}

func generateSerialNumber() (*big.Int, error) {
	// Generate 20 bytes (160 bits) of randomness per CA/Browser Forum requirements.
	serialBytes := make([]byte, cryptoutilCAMagic.SerialNumberLength)
	if _, err := crand.Read(serialBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Ensure positive by clearing the high bit.
	serialBytes[0] &= 0x7F

	// Ensure non-zero.
	if serialBytes[0] == 0 {
		serialBytes[0] = 0x01
	}

	serial := new(big.Int).SetBytes(serialBytes)

	return serial, nil
}

func keyAlgorithmName(pub crypto.PublicKey) string {
	switch pub.(type) {
	case *rsa.PublicKey:
		return "RSA"
	case *ecdsa.PublicKey:
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	default:
		return "Unknown"
	}
}

func certificateFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)

	return fmt.Sprintf("%x", hash)
}

func collectSANs(cert *x509.Certificate) []string {
	capacity := len(cert.DNSNames) + len(cert.IPAddresses) + len(cert.EmailAddresses) + len(cert.URIs)
	sans := make([]string, 0, capacity)

	for _, dns := range cert.DNSNames {
		sans = append(sans, "DNS:"+dns)
	}

	for _, ip := range cert.IPAddresses {
		sans = append(sans, "IP:"+ip.String())
	}

	for _, email := range cert.EmailAddresses {
		sans = append(sans, "email:"+email)
	}

	for _, uri := range cert.URIs {
		sans = append(sans, "URI:"+uri.String())
	}

	return sans
}
