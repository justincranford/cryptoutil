// Copyright (c) 2025 Justin Cranford

package tls_generator

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	// pemTypeCertificate is the PEM type identifier for X.509 certificates.
	pemTypeCertificate = "CERTIFICATE"
)

// GenerateTLSMaterial creates TLS configuration based on the specified mode.
//
// Supports three modes:
//   - TLSModeStatic: Uses pre-provided certificate chain and private key.
//   - TLSModeMixed: Uses pre-provided CA to sign dynamically generated server certificate.
//   - TLSModeAuto: Fully auto-generates 3-tier CA hierarchy and server certificate.
//
// Returns TLSMaterial containing tls.Config and certificate pools for client validation.
func GenerateTLSMaterial(cfg *TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
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
func generateTLSMaterialStatic(cfg *TLSGeneratedSettings) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
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
		if block.Type == pemTypeCertificate {
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

	return &cryptoutilAppsTemplateServiceConfig.TLSMaterial{
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
	if block == nil || block.Type != pemTypeCertificate {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse CA private key.
	keyBlock, _ := pem.Decode(caKeyPEM)
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	var caPrivateKey any

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		caPrivateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		caPrivateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		caPrivateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		return nil, fmt.Errorf("unsupported CA private key type: %s", keyBlock.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Generate server key pair (ECDSA P-384).
	serverKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
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
		validityDays = 365
	}

	duration := time.Duration(validityDays) * 24 * time.Hour //nolint:mnd // Duration calculation (24 hours per day).

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
	serverSubject, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
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
	_, _, _, err = cryptoutilSharedCryptoCertificate.BuildTLSCertificate(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	// Compose PEM chain: server cert first, then issuer certs (if any) from issuerSubject.KeyMaterial.CertificateChain
	var pemChain []byte
	for _, cert := range serverSubject.KeyMaterial.CertificateChain {
		pemChain = append(pemChain, pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: cert.Raw})...)
	}

	// Private key PEM
	// Marshal PKCS8 private key DER then encode to PEM.
	dk, err := x509.MarshalPKCS8PrivateKey(serverKeyPair.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server private key to DER: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: dk})

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
		validityDays = 365
	}

	// Calculate validity period.
	duration := time.Duration(validityDays) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	// Generate 3-tier CA hierarchy (Root → Intermediate → Server CA keys).
	caKeyPairs := make([]*cryptoutilSharedCryptoKeygen.KeyPair, cryptoutilSharedMagic.DefaultTLSAutoCAChainTiers-1) // Root CA and Intermediate CA (2 CAs).

	var err error

	for i := range caKeyPairs {
		caKeyPairs[i], err = cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA key pair %d: %w", i, err)
		}
	}

	// Create CA subjects (Root → Intermediate).
	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects(caKeyPairs, "Auto-Generated CA", duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	// CRITICAL: Save issuing CA private key before CreateCASubjects clears it.
	issuingCAPrivateKey := caKeyPairs[len(caKeyPairs)-1].Private

	// Generate server key pair.
	serverKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
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

	serverSubject, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
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
	_, _, _, err = cryptoutilSharedCryptoCertificate.BuildTLSCertificate(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	// Compose PEM chain from serverSubject.KeyMaterial.CertificateChain
	var pemChain []byte
	for _, cert := range serverSubject.KeyMaterial.CertificateChain {
		pemChain = append(pemChain, pem.EncodeToMemory(&pem.Block{Type: pemTypeCertificate, Bytes: cert.Raw})...)
	}

	// Marshal server private key to PKCS8 DER then PEM.
	dk, err := x509.MarshalPKCS8PrivateKey(serverKeyPair.Private)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal server private key to DER: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: dk})

	return &TLSGeneratedSettings{
		StaticCertPEM: pemChain,
		StaticKeyPEM:  privateKeyPEM,
	}, nil
}

// GenerateTestCA creates a test CA certificate and private key for use in mixed mode tests.
// Returns CA certificate PEM, CA private key PEM, and any error.
func GenerateTestCA() (caCertPEM []byte, caKeyPEM []byte, err error) {
	// Generate CA key pair using ECDSA P-384.
	caKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPair(elliptic.P384())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate CA key pair: %w", err)
	}

	// Create 1-tier CA (just a root CA for simplicity in tests).
	duration := time.Duration(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year) * cryptoutilSharedMagic.HoursPerDay * time.Hour

	caSubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects([]*cryptoutilSharedCryptoKeygen.KeyPair{caKeyPair}, "Test CA", duration)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	// Get the CA certificate.
	ca := caSubjects[0]
	caCert := ca.KeyMaterial.CertificateChain[0]

	// Encode CA certificate to PEM.
	caCertPEM = pem.EncodeToMemory(&pem.Block{
		Type:  pemTypeCertificate,
		Bytes: caCert.Raw,
	})

	// Encode CA private key to PKCS8 PEM.
	caKeyBytes, err := x509.MarshalPKCS8PrivateKey(caKeyPair.Private)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal CA private key: %w", err)
	}

	caKeyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caKeyBytes,
	})

	return caCertPEM, caKeyPEM, nil
}
