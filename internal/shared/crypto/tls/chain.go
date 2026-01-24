// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/elliptic"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// fqdnPattern validates FQDN-style names (per Session 3 Q3).
// Allows alphanumeric, hyphens, and dots. Must start/end with alphanumeric.
var fqdnPattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$`)

// ValidateFQDN checks if the given name is a valid FQDN-style name.
// This is used for certificate CNs per Session 3 Q3.
func ValidateFQDN(name string) error {
	if name == "" {
		return fmt.Errorf("FQDN cannot be empty")
	}

	if len(name) > cryptoutilSharedMagic.FQDNMaxLength {
		return fmt.Errorf("FQDN too long: %d characters (max %d)", len(name), cryptoutilSharedMagic.FQDNMaxLength)
	}

	if !fqdnPattern.MatchString(name) {
		return fmt.Errorf("invalid FQDN format: %s", name)
	}

	// Validate each label.
	labels := strings.Split(name, ".")
	for _, label := range labels {
		if len(label) > cryptoutilSharedMagic.FQDNLabelMaxLength {
			return fmt.Errorf("FQDN label too long: %s (%d characters, max %d)", label, len(label), cryptoutilSharedMagic.FQDNLabelMaxLength)
		}
	}

	return nil
}

// CNStyle specifies the style of Common Name for certificates.
type CNStyle string

const (
	// CNStyleFQDN uses FQDN format (e.g., "kms.cryptoutil.demo.local").
	// This is the default per Session 3 Q3.
	CNStyleFQDN CNStyle = "fqdn"

	// CNStyleDescriptive uses descriptive format (e.g., "KMS Root CA").
	// Useful for human-readable CA certificates.
	CNStyleDescriptive CNStyle = "descriptive"
)

// ECCurve defines the elliptic curve to use for key generation.
type ECCurve string

const (
	// CurveP256 is NIST P-256 (secp256r1).
	CurveP256 ECCurve = "P-256"

	// CurveP384 is NIST P-384 (secp384r1).
	CurveP384 ECCurve = "P-384"

	// CurveP521 is NIST P-521 (secp521r1).
	CurveP521 ECCurve = "P-521"
)

// DefaultECCurve is P-256 (most commonly used, good balance of security and performance).
const DefaultECCurve = CurveP256

// DefaultCNStyle is FQDN per Session 3 Q3.
const DefaultCNStyle = CNStyleFQDN

// CAChainOptions holds options for creating a CA chain.
type CAChainOptions struct {
	// ChainLength is the number of CAs in the chain (including root).
	// Default: 3 (Root → Policy → Issuing per Session 3 Q2).
	ChainLength int

	// CommonNamePrefix is the prefix for CA common names.
	// For FQDN style: "cryptoutil.demo" produces "root.cryptoutil.demo", "policy.cryptoutil.demo", etc.
	// For Descriptive style: "cryptoutil.demo" produces "cryptoutil.demo Root CA", etc.
	CommonNamePrefix string

	// CNStyle specifies the Common Name style.
	// Default: FQDN (per Session 3 Q3).
	CNStyle CNStyle

	// Duration is the validity duration for all CA certificates.
	Duration time.Duration

	// Curve specifies the elliptic curve for CA key pairs.
	// Default: P-256.
	Curve ECCurve
}

// DefaultCAChainOptions returns CA chain options with sensible defaults.
func DefaultCAChainOptions(commonNamePrefix string) *CAChainOptions {
	return &CAChainOptions{
		ChainLength:      DefaultCAChainLength,
		CommonNamePrefix: commonNamePrefix,
		CNStyle:          DefaultCNStyle,
		Duration:         DefaultCADuration,
		Curve:            DefaultECCurve,
	}
}

// DefaultCAChainLength is 3: Root → Policy → Issuing (per Session 3 Q2).
const DefaultCAChainLength = 3

// DefaultCADuration is 10 years for CA certificates.
const DefaultCADuration = 10 * 365 * 24 * time.Hour

// DefaultEndEntityDuration is 1 year for end entity certificates.
const DefaultEndEntityDuration = 365 * 24 * time.Hour

// EndEntityOptions holds options for creating an end entity certificate.
type EndEntityOptions struct {
	// SubjectName is the common name for the certificate.
	// Per Session 3 Q3: Use FQDN style (e.g., "kms.cryptoutil.demo.local").
	SubjectName string

	// Duration is the validity duration for the certificate.
	// Default: 1 year.
	Duration time.Duration

	// DNSNames are the DNS SANs for the certificate.
	DNSNames []string

	// IPAddresses are the IP SANs for the certificate.
	IPAddresses []net.IP

	// EmailAddresses are the email SANs for the certificate.
	EmailAddresses []string

	// URIs are the URI SANs for the certificate.
	URIs []*url.URL

	// KeyUsage specifies the key usage flags.
	KeyUsage x509.KeyUsage

	// ExtKeyUsage specifies the extended key usage flags.
	ExtKeyUsage []x509.ExtKeyUsage

	// Curve specifies the elliptic curve.
	// Default: P-256.
	Curve ECCurve
}

// ServerEndEntityOptions returns options for a TLS server certificate.
func ServerEndEntityOptions(subjectName string, dnsNames []string, ipAddresses []net.IP) *EndEntityOptions {
	return &EndEntityOptions{
		SubjectName: subjectName,
		Duration:    DefaultEndEntityDuration,
		DNSNames:    dnsNames,
		IPAddresses: ipAddresses,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		Curve:       DefaultECCurve,
	}
}

// ClientEndEntityOptions returns options for a TLS client certificate.
func ClientEndEntityOptions(subjectName string) *EndEntityOptions {
	return &EndEntityOptions{
		SubjectName: subjectName,
		Duration:    DefaultEndEntityDuration,
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		Curve:       DefaultECCurve,
	}
}

// CAChain represents a CA certificate chain.
type CAChain struct {
	// CAs is the list of CA subjects, from issuing CA (index 0) to root CA (last index).
	CAs []*cryptoutilSharedCryptoCertificate.Subject

	// IssuingCA is a convenience reference to CAs[0], the CA that issues end entity certs.
	IssuingCA *cryptoutilSharedCryptoCertificate.Subject

	// RootCA is a convenience reference to the last CA in the chain.
	RootCA *cryptoutilSharedCryptoCertificate.Subject
}

// curveToElliptic converts ECCurve to elliptic.Curve.
func curveToElliptic(curve ECCurve) elliptic.Curve {
	switch curve {
	case CurveP384:
		return elliptic.P384()
	case CurveP521:
		return elliptic.P521()
	default:
		return elliptic.P256()
	}
}

// CreateCAChain creates a CA chain with the specified options.
func CreateCAChain(opts *CAChainOptions) (*CAChain, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	} else if opts.ChainLength < 1 {
		return nil, fmt.Errorf("chain length must be at least 1")
	} else if opts.CommonNamePrefix == "" {
		return nil, fmt.Errorf("common name prefix cannot be empty")
	} else if opts.Duration <= 0 {
		return nil, fmt.Errorf("duration must be positive")
	}

	// Validate FQDN style prefix if using FQDN style.
	cnStyle := opts.CNStyle
	if cnStyle == "" {
		cnStyle = DefaultCNStyle
	}

	if cnStyle == CNStyleFQDN {
		if err := ValidateFQDN(opts.CommonNamePrefix); err != nil {
			return nil, fmt.Errorf("invalid common name prefix for FQDN style: %w", err)
		}
	}

	// Generate key pairs for all CAs.
	ellipticCurve := curveToElliptic(opts.Curve)
	keyPairs := make([]*cryptoutilSharedCryptoKeygen.KeyPair, opts.ChainLength)

	for i := range keyPairs {
		keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(ellipticCurve)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair for CA %d: %w", i, err)
		}

		keyPairs[i] = keyPair
	}

	// Build CA subject name prefix based on style.
	// For FQDN: "cryptoutil.demo" → "ca.cryptoutil.demo" (the certificate package adds " 0", " 1" suffixes)
	// For Descriptive: "cryptoutil.demo" → "cryptoutil.demo CA" (the certificate package adds " 0", " 1" suffixes)
	var caSubjectNamePrefix string

	switch cnStyle {
	case CNStyleFQDN:
		caSubjectNamePrefix = "ca." + opts.CommonNamePrefix
	default:
		caSubjectNamePrefix = opts.CommonNamePrefix + " CA"
	}

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects(keyPairs, caSubjectNamePrefix, opts.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	return &CAChain{
		CAs:       caSubjects,
		IssuingCA: caSubjects[0],
		RootCA:    caSubjects[len(caSubjects)-1],
	}, nil
}

// CreateEndEntity creates an end entity certificate signed by the issuing CA.
func (c *CAChain) CreateEndEntity(opts *EndEntityOptions) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	} else if opts.SubjectName == "" {
		return nil, fmt.Errorf("subject name cannot be empty")
	} else if c.IssuingCA == nil {
		return nil, fmt.Errorf("no issuing CA available")
	}

	// Generate key pair for end entity.
	ellipticCurve := curveToElliptic(opts.Curve)

	keyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(ellipticCurve)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	duration := opts.Duration
	if duration <= 0 {
		duration = DefaultEndEntityDuration
	}

	subject, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
		c.IssuingCA,
		keyPair,
		opts.SubjectName,
		duration,
		opts.DNSNames,
		opts.IPAddresses,
		opts.EmailAddresses,
		opts.URIs,
		opts.KeyUsage,
		opts.ExtKeyUsage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create end entity subject: %w", err)
	}

	return subject, nil
}

// RootCAsPool returns a cert pool containing only the root CA.
func (c *CAChain) RootCAsPool() *x509.CertPool {
	pool := x509.NewCertPool()
	if c.RootCA != nil && len(c.RootCA.KeyMaterial.CertificateChain) > 0 {
		pool.AddCert(c.RootCA.KeyMaterial.CertificateChain[0])
	}

	return pool
}

// IntermediateCAsPool returns a cert pool containing intermediate CAs (excluding root).
func (c *CAChain) IntermediateCAsPool() *x509.CertPool {
	pool := x509.NewCertPool()
	// Add all CAs except the root (last one).
	for i := 0; i < len(c.CAs)-1; i++ {
		if len(c.CAs[i].KeyMaterial.CertificateChain) > 0 {
			pool.AddCert(c.CAs[i].KeyMaterial.CertificateChain[0])
		}
	}

	return pool
}
