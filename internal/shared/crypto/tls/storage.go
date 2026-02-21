// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	storageMkdirAllFn    = os.MkdirAll
	storageWriteFileFn   = os.WriteFile
	storageMarshalPKCS8Fn = func(key any) ([]byte, error) { return x509.MarshalPKCS8PrivateKey(key) }
)

// StorageFormat defines the format for storing certificates and keys.
type StorageFormat string

const (
	// FormatPEM stores certificates and keys in PEM format (default per Session 4 Q2).
	FormatPEM StorageFormat = "pem"

	// FormatPKCS12 stores certificates and keys in PKCS#12 format.
	FormatPKCS12 StorageFormat = "pkcs12"
)

// DefaultStorageFormat is PEM (per Session 4 Q2).
const DefaultStorageFormat = FormatPEM

// StorageOptions holds options for storing certificates and keys.
type StorageOptions struct {
	// Format specifies the storage format (PEM or PKCS#12).
	Format StorageFormat

	// Directory is the directory where files will be stored.
	Directory string

	// CertificateFilename is the filename for the certificate chain.
	// Default: "cert.pem" or "cert.p12".
	CertificateFilename string

	// PrivateKeyFilename is the filename for the private key (PEM only).
	// Default: "key.pem".
	PrivateKeyFilename string

	// IncludePrivateKey determines whether to include the private key.
	IncludePrivateKey bool

	// FileMode is the permission mode for created files.
	// Default: 0600 for private keys, 0644 for certificates.
	FileMode os.FileMode

	// DirMode is the permission mode for created directories.
	// Default: 0755.
	DirMode os.FileMode
}

// DefaultStorageOptions returns storage options with sensible defaults.
func DefaultStorageOptions(directory string) *StorageOptions {
	return &StorageOptions{
		Format:              DefaultStorageFormat,
		Directory:           directory,
		CertificateFilename: "cert.pem",
		PrivateKeyFilename:  "key.pem",
		IncludePrivateKey:   true,
		FileMode:            cryptoutilSharedMagic.FilePermOwnerReadWriteOnly,
		DirMode:             cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
	}
}

// StoredCertificate represents a certificate stored on disk.
type StoredCertificate struct {
	// CertificatePath is the full path to the certificate file.
	CertificatePath string

	// PrivateKeyPath is the full path to the private key file (PEM only).
	// Empty for PKCS#12 format where key is bundled with certificate.
	PrivateKeyPath string

	// Format is the storage format used.
	Format StorageFormat
}

// StoreCertificate stores a certificate subject to disk in the specified format.
func StoreCertificate(subject *cryptoutilSharedCryptoCertificate.Subject, opts *StorageOptions) (*StoredCertificate, error) {
	if subject == nil {
		return nil, fmt.Errorf("subject cannot be nil")
	} else if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	} else if opts.Directory == "" {
		return nil, fmt.Errorf("directory cannot be empty")
	}

	// Create directory if it doesn't exist.
	if err := storageMkdirAllFn(opts.Directory, opts.DirMode); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", opts.Directory, err)
	}

	switch opts.Format {
	case FormatPEM:
		return storePEM(subject, opts)
	case FormatPKCS12:
		return storePKCS12(subject, opts)
	default:
		return nil, fmt.Errorf("unsupported storage format: %s", opts.Format)
	}
}

// storePEM stores a certificate in PEM format.
func storePEM(subject *cryptoutilSharedCryptoCertificate.Subject, opts *StorageOptions) (*StoredCertificate, error) {
	certPath := filepath.Join(opts.Directory, opts.CertificateFilename)
	keyPath := filepath.Join(opts.Directory, opts.PrivateKeyFilename)

	// Build certificate chain PEM.
	var certChainPEM []byte

	for _, cert := range subject.KeyMaterial.CertificateChain {
		block := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		certChainPEM = append(certChainPEM, pem.EncodeToMemory(block)...)
	}

	// Write certificate chain.
	if err := storageWriteFileFn(certPath, certChainPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteGroupRead); err != nil {
		return nil, fmt.Errorf("failed to write certificate: %w", err)
	}

	stored := &StoredCertificate{
		CertificatePath: certPath,
		Format:          FormatPEM,
	}

	// Write private key if requested.
	if opts.IncludePrivateKey && subject.KeyMaterial.PrivateKey != nil { //nolint:gosec // Explicit check for private key
		keyDER, err := storageMarshalPKCS8Fn(subject.KeyMaterial.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key: %w", err)
		}

		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyDER,
		})

		if err := storageWriteFileFn(keyPath, keyPEM, opts.FileMode); err != nil {
			return nil, fmt.Errorf("failed to write private key: %w", err)
		}

		stored.PrivateKeyPath = keyPath
	}

	return stored, nil
}

// storePKCS12 stores a certificate in PKCS#12 format.
// PKCS#12 (also known as PFX) is a binary format that stores certificates with private keys.
//
// Future implementation will:
//   - Use software.sslmate.com/src/go-pkcs12 for PKCS#12 encoding
//   - Support password-protected keystores
//   - Enable cross-platform certificate exchange (Windows, macOS, browsers)
//
// Note: PKCS#11 (HSM) and YubiKey support planned for future phases.
// Reference: Session 4 Q2 notes.
func storePKCS12(_ *cryptoutilSharedCryptoCertificate.Subject, _ *StorageOptions) (*StoredCertificate, error) {
	return nil, fmt.Errorf("PKCS#12 storage format not yet implemented (planned for future)")
}

// LoadCertificatePKCS12 loads a certificate and private key from a PKCS#12 file.
// This is a placeholder for future implementation.
//
// Future implementation will:
//   - Use software.sslmate.com/src/go-pkcs12 for PKCS#12 decoding
//   - Support password-protected keystores
//   - Parse certificate chains and private keys from PFX/P12 files
//
// Reference: Session 4 Q2 notes.
func LoadCertificatePKCS12(_ string, _ string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	return nil, fmt.Errorf("PKCS#12 loading not yet implemented (planned for future)")
}

// LoadCertificatePEM loads a certificate and private key from PEM files.
func LoadCertificatePEM(certPath, keyPath string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	// Read certificate file.
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Parse certificate chain.
	var certs []*x509.Certificate

	for {
		block, rest := pem.Decode(certPEM)
		if block == nil {
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %w", err)
			}

			certs = append(certs, cert)
		}

		certPEM = rest
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in file")
	}

	subject := &cryptoutilSharedCryptoCertificate.Subject{
		SubjectName: certs[0].Subject.CommonName,
		IssuerName:  certs[0].Issuer.CommonName,
		Duration:    certs[0].NotAfter.Sub(certs[0].NotBefore),
		IsCA:        certs[0].IsCA,
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			CertificateChain: certs,
			PublicKey:        certs[0].PublicKey,
		},
	}

	// Read private key if path provided.
	if keyPath != "" {
		keyPEM, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}

		block, _ := pem.Decode(keyPEM)
		if block == nil {
			return nil, fmt.Errorf("failed to decode private key PEM")
		}

		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		subject.KeyMaterial.PrivateKey = key
	}

	// Set additional fields from certificate.
	if !certs[0].IsCA {
		subject.DNSNames = certs[0].DNSNames
		subject.IPAddresses = certs[0].IPAddresses
		subject.EmailAddresses = certs[0].EmailAddresses
		subject.URIs = certs[0].URIs
	} else {
		subject.MaxPathLen = certs[0].MaxPathLen
	}

	return subject, nil
}
