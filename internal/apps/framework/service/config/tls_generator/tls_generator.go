// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Injectable vars for testing error paths.
var (
	generateECDSAKeyPairFn = func(curve elliptic.Curve) (*cryptoutilSharedCryptoKeygen.KeyPair, error) {
		return cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(curve)
	}
	createCASubjectsFn = func(keyPairs []*cryptoutilSharedCryptoKeygen.KeyPair, commonNamePrefix string, validity time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error) {
		return cryptoutilSharedCryptoCertificate.CreateCASubjects(keyPairs, commonNamePrefix, validity)
	}
	createEndEntitySubjectFn = func(issuer *cryptoutilSharedCryptoCertificate.Subject, keyPair *cryptoutilSharedCryptoKeygen.KeyPair, commonName string, validity time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
		return cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(issuer, keyPair, commonName, validity, dnsNames, ipAddresses, emailAddresses, uris, keyUsage, extKeyUsage)
	}
	buildTLSCertificateFn = func(subject *cryptoutilSharedCryptoCertificate.Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
		return cryptoutilSharedCryptoCertificate.BuildTLSCertificate(subject)
	}
	pemEncodeKeyFn = func(key any) ([]byte, error) {
		return cryptoutilSharedCryptoAsn1.PEMEncode(key)
	}
)

// GenerateTLSMaterial creates TLS configuration based on the specified mode.
//
// Supports three modes:
//   - TLSModeStatic: Uses pre-provided certificate chain and private key.
//   - TLSModeMixed: Uses pre-provided CA to sign dynamically generated server certificate.
//   - TLSModeAuto: Fully auto-generates 3-tier CA hierarchy and server certificate.
//
// Returns TLSMaterial containing tls.Config and certificate pools for client validation.
func GenerateTLSMaterial(cfg *TLSGeneratedSettings) (*cryptoutilAppsFrameworkServiceConfig.TLSMaterial, error) {
	if cfg == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	// Prefer explicit static certificates when provided (server cert + chain + key).
	if len(cfg.StaticCertPEM) > 0 && len(cfg.StaticKeyPEM) > 0 {
		return generateTLSMaterialStatic(cfg)
	}

	// If only partial static material is provided, return informative errors.
	if len(cfg.StaticCertPEM) == 0 && len(cfg.StaticKeyPEM) > 0 {
		return nil, fmt.Errorf("static mode requires StaticCertPEM")
	}

	if len(cfg.StaticKeyPEM) == 0 && len(cfg.StaticCertPEM) > 0 {
		return nil, fmt.Errorf("static mode requires StaticKeyPEM")
	}

	// If only CA material present, instruct caller to pre-generate server certificate.
	if len(cfg.MixedCACertPEM) > 0 || len(cfg.MixedCAKeyPEM) > 0 {
		return nil, fmt.Errorf("mixed CA material provided; please generate server certificate using GenerateServerCertFromCA and supply StaticCertPEM/StaticKeyPEM")
	}

	return nil, fmt.Errorf("no TLS certificate material provided; populate StaticCertPEM and StaticKeyPEM or use helper generation functions")
}

// generateTLSMaterialStatic uses pre-provided TLS certificates (production mode).
func generateTLSMaterialStatic(cfg *TLSGeneratedSettings) (*cryptoutilAppsFrameworkServiceConfig.TLSMaterial, error) {
	// Parse TLS certificate and private key.
	cert, err := tls.X509KeyPair(cfg.StaticCertPEM, cfg.StaticKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static TLS certificate: %w", err)
	}

	// Build certificate pools from chain.
	rootCAPool := x509.NewCertPool()
	intermediateCAPool := x509.NewCertPool()

	// Parse all certificates in chain.
	block, rest := pem.Decode(cfg.StaticCertPEM)

	var certCount int

	for block != nil {
		if block.Type == cryptoutilSharedMagic.StringPEMTypeCertificate {
			parsedCert, parseErr := x509.ParseCertificate(block.Bytes)
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse certificate %d: %w", certCount, parseErr)
			}

			certCount++

			// First cert is leaf (server), skip.
			// Last cert is root CA, add to rootCAPool.
			// Middle certs are intermediates, add to intermediateCAPool.
			if certCount > 1 {
				// Determine if this is a root CA (self-signed: Subject == Issuer).
				if parsedCert.Subject.String() == parsedCert.Issuer.String() {
					rootCAPool.AddCert(parsedCert)
				} else {
					intermediateCAPool.AddCert(parsedCert)
				}
			}
		}

		block, rest = pem.Decode(rest)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		ClientCAs:    rootCAPool,
		ClientAuth:   tls.NoClientCert, // Can be upgraded to RequireAndVerifyClientCert if needed.
	}

	return &cryptoutilAppsFrameworkServiceConfig.TLSMaterial{
		Config:             tlsConfig,
		RootCAPool:         rootCAPool,
		IntermediateCAPool: intermediateCAPool,
	}, nil
}

