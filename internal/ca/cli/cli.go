// Copyright (c) 2025 Justin Cranford

// Package cli provides command-line interface tools for the CA.
package cli

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CommandOptions holds common options for CA CLI commands.
type CommandOptions struct {
	OutputDir    string
	OutputFormat string // pem, der
	Verbose      bool
	Force        bool
}

// KeyGenOptions holds options for key generation.
type KeyGenOptions struct {
	Algorithm string // RSA, ECDSA, Ed25519
	KeySize   int    // RSA key size or EC curve bits
	Curve     string // P-256, P-384, P-521
}

// CertGenOptions holds options for certificate generation.
type CertGenOptions struct {
	Subject      pkix.Name
	DNSNames     []string
	IPAddresses  []string
	EmailAddrs   []string
	URIs         []string
	ValidityDays int
	IsCA         bool
	PathLen      int
	KeyUsage     x509.KeyUsage
	ExtKeyUsage  []x509.ExtKeyUsage
}

// CLI provides certificate authority command-line operations.
type CLI struct {
	out    io.Writer
	errOut io.Writer
}

// NewCLI creates a new CLI instance.
func NewCLI(out, errOut io.Writer) *CLI {
	if out == nil {
		out = os.Stdout
	}

	if errOut == nil {
		errOut = os.Stderr
	}

	return &CLI{
		out:    out,
		errOut: errOut,
	}
}

// GenerateKey generates a new private key.
func (c *CLI) GenerateKey(_ context.Context, opts *KeyGenOptions, cmdOpts *CommandOptions) (any, error) {
	if opts == nil {
		opts = &KeyGenOptions{Algorithm: "ECDSA", Curve: "P-256"}
	}

	var (
		key any
		err error
	)

	switch opts.Algorithm {
	case "RSA", "rsa":
		keySize := opts.KeySize
		if keySize == 0 {
			keySize = defaultRSAKeySize
		}

		key, err = rsa.GenerateKey(crand.Reader, keySize)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}

	case "ECDSA", "ecdsa", "EC", "ec":
		curve := c.getCurve(opts.Curve)

		key, err = ecdsa.GenerateKey(curve, crand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ECDSA key: %w", err)
		}

	case "Ed25519", "ed25519", "EdDSA", "eddsa":
		_, key, err = ed25519.GenerateKey(crand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Ed25519 key: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", opts.Algorithm)
	}

	// Write to file if output directory specified.
	if cmdOpts != nil && cmdOpts.OutputDir != "" {
		err = c.writeKeyToFile(key, cmdOpts)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

// GenerateSelfSignedCA generates a self-signed CA certificate.
func (c *CLI) GenerateSelfSignedCA(_ context.Context, key any, opts *CertGenOptions, cmdOpts *CommandOptions) (*x509.Certificate, error) {
	if key == nil {
		return nil, errors.New("private key is required")
	}

	if opts == nil {
		opts = &CertGenOptions{
			Subject: pkix.Name{
				CommonName:   "Test CA",
				Organization: []string{"Test"},
			},
			ValidityDays: defaultCAValidityDays,
			IsCA:         true,
		}
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now().UTC()
	validityDays := opts.ValidityDays

	if validityDays == 0 {
		validityDays = defaultCAValidityDays
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               opts.Subject,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, validityDays),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            opts.PathLen,
		MaxPathLenZero:        opts.PathLen == 0,
	}

	pubKey := publicKey(key)

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, pubKey, key)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Write to file if output directory specified.
	if cmdOpts != nil && cmdOpts.OutputDir != "" {
		err = c.writeCertToFile(cert, "ca", cmdOpts)
		if err != nil {
			return nil, err
		}
	}

	return cert, nil
}

// GenerateIntermediateCA generates an intermediate CA certificate.
func (c *CLI) GenerateIntermediateCA(_ context.Context, key any, parentCert *x509.Certificate, parentKey any, opts *CertGenOptions, cmdOpts *CommandOptions) (*x509.Certificate, error) {
	if key == nil {
		return nil, errors.New("private key is required")
	}

	if parentCert == nil {
		return nil, errors.New("parent certificate is required")
	}

	if parentKey == nil {
		return nil, errors.New("parent key is required")
	}

	if opts == nil {
		opts = &CertGenOptions{
			Subject: pkix.Name{
				CommonName:   "Intermediate CA",
				Organization: []string{"Test"},
			},
			ValidityDays: defaultIntermediateValidityDays,
			IsCA:         true,
		}
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now().UTC()
	validityDays := opts.ValidityDays

	if validityDays == 0 {
		validityDays = defaultIntermediateValidityDays
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               opts.Subject,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, validityDays),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	pubKey := publicKey(key)

	certDER, err := x509.CreateCertificate(crand.Reader, template, parentCert, pubKey, parentKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Write to file if output directory specified.
	if cmdOpts != nil && cmdOpts.OutputDir != "" {
		err = c.writeCertToFile(cert, "intermediate", cmdOpts)
		if err != nil {
			return nil, err
		}
	}

	return cert, nil
}

// GenerateEndEntityCert generates an end-entity certificate.
func (c *CLI) GenerateEndEntityCert(_ context.Context, key any, caCert *x509.Certificate, caKey any, opts *CertGenOptions, cmdOpts *CommandOptions) (*x509.Certificate, error) {
	if key == nil {
		return nil, errors.New("private key is required")
	}

	if caCert == nil {
		return nil, errors.New("CA certificate is required")
	}

	if caKey == nil {
		return nil, errors.New("CA key is required")
	}

	if opts == nil {
		return nil, errors.New("certificate options are required")
	}

	serialNumber, err := generateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	now := time.Now().UTC()
	validityDays := opts.ValidityDays

	if validityDays == 0 {
		validityDays = defaultEndEntityValidityDays
	}

	keyUsage := opts.KeyUsage
	if keyUsage == 0 {
		keyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	}

	template := &x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               opts.Subject,
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, validityDays),
		KeyUsage:              keyUsage,
		ExtKeyUsage:           opts.ExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              opts.DNSNames,
		EmailAddresses:        opts.EmailAddrs,
	}

	pubKey := publicKey(key)

	certDER, err := x509.CreateCertificate(crand.Reader, template, caCert, pubKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Write to file if output directory specified.
	if cmdOpts != nil && cmdOpts.OutputDir != "" {
		name := "cert"
		if opts.Subject.CommonName != "" {
			name = sanitizeFilename(opts.Subject.CommonName)
		}

		err = c.writeCertToFile(cert, name, cmdOpts)
		if err != nil {
			return nil, err
		}
	}

	return cert, nil
}

// ValidateCertificate validates a certificate against a CA.
func (c *CLI) ValidateCertificate(_ context.Context, cert *x509.Certificate, roots *x509.CertPool, intermediates *x509.CertPool) error {
	if cert == nil {
		return errors.New("certificate is required")
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now().UTC(),
	}

	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate validation failed: %w", err)
	}

	return nil
}

// getCurve returns the elliptic curve for the given name.
func (c *CLI) getCurve(name string) elliptic.Curve {
	switch name {
	case "P-384", "p384", "384":
		return elliptic.P384()
	case "P-521", "p521", "521":
		return elliptic.P521()
	default:
		return elliptic.P256()
	}
}

// writeKeyToFile writes a private key to a file.
func (c *CLI) writeKeyToFile(key any, opts *CommandOptions) error {
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	filename := filepath.Join(opts.OutputDir, "key.pem")
	if opts.OutputFormat == formatDER {
		filename = filepath.Join(opts.OutputDir, "key.der")

		if err := os.WriteFile(filename, keyDER, filePermKey); err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}

		return nil
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: keyDER,
	})

	if err := os.WriteFile(filename, keyPEM, filePermKey); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// writeCertToFile writes a certificate to a file.
