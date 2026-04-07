// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// validateTargetDir ensures the target directory is non-existent or empty.
func (g *Generator) validateTargetDir(targetDir string) error {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to read target directory %q: %w", targetDir, err)
	}

	if len(entries) > 0 {
		return fmt.Errorf("target directory %q is not empty", targetDir)
	}

	return nil
}

// generateCAChain generates a 2-tier CA (root + issuing) under parentDir/.
func (g *Generator) generateCAChain(parentDir, name string, validity time.Duration) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	rootKP := g.getKeyPairFn()
	issuingKP := g.getKeyPairFn()

	rootName := name + "-ca-root"
	issuingName := name + "-ca-issuing"

	root, err := g.createCAFn(nil, nil, rootName, rootKP, validity, 1)
	if err != nil {
		return nil, fmt.Errorf("root CA %s: %w", rootName, err)
	}

	if err := g.writeCertAndKey(filepath.Join(parentDir, rootName), rootName, root); err != nil {
		return nil, err
	}

	issuing, err := g.createCAFn(root, root.KeyMaterial.PrivateKey, issuingName, issuingKP, validity, 0)
	if err != nil {
		return nil, fmt.Errorf("issuing CA %s: %w", issuingName, err)
	}

	if err := g.writeCertAndKey(filepath.Join(parentDir, issuingName), issuingName, issuing); err != nil {
		return nil, err
	}

	return issuing, nil
}

// generateServerLeaf generates a TLS server leaf cert issued by the given CA.
func (g *Generator) generateServerLeaf(parentDir, name string, issuer *cryptoutilSharedCryptoCertificate.Subject, validity time.Duration, dns []string, ips []net.IP) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, name, validity, dns, ips, nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return fmt.Errorf("server leaf %s: %w", name, err)
	}

	return g.writeCertAndKey(filepath.Join(parentDir, name), name, leaf)
}

// generateClientLeaf generates a TLS client leaf cert issued by the given CA.
func (g *Generator) generateClientLeaf(parentDir, dirName, leafName string, issuer *cryptoutilSharedCryptoCertificate.Subject, validity time.Duration) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, leafName, validity, nil, nil, nil,
		x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	)
	if err != nil {
		return fmt.Errorf("client leaf %s: %w", leafName, err)
	}

	return g.writeCertAndKey(filepath.Join(parentDir, dirName), leafName, leaf)
}

// generateMutualLeaf generates a combined client+server TLS leaf cert.
func (g *Generator) generateMutualLeaf(parentDir, name string, issuer *cryptoutilSharedCryptoCertificate.Subject, validity time.Duration, dns []string, ips []net.IP) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, name, validity, dns, ips, nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	)
	if err != nil {
		return fmt.Errorf("mutual leaf %s: %w", name, err)
	}

	return g.writeCertAndKey(filepath.Join(parentDir, name), name, leaf)
}

// writeIssuerCert writes a CA's cert chain PEM (without private key) to a named subdirectory.
// Used for trust anchor copies per the tls-structure.md layout.
func (g *Generator) writeIssuerCert(parentDir, name string, ca *cryptoutilSharedCryptoCertificate.Subject) error {
	dir := filepath.Join(parentDir, name)

	if err := g.mkdirAllFn(dir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	certPEM, err := encodeCertChainPEM(ca)
	if err != nil {
		return fmt.Errorf("failed to PEM-encode issuer cert chain %s: %w", name, err)
	}

	certPath := filepath.Join(dir, name+"-crt.pem")

	if err := g.writeFileFn(certPath, certPEM, cryptoutilSharedMagic.PKIInitCertFileMode); err != nil {
		return fmt.Errorf("failed to write issuer cert %s: %w", certPath, err)
	}

	return nil
}

// writeCertAndKey writes a certificate chain PEM and private key PEM to the specified directory.
func (g *Generator) writeCertAndKey(dir, baseName string, subject *cryptoutilSharedCryptoCertificate.Subject) error {
	if err := g.mkdirAllFn(dir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	certPEM, err := encodeCertChainPEM(subject)
	if err != nil {
		return fmt.Errorf("failed to PEM-encode cert chain for %s: %w", baseName, err)
	}

	certPath := filepath.Join(dir, baseName+"-crt.pem")

	if err := g.writeFileFn(certPath, certPEM, cryptoutilSharedMagic.PKIInitCertFileMode); err != nil {
		return fmt.Errorf("failed to write cert %s: %w", certPath, err)
	}

	if subject.KeyMaterial.PrivateKey != nil {
		keyPEM, err := cryptoutilSharedCryptoAsn1.PEMEncode(subject.KeyMaterial.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to encode private key for %s: %w", baseName, err)
		}

		keyPath := filepath.Join(dir, baseName+"-key.pem")

		if err := g.writeFileFn(keyPath, keyPEM, cryptoutilSharedMagic.PKIInitCertFileMode); err != nil {
			return fmt.Errorf("failed to write key %s: %w", keyPath, err)
		}
	}

	return nil
}
