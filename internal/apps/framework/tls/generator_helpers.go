// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// errTargetDirExists is returned by validateTargetDir when the target directory
// is non-empty, signalling that cert generation was already completed.
var errTargetDirExists = errors.New("target directory already exists")

// validateTargetDir ensures the base path is non-existent or an empty directory.
// Returns errTargetDirExists when the directory already contains files, allowing
// callers to skip generation rather than treating it as a hard error.
// It uses os.Stat first to distinguish "not found" from "exists but is a file".
func (g *Generator) validateTargetDir(basePath string) error {
	fi, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to stat target directory %q: %w", basePath, err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("target directory %q exists but is not a directory", basePath)
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read target directory %q: %w", basePath, err)
	}

	if len(entries) > 0 {
		return errTargetDirExists
	}

	return nil
}

// generateCAChain creates a 2-tier CA pair (root + issuing) under basePath.
// Two keystore directories are created, each with a truststore/ subdirectory:
//
//	{prefix}-root-{suffix}/             (root CA keystore with private key + trust anchor)
//	{prefix}-root-{suffix}/truststore/  (root CA public trust anchor only)
//	{prefix}-issuing-{suffix}/          (issuing CA keystore with private key)
//	{prefix}-issuing-{suffix}/truststore/ (issuing CA public trust anchor only)
//
// Returns the issuing CA Subject for signing leaf certificates.
func (g *Generator) generateCAChain(basePath, prefix, suffix string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	rootKP := g.getKeyPairFn()
	issuingKP := g.getKeyPairFn()

	rootDirBase := prefix + "-root-" + suffix
	issuingDirBase := prefix + "-issuing-" + suffix

	root, err := g.createCAFn(nil, nil, rootDirBase, rootKP, cryptoutilSharedMagic.PKIInitValidityRootCA, 1)
	if err != nil {
		return nil, fmt.Errorf("root CA %s: %w", rootDirBase, err)
	}

	if err := g.writeKeystore(basePath, rootDirBase, root); err != nil {
		return nil, fmt.Errorf("root CA keystore %s: %w", rootDirBase, err)
	}

	if err := g.writeTruststore(basePath, rootDirBase, root); err != nil {
		return nil, fmt.Errorf("root CA truststore %s: %w", rootDirBase, err)
	}

	issuing, err := g.createCAFn(root, root.KeyMaterial.PrivateKey, issuingDirBase, issuingKP, cryptoutilSharedMagic.PKIInitValidityIssuingCA, 0)
	if err != nil {
		return nil, fmt.Errorf("issuing CA %s: %w", issuingDirBase, err)
	}

	if err := g.writeKeystore(basePath, issuingDirBase, issuing); err != nil {
		return nil, fmt.Errorf("issuing CA keystore %s: %w", issuingDirBase, err)
	}

	if err := g.writeTruststore(basePath, issuingDirBase, issuing); err != nil {
		return nil, fmt.Errorf("issuing CA truststore %s: %w", issuingDirBase, err)
	}

	return issuing, nil
}

// generateServerLeafDir creates a keystore directory for a TLS server leaf cert.
func (g *Generator) generateServerLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject, dns []string, ips []net.IP) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, dirName, cryptoutilSharedMagic.PKIInitValidityLeaf, dns, ips, nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return fmt.Errorf("server leaf %s: %w", dirName, err)
	}

	return g.writeKeystore(basePath, dirName, leaf)
}

// generateClientLeafDir creates a keystore directory for a TLS client leaf cert.
func (g *Generator) generateClientLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, dirName, cryptoutilSharedMagic.PKIInitValidityLeaf, nil, nil, nil,
		x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	)
	if err != nil {
		return fmt.Errorf("client leaf %s: %w", dirName, err)
	}

	return g.writeKeystore(basePath, dirName, leaf)
}

