// Copyright (c) 2025 Justin Cranford

// Package intermediate provides Intermediate CA provisioning workflows.
// It implements creation and signing of intermediate CAs by a root or superior CA.
package intermediate

import (
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	crand "crypto/rand"
	rsa "crypto/rsa"
	sha256 "crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAMagic "cryptoutil/internal/ca/magic"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"
)

// IntermediateCAConfig defines configuration for intermediate CA provisioning.
type IntermediateCAConfig struct {
	// Name is the unique identifier for this intermediate CA.
	Name string

	// SubjectProfile defines subject DN constraints.
	SubjectProfile *cryptoutilCAProfileSubject.Profile

	// KeySpec defines the key algorithm and parameters.
	KeySpec cryptoutilCACrypto.KeySpec

	// OutputDir is the directory for storing CA materials.
	OutputDir string

	// ValidityDuration is the intermediate CA certificate lifetime.
	ValidityDuration time.Duration

	// PathLenConstraint limits subordinate CA depth.
	PathLenConstraint int

	// IssuerCertificate is the signing CA's certificate.
	IssuerCertificate *x509.Certificate

	// IssuerPrivateKey is the signing CA's private key.
	IssuerPrivateKey crypto.PrivateKey
}

// IntermediateCA represents a provisioned intermediate CA.
type IntermediateCA struct {
	// Name is the CA identifier.
	Name string

	// Certificate is the intermediate CA certificate.
	Certificate *x509.Certificate

	// PrivateKey is the CA signing key.
	PrivateKey crypto.PrivateKey

	// PublicKey is the CA public key.
	PublicKey crypto.PublicKey

	// CertificatePEM is the PEM-encoded certificate.
	CertificatePEM []byte

	// CertificateChainPEM is the full chain including issuer certificates.
	CertificateChainPEM []byte

	// ProvisionTime records when the CA was created.
	ProvisionTime time.Time
}

// AuditEntry records an intermediate CA provisioning operation.
type AuditEntry struct {
	// Timestamp is when the operation occurred.
	Timestamp time.Time `json:"timestamp"`

	// Operation describes what was done.
	Operation string `json:"operation"`

	// CAName identifies the CA.
	CAName string `json:"ca_name"`

	// IssuerName identifies the signing CA.
	IssuerName string `json:"issuer_name"`

	// SerialNumber is the certificate serial.
	SerialNumber string `json:"serial_number"`

	// SubjectDN is the certificate subject.
	SubjectDN string `json:"subject_dn"`

	// NotBefore is the certificate validity start.
	NotBefore time.Time `json:"not_before"`

	// NotAfter is the certificate validity end.
	NotAfter time.Time `json:"not_after"`

	// KeyAlgorithm identifies the key type.
	KeyAlgorithm string `json:"key_algorithm"`

	// Fingerprint is the certificate SHA-256 fingerprint.
	Fingerprint string `json:"fingerprint"`
}

// Provisioner handles intermediate CA creation.
type Provisioner struct {
	provider cryptoutilCACrypto.Provider
}

// NewProvisioner creates a new intermediate CA provisioner.
func NewProvisioner(provider cryptoutilCACrypto.Provider) *Provisioner {
	return &Provisioner{
		provider: provider,
	}
}

