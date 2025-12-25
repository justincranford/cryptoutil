// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"crypto/elliptic"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"time"

	cryptoutilCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
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
func GenerateTLSMaterial(cfg *TLSConfig) (*TLSMaterial, error) {
	if cfg == nil {
		return nil, fmt.Errorf("TLS config cannot be nil")
	}

	switch cfg.Mode {
	case TLSModeStatic:
		return generateTLSMaterialStatic(cfg)
	case TLSModeMixed:
		return generateTLSMaterialMixed(cfg)
	case TLSModeAuto:
		return generateTLSMaterialAuto(cfg)
	default:
		return nil, fmt.Errorf("unknown TLS mode: %s (must be static, mixed, or auto)", cfg.Mode)
	}
}

// generateTLSMaterialStatic uses pre-provided TLS certificates (production mode).
func generateTLSMaterialStatic(cfg *TLSConfig) (*TLSMaterial, error) {
	if len(cfg.StaticCertPEM) == 0 {
		return nil, fmt.Errorf("static mode requires StaticCertPEM")
	}

	if len(cfg.StaticKeyPEM) == 0 {
		return nil, fmt.Errorf("static mode requires StaticKeyPEM")
	}

	// Parse TLS certificate and private key.
	cert, err := tls.X509KeyPair(cfg.StaticCertPEM, cfg.StaticKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static TLS certificate: %w", err)
	}

	// Parse leaf certificate for validation.
	if cert.Leaf == nil && len(cert.Certificate) > 0 {
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			return nil, fmt.Errorf("failed to parse leaf certificate: %w", err)
		}
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

	return &TLSMaterial{
		Config:             tlsConfig,
		RootCAPool:         rootCAPool,
		IntermediateCAPool: intermediateCAPool,
	}, nil
}

// generateTLSMaterialMixed uses static CA to sign dynamically generated server certificate (staging/QA mode).
func generateTLSMaterialMixed(cfg *TLSConfig) (*TLSMaterial, error) {
	if len(cfg.MixedCACertPEM) == 0 {
		return nil, fmt.Errorf("mixed mode requires MixedCACertPEM")
	}

	if len(cfg.MixedCAKeyPEM) == 0 {
		return nil, fmt.Errorf("mixed mode requires MixedCAKeyPEM")
	}

	// Parse CA certificate chain.
	block, _ := pem.Decode(cfg.MixedCACertPEM)
	if block == nil || block.Type != pemTypeCertificate {
		return nil, fmt.Errorf("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Parse CA private key.
	keyBlock, _ := pem.Decode(cfg.MixedCAKeyPEM)
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
	serverKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key pair: %w", err)
	}

	// Parse IP addresses.
	ipAddresses := make([]net.IP, len(cfg.AutoIPAddresses))
	for i, ipStr := range cfg.AutoIPAddresses {
		ipAddresses[i] = net.ParseIP(ipStr)
		if ipAddresses[i] == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}
	}

	// Set default validity if not specified.
	validityDays := cfg.AutoValidityDays
	if validityDays <= 0 {
		const defaultValidityDays = 365

		validityDays = defaultValidityDays
	}

	duration := time.Duration(validityDays) * 24 * time.Hour //nolint:mnd // Duration calculation (24 hours per day).

	// Create CA Subject from parsed certificate.
	issuerSubject := &cryptoutilCertificate.Subject{
		SubjectName: caCert.Subject.CommonName,
		IssuerName:  caCert.Issuer.CommonName,
		IsCA:        true,
		KeyMaterial: cryptoutilCertificate.KeyMaterial{
			CertificateChain: []*x509.Certificate{caCert},
			PrivateKey:       caPrivateKey,
		},
	}

	// Generate server certificate signed by CA.
	serverSubject, err := cryptoutilCertificate.CreateEndEntitySubject(
		issuerSubject,
		serverKeyPair,
		"Server Certificate",
		duration,
		cfg.AutoDNSNames,
		ipAddresses,
		nil, // No email addresses.
		nil, // No URIs.
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Build TLS certificate from server subject.
	tlsCert, rootCAPool, intermediateCAPool, err := cryptoutilCertificate.BuildTLSCertificate(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		MinVersion:   tls.VersionTLS13,
		ClientCAs:    rootCAPool,
		ClientAuth:   tls.NoClientCert,
	}

	return &TLSMaterial{
		Config:             tlsConfig,
		RootCAPool:         rootCAPool,
		IntermediateCAPool: intermediateCAPool,
	}, nil
}

// generateTLSMaterialAuto fully auto-generates CA hierarchy and server certificate (development/testing mode).
func generateTLSMaterialAuto(cfg *TLSConfig) (*TLSMaterial, error) {
	// Set default validity if not specified.
	validityDays := cfg.AutoValidityDays
	if validityDays <= 0 {
		const defaultValidityDays = 365

		validityDays = defaultValidityDays
	}

	duration := time.Duration(validityDays) * 24 * time.Hour //nolint:mnd // Duration calculation (24 hours per day).

	// Generate 3-tier CA hierarchy (Root → Intermediate → Server).
	const caTiers = 3

	caKeyPairs := make([]*cryptoutilKeyGen.KeyPair, caTiers-1) // Root CA and Intermediate CA (2 CAs).

	var err error

	for i := range caKeyPairs {
		caKeyPairs[i], err = cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
		if err != nil {
			return nil, fmt.Errorf("failed to generate CA key pair %d: %w", i, err)
		}
	}

	// Create CA subjects (Root → Intermediate).
	caSubjects, err := cryptoutilCertificate.CreateCASubjects(caKeyPairs, "Auto-Generated CA", duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA subjects: %w", err)
	}

	// Generate server key pair.
	serverKeyPair, err := cryptoutilKeyGen.GenerateECDSAKeyPair(elliptic.P384())
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key pair: %w", err)
	}

	// Parse IP addresses.
	ipAddresses := make([]net.IP, len(cfg.AutoIPAddresses))
	for i, ipStr := range cfg.AutoIPAddresses {
		ipAddresses[i] = net.ParseIP(ipStr)
		if ipAddresses[i] == nil {
			return nil, fmt.Errorf("invalid IP address: %s", ipStr)
		}
	}

	// Use the issuing CA (last in chain) to sign server certificate.
	issuingCA := caSubjects[len(caSubjects)-1]

	serverSubject, err := cryptoutilCertificate.CreateEndEntitySubject(
		issuingCA,
		serverKeyPair,
		"Auto-Generated Server Certificate",
		duration,
		cfg.AutoDNSNames,
		ipAddresses,
		nil, // No email addresses.
		nil, // No URIs.
		x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Build TLS certificate from server subject.
	tlsCert, rootCAPool, intermediateCAPool, err := cryptoutilCertificate.BuildTLSCertificate(serverSubject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		MinVersion:   tls.VersionTLS13,
		ClientCAs:    rootCAPool,
		ClientAuth:   tls.NoClientCert,
	}

	return &TLSMaterial{
		Config:             tlsConfig,
		RootCAPool:         rootCAPool,
		IntermediateCAPool: intermediateCAPool,
	}, nil
}