// generateMutualLeafDir creates a keystore directory for a combined client+server (mTLS) leaf cert.
func (g *Generator) generateMutualLeafDir(basePath, dirName string, issuer *cryptoutilSharedCryptoCertificate.Subject, dns []string, ips []net.IP) error {
	kp := g.getKeyPairFn()

	leaf, err := g.createLeafFn(issuer, kp, dirName, cryptoutilSharedMagic.PKIInitValidityLeaf, dns, ips, nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	)
	if err != nil {
		return fmt.Errorf("mutual leaf %s: %w", dirName, err)
	}

	return g.writeKeystore(basePath, dirName, leaf)
}

// writeKeystore writes a PKCS#12 keystore directory with 3 files:
//
//	{dirName}.p12  (0440 — PKCS#12 bundle with private key)
//	{dirName}.crt  (0444 — PEM certificate chain)
//	{dirName}.key  (0440 — PEM private key)
func (g *Generator) writeKeystore(basePath, dirName string, subject *cryptoutilSharedCryptoCertificate.Subject) error {
	dir := filepath.Join(basePath, dirName)

	if err := g.mkdirAllFn(dir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	chain := subject.KeyMaterial.CertificateChain

	var caCerts []*x509.Certificate

	if len(chain) > 1 {
		caCerts = chain[1:]
	}

	p12Data, err := g.encodePKCS12Fn(subject.KeyMaterial.PrivateKey, chain[0], caCerts)
	if err != nil {
		return fmt.Errorf("pkcs12 encode %s: %w", dirName, err)
	}

	if err := g.writeFileFn(filepath.Join(dir, dirName+".p12"), p12Data, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode); err != nil {
		return fmt.Errorf("write .p12 %s: %w", dirName, err)
	}

	certPEM, err := cryptoutilSharedCryptoAsn1.PEMEncodeCertChain(chain)
	if err != nil {
		return fmt.Errorf("pem encode cert %s: %w", dirName, err)
	}

	if err := g.writeFileFn(filepath.Join(dir, dirName+".crt"), certPEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode); err != nil {
		return fmt.Errorf("write .crt %s: %w", dirName, err)
	}

	keyPEM, err := cryptoutilSharedCryptoAsn1.PEMEncode(subject.KeyMaterial.PrivateKey)
	if err != nil {
		return fmt.Errorf("pem encode key %s: %w", dirName, err)
	}

	if err := g.writeFileFn(filepath.Join(dir, dirName+".key"), keyPEM, cryptoutilSharedMagic.PKIInitPrivateKeyFileMode); err != nil {
		return fmt.Errorf("write .key %s: %w", dirName, err)
	}

	return nil
}

// writeTruststore writes a PKCS#12 truststore as a subdirectory of the keystore directory.
// The truststore/ subdir is created inside basePath/dirName (the keystore directory).
// Files use SAME-AS-KEYSTORE-DIR-NAME convention and are named after dirName, not "truststore":
//
//	{basePath}/{dirName}/truststore/{dirName}.p12  (0444 — PKCS#12 truststore, no private key)
//	{basePath}/{dirName}/truststore/{dirName}.crt  (0444 — PEM certificate chain)
func (g *Generator) writeTruststore(basePath, dirName string, subject *cryptoutilSharedCryptoCertificate.Subject) error {
	dir := filepath.Join(basePath, dirName, "truststore")

	if err := g.mkdirAllFn(dir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	chain := subject.KeyMaterial.CertificateChain

	p12Data, err := g.encodeTrustPKCS12Fn(chain)
	if err != nil {
		return fmt.Errorf("pkcs12 trust encode %s: %w", dirName, err)
	}

	if err := g.writeFileFn(filepath.Join(dir, dirName+".p12"), p12Data, cryptoutilSharedMagic.PKIInitPublicCertFileMode); err != nil {
		return fmt.Errorf("write .p12 %s: %w", dirName, err)
	}

	certPEM, err := cryptoutilSharedCryptoAsn1.PEMEncodeCertChain(chain)
	if err != nil {
		return fmt.Errorf("pem encode cert %s: %w", dirName, err)
	}

	if err := g.writeFileFn(filepath.Join(dir, dirName+".crt"), certPEM, cryptoutilSharedMagic.PKIInitPublicCertFileMode); err != nil {
		return fmt.Errorf("write .crt %s: %w", dirName, err)
	}

	return nil
}