// GenerateServerCertFromCA generates a server certificate signed by the provided CA
// and returns a TLSGeneratedSettings containing StaticCertPEM (server cert + CA chain)
// and StaticKeyPEM (server private key). The CA material is expected to be PEM-encoded.
func GenerateServerCertFromCA(caCertPEM, caKeyPEM []byte, dns []string, ips []string, validityDays int) (*TLSGeneratedSettings, error) {
	if len(caCertPEM) == 0 {
		return nil, fmt.Errorf("mixed mode requires CA certificate PEM")
	} else if len(caKeyPEM) == 0 {
		return nil, fmt.Errorf("mixed mode requires CA private key PEM")
	}

	// Parse CA certificate chain (use first block as issuer).
	block, _ := pem.Decode(caCertPEM)
	if block == nil || block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse CA private key using shared PEM decoder (handles RSA, EC, PKCS8 automatically).
	caPrivateKey, err := cryptoutilSharedCryptoAsn1.PEMDecode(caKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM: %w", err)
	}

	// Generate server key pair (ECDSA P-384).
	serverKeyPair, err := generateECDSAKeyPairFn(elliptic.P384())
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key pair: %w", err)
	}

	// Parse IP addresses.
	ipAddresses := make([]net.IP, len(ips))
	for i, ipStr := range ips {
		ipAddresses[i] = net.ParseIP(ipStr)
		if ipAddresses[i] == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}
	}

	// Set default validity if not specified.
	if validityDays <= 0 {
		validityDays = cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year
	}

	duration := time.Duration(validityDays) * cryptoutilSharedMagic.HoursPerDay * time.Hour //nolint:mnd // Duration calculation (24 hours per day).

	// Create CA Subject from parsed certificate.
	issuerSubject := &cryptoutilSharedCryptoCertificate.Subject{
		SubjectName: caCert.Subject.CommonName,
		IssuerName:  caCert.Issuer.CommonName,
		IsCA:        true,
		KeyMaterial: cryptoutilSharedCryptoCertificate.KeyMaterial{
			CertificateChain: []*x509.Certificate{caCert},
			PrivateKey:       caPrivateKey,
		},
	}

	// Generate server certificate signed by CA.
	serverSubject, err := createEndEntitySubjectFn(
		issuerSubject,
		serverKeyPair,
		"Server Certificate",
		duration,
		dns,
		ipAddresses,
		nil,
		nil,
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Build TLS certificate from server subject to validate and obtain chains.
	_, _, _, err = buildTLSCertificateFn(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	// Compose PEM chain: server cert first, then issuer certs.
	pemChain, err := encodeCertChainPEM(serverSubject.KeyMaterial.CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode certificate chain: %w", err)
	}

	// PEM-encode server private key.
	privateKeyPEM, err := pemEncodeKeyFn(serverKeyPair.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode server private key: %w", err)
	}

	return &TLSGeneratedSettings{
		StaticCertPEM: pemChain,
		StaticKeyPEM:  privateKeyPEM,
		// Echo back CA material for reference.
		MixedCACertPEM: caCertPEM,
		MixedCAKeyPEM:  caKeyPEM,
	}, nil
}

// GenerateAutoTLSGeneratedSettings creates a new CA hierarchy and a server certificate
// for the provided DNS names and IP addresses, returning a TLSGeneratedSettings with
// StaticCertPEM (server cert + chain) and StaticKeyPEM populated.
func GenerateAutoTLSGeneratedSettings(dns []string, ips []string, validityDays int) (*TLSGeneratedSettings, error) {
	// Default validity if not provided.
	if validityDays <= 0 {
		validityDays = cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year
	}

	// Calculate validity period.
	duration := time.Duration(validityDays) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	// Generate 3-tier CA hierarchy (Root → Intermediate → Server CA keys).
	caKeyPairs := make([]*cryptoutilSharedCryptoKeygen.KeyPair, cryptoutilSharedMagic.DefaultTLSAutoCAChainTiers-1) // Root CA and Intermediate CA (2 CAs).

	var err error

	for i := range caKeyPairs {
		caKeyPairs[i], err = generateECDSAKeyPairFn(elliptic.P384())
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA key pair %d: %w", i, err)
		}
	}

	// Create CA subjects (Root → Intermediate).
	caSubjects, err := createCASubjectsFn(caKeyPairs, "Auto-Generated CA", duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	// CRITICAL: Save issuing CA private key before CreateCASubjects clears it.
	issuingCAPrivateKey := caKeyPairs[len(caKeyPairs)-1].Private

	// Generate server key pair.
	serverKeyPair, err := generateECDSAKeyPairFn(elliptic.P384())
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key pair: %w", err)
	}

	// Parse IP addresses.
	ipAddresses := make([]net.IP, len(ips))
	for i, ipStr := range ips {
		ipAddresses[i] = net.ParseIP(ipStr)
		if ipAddresses[i] == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}
	}

	// Use the issuing CA (last in chain) to sign server certificate.
	// Restore the private key that was cleared by CreateCASubjects.
	issuingCA := caSubjects[len(caSubjects)-1]
	issuingCA.KeyMaterial.PrivateKey = issuingCAPrivateKey

	serverSubject, err := createEndEntitySubjectFn(
		issuingCA,
		serverKeyPair,
		"Auto-Generated Server Certificate",
		duration,
		dns,
		ipAddresses,
		nil, // No email addresses.
		nil, // No URIs.
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Validate by building TLS certificate.
	_, _, _, err = buildTLSCertificateFn(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	// Compose PEM chain from server certificate chain.
	pemChain, err := encodeCertChainPEM(serverSubject.KeyMaterial.CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode certificate chain: %w", err)
	}

	// PEM-encode server private key.
	privateKeyPEM, err := pemEncodeKeyFn(serverKeyPair.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to PEM-encode server private key: %w", err)
	}

	return &TLSGeneratedSettings{
		StaticCertPEM: pemChain,
		StaticKeyPEM:  privateKeyPEM,
	}, nil
}

// GenerateTestCA creates a test CA certificate and private key for use in mixed mode tests.
// Returns CA certificate PEM, CA private key PEM, and any error.
func GenerateTestCA() (caCertPEM []byte, caKeyPEM []byte, err error) {
	// Generate CA key pair using ECDSA P-384.
	caKeyPair, err := generateECDSAKeyPairFn(elliptic.P384())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key pair: %w", err)
	}

	// Create 1-tier CA (just a root CA for simplicity in tests).
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	caSubjects, err := createCASubjectsFn([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	// Get the CA certificate.
	ca := caSubjects[0]
	caCert := ca.KeyMaterial.CertificateChain[0]

	// Encode CA certificate to PEM.
	caCertPEM, err = cryptoutilSharedCryptoAsn1.PEMEncode(caCert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to PEM-encode CA certificate: %w", err)
	}

	// Encode CA private key to PEM.
	caKeyPEM, err = pemEncodeKeyFn(caKeyPair.Private)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to PEM-encode CA private key: %w", err)
	}

	return caCertPEM, caKeyPEM, nil
}

// encodeCertChainPEM creates concatenated PEM from a certificate chain using shared ASN.1 utilities.
func encodeCertChainPEM(certs []*x509.Certificate) ([]byte, error) {
	var pemChain []byte

	for _, cert := range certs {
		pemBytes, err := cryptoutilSharedCryptoAsn1.PEMEncode(cert)
		if err != nil {
			return nil, fmt.Errorf("failed to PEM-encode certificate: %w", err)
		}

		pemChain = append(pemChain, pemBytes...)
	}

	return pemChain, nil
}