func (c *CLI) writeCertToFile(cert *x509.Certificate, name string, opts *CommandOptions) error {
	filename := filepath.Join(opts.OutputDir, name+".pem")
	if opts.OutputFormat == formatDER {
		filename = filepath.Join(opts.OutputDir, name+".der")

		if err := os.WriteFile(filename, cert.Raw, filePermCert); err != nil {
			return fmt.Errorf("failed to write certificate file: %w", err)
		}

		return nil
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	if err := os.WriteFile(filename, certPEM, filePermCert); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	return nil
}

// File permission constants.
const (
	filePermKey  = 0o600
	filePermCert = 0o644
)

// Output format constants.
const (
	formatDER = "der"
	formatPEM = "pem"
)

// Validity constants.
const (
	defaultRSAKeySize               = 4096
	defaultCAValidityDays           = 3650 // 10 years.
	defaultIntermediateValidityDays = 1825 // 5 years.
	defaultEndEntityValidityDays    = 365  // 1 year.
)

// Serial number generation constants.
const serialNumberBits = 128

// generateSerialNumber generates a random serial number.
func generateSerialNumber() (*big.Int, error) {
	serialNumber := make([]byte, serialNumberBits/8)

	_, err := crand.Read(serialNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Ensure positive by clearing the top bit.
	serialNumber[0] &= 0x7F

	return new(big.Int).SetBytes(serialNumber), nil
}

// publicKey extracts the public key from a private key.
func publicKey(priv any) any {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public()
	default:
		return nil
	}
}

// sanitizeFilename removes/replaces characters unsuitable for filenames.
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))

	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			result = append(result, c)
		} else if c == ' ' || c == '/' || c == '\\' {
			result = append(result, '_')
		}
	}

	return string(result)
}
