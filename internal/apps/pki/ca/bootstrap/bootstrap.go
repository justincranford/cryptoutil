// Copyright (c) 2025 Justin Cranford

// Package bootstrap provides Root CA bootstrap workflows for offline root CA creation.
// It implements deterministic serial numbers, key storage, and audit logging.
package bootstrap

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

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// RootCAConfig defines configuration for root CA bootstrap.
type RootCAConfig struct {
	// Name is the unique identifier for this root CA.
	Name string

	// SubjectProfile defines subject DN constraints.
	SubjectProfile *cryptoutilCAProfileSubject.Profile

	// CertProfile defines certificate policy.
	CertProfile *cryptoutilCAProfileCertificate.Profile

	// KeySpec defines the key algorithm and parameters.
	KeySpec cryptoutilCACrypto.KeySpec

	// OutputDir is the directory for storing CA materials.
	OutputDir string

	// ValidityDuration is the root CA certificate lifetime.
	ValidityDuration time.Duration

	// PathLenConstraint limits subordinate CA depth.
	PathLenConstraint int
}

// RootCA represents a bootstrapped root CA.
type RootCA struct {
	// Name is the CA identifier.
	Name string

	// Certificate is the root CA certificate.
	Certificate *x509.Certificate

	// PrivateKey is the CA signing key (only available during bootstrap).
	PrivateKey crypto.PrivateKey

	// PublicKey is the CA public key.
	PublicKey crypto.PublicKey

	// CertificatePEM is the PEM-encoded certificate.
	CertificatePEM []byte

	// BootstrapTime records when the CA was created.
	BootstrapTime time.Time
}

// AuditEntry records a bootstrap operation.
type AuditEntry struct {
	// Timestamp is when the operation occurred.
	Timestamp time.Time `json:"timestamp"`

	// Operation describes what was done.
	Operation string `json:"operation"`

	// CAName identifies the CA.
	CAName string `json:"ca_name"`

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

// Bootstrapper handles root CA creation.
type Bootstrapper struct {
	provider cryptoutilCACrypto.Provider
}

// NewBootstrapper creates a new root CA bootstrapper.
func NewBootstrapper(provider cryptoutilCACrypto.Provider) *Bootstrapper {
	return &Bootstrapper{
		provider: provider,
	}
}

// Bootstrap creates a new root CA with the given configuration.
func (b *Bootstrapper) Bootstrap(config *RootCAConfig) (*RootCA, *AuditEntry, error) {
	if err := validateConfig(config); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	// Generate key pair.
	keyPair, err := b.provider.GenerateKeyPair(config.KeySpec)
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
	notBefore := now.Add(-cryptoutilSharedMagic.BackdateBuffer)
	notAfter := now.Add(config.ValidityDuration)

	// Build certificate template.
	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subjectDN,
		Issuer:                subjectDN, // Self-signed.
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            config.PathLenConstraint,
		MaxPathLenZero:        config.PathLenConstraint == 0,
	}

	// Determine signature algorithm.
	sigAlg, err := b.provider.GetSignatureAlgorithm(keyPair.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to select signature algorithm: %w", err)
	}

	template.SignatureAlgorithm = sigAlg

	// Self-sign the certificate.
	certDER, err := x509.CreateCertificate(crand.Reader, template, template, keyPair.PublicKey, keyPair.PrivateKey)
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

	rootCA := &RootCA{
		Name:           config.Name,
		Certificate:    cert,
		PrivateKey:     keyPair.PrivateKey,
		PublicKey:      keyPair.PublicKey,
		CertificatePEM: certPEM,
		BootstrapTime:  now,
	}

	// Create audit entry.
	audit := &AuditEntry{
		Timestamp:    now,
		Operation:    "root_ca_bootstrap",
		CAName:       config.Name,
		SerialNumber: cert.SerialNumber.Text(cryptoutilSharedMagic.HexBase),
		SubjectDN:    cert.Subject.String(),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		KeyAlgorithm: keyAlgorithmName(keyPair.PublicKey),
		Fingerprint:  certificateFingerprint(cert),
	}

	// Persist materials if output directory specified.
	if config.OutputDir != "" {
		if err := b.persistMaterials(config, rootCA); err != nil {
			return nil, nil, fmt.Errorf("failed to persist materials: %w", err)
		}
	}

	return rootCA, audit, nil
}

func validateConfig(config *RootCAConfig) error {
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

	return nil
}

func resolveSubjectDN(config *RootCAConfig) (pkix.Name, error) {
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
	serialBytes := make([]byte, cryptoutilSharedMagic.SerialNumberLength)
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

func (b *Bootstrapper) persistMaterials(config *RootCAConfig, rootCA *RootCA) error {
	// Create output directory.
	if err := os.MkdirAll(config.OutputDir, cryptoutilSharedMagic.DirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write certificate.
	certPath := filepath.Join(config.OutputDir, config.Name+".crt")
	if err := os.WriteFile(certPath, rootCA.CertificatePEM, cryptoutilSharedMagic.FilePermissions); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key (with restrictive permissions).
	keyDER, err := x509.MarshalPKCS8PrivateKey(rootCA.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	keyPath := filepath.Join(config.OutputDir, config.Name+".key")
	if err := os.WriteFile(keyPath, keyPEM, cryptoutilSharedMagic.KeyFilePermissions); err != nil {
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