// Provision creates a new intermediate CA signed by the issuer.
func (p *Provisioner) Provision(config *IntermediateCAConfig) (*IntermediateCA, *AuditEntry, error) {
	if err := validateConfig(config); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	// Generate key pair.
	keyPair, err := p.provider.GenerateKeyPair(config.KeySpec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Prepare subject DN.
	subjectDN, err := resolveSubjectDN(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve subject DN: %w", err)
	}

	// Generate serial number.
	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	// Calculate validity period.
	now := time.Now().UTC()
	notBefore := now.Add(-cryptoutilCAMagic.BackdateBuffer)
	notAfter := now.Add(config.ValidityDuration)

	// Ensure intermediate doesn't outlive issuer.
	if notAfter.After(config.IssuerCertificate.NotAfter) {
		notAfter = config.IssuerCertificate.NotAfter
	}

	// Build certificate template.
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectDN,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            config.PathLenConstraint,
		MaxPathLenZero:        config.PathLenConstraint == 0,
	}

	// Determine signature algorithm based on issuer key.
	sigAlg, err := p.provider.GetSignatureAlgorithm(config.IssuerCertificate.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select signature algorithm: %w", err)
	}

	template.SignatureAlgorithm = sigAlg

	// Sign the certificate with issuer's key.
	certDER, err := x509.CreateCertificate(crand.Reader, template, config.IssuerCertificate, keyPair.PublicKey, config.IssuerPrivateKey)
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

	// Build chain PEM (intermediate + issuer).
	issuerPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: config.IssuerCertificate.Raw,
	})
	chainPEM := append(certPEM, issuerPEM...)

	intermediateCA := &IntermediateCA{
		Name:                config.Name,
		Certificate:         cert,
		PrivateKey:          keyPair.PrivateKey,
		PublicKey:           keyPair.PublicKey,
		CertificatePEM:      certPEM,
		CertificateChainPEM: chainPEM,
		ProvisionTime:       now,
	}

	// Create audit entry.
	audit := &AuditEntry{
		Timestamp:    now,
		Operation:    "intermediate_ca_provision",
		CAName:       config.Name,
		IssuerName:   config.IssuerCertificate.Subject.CommonName,
		SerialNumber: cert.SerialNumber.Text(cryptoutilCAMagic.HexBase),
		SubjectDN:    cert.Subject.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyAlgorithm: keyAlgorithmName(keyPair.PublicKey),
		Fingerprint:  certificateFingerprint(cert),
	}

	// Persist materials if output directory specified.
	if config.OutputDir != "" {
		if err := p.persistMaterials(config, intermediateCA); err != nil {
			return nil, nil, fmt.Errorf("failed to persist materials: %w", err)
		}
	}

	return intermediateCA, audit, nil
}

func validateConfig(config *IntermediateCAConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.Name == "" {
		return fmt.Errorf("CA name is required")
	}

	if config.ValidityDuration <= 0 {
		return fmt.Errorf("validity duration must be positive")
	}

	if config.PathLenConstraint < 0 {
		return fmt.Errorf("path length constraint cannot be negative")
	}

	if config.IssuerCertificate == nil {
		return fmt.Errorf("issuer certificate is required")
	}

	if config.IssuerPrivateKey == nil {
		return fmt.Errorf("issuer private key is required")
	}

	// Verify issuer is a CA.
	if !config.IssuerCertificate.IsCA {
		return fmt.Errorf("issuer certificate is not a CA")
	}

	// Check path length constraints.
	if config.IssuerCertificate.MaxPathLenZero {
		return fmt.Errorf("issuer CA has path length 0, cannot sign subordinate CAs")
	}

	if config.IssuerCertificate.MaxPathLen > 0 && config.PathLenConstraint >= config.IssuerCertificate.MaxPathLen {
		return fmt.Errorf("intermediate path length (%d) must be less than issuer path length (%d)",
			config.PathLenConstraint, config.IssuerCertificate.MaxPathLen)
	}

	return nil
}

func resolveSubjectDN(config *IntermediateCAConfig) (pkix.Name, error) {
	if config.SubjectProfile == nil {
		// Use default subject with CA name as CN.
		return pkix.Name{
			CommonName: config.Name,
		}, nil
	}

	// Use profile defaults.
	return pkix.Name{
		CommonName:         config.Name,
		Organization:       config.SubjectProfile.Subject.Organization,
		OrganizationalUnit: config.SubjectProfile.Subject.OrganizationalUnit,
		Country:            config.SubjectProfile.Subject.Country,
		Province:           config.SubjectProfile.Subject.State,
		Locality:           config.SubjectProfile.Subject.Locality,
	}, nil
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

func (p *Provisioner) persistMaterials(config *IntermediateCAConfig, intermediateCA *IntermediateCA) error {
	// Create output directory.
	if err := os.MkdirAll(config.OutputDir, cryptoutilCAMagic.DirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write certificate.
	certPath := filepath.Join(config.OutputDir, config.Name+".crt")
	if err := os.WriteFile(certPath, intermediateCA.CertificatePEM, cryptoutilCAMagic.FilePermissions); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write certificate chain.
	chainPath := filepath.Join(config.OutputDir, config.Name+"-chain.crt")
	if err := os.WriteFile(chainPath, intermediateCA.CertificateChainPEM, cryptoutilCAMagic.FilePermissions); err != nil {
		return fmt.Errorf("failed to write certificate chain: %w", err)
	}

	// Write private key (with restrictive permissions).
	keyDER, err := x509.MarshalPKCS8PrivateKey(intermediateCA.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	keyPath := filepath.Join(config.OutputDir, config.Name+".key")
	if err := os.WriteFile(keyPath, keyPEM, cryptoutilCAMagic.KeyFilePermissions); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
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
