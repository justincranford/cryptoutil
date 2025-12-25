// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"crypto/tls"
	"crypto/x509"
)

// TLSMode defines the three supported TLS certificate provisioning modes.
type TLSMode string

const (
	// TLSModeStatic uses pre-generated TLS certificates (production).
	// Requires: TLS certificate chain (PEM), private key (PEM).
	// Source: Docker secrets, Kubernetes secrets, CA-signed certificates.
	TLSModeStatic TLSMode = "static"

	// TLSModeMixed uses static CA to sign dynamically generated server certificates (staging/QA).
	// Requires: CA certificate chain (PEM), CA private key (PEM).
	// Auto-generates: Server certificate signed by provided CA on startup.
	TLSModeMixed TLSMode = "mixed"

	// TLSModeAuto fully auto-generates CA hierarchy and server certificates (development/testing).
	// Requires: Configuration parameters only (DNS names, IP addresses, validity).
	// Auto-generates: 3-tier CA hierarchy (Root → Intermediate → Server).
	TLSModeAuto TLSMode = "auto"
)

// TLSConfig holds configuration for TLS certificate provisioning.
type TLSConfig struct {
	// Mode determines certificate provisioning strategy.
	Mode TLSMode

	// StaticCertPEM is the PEM-encoded certificate chain (for TLSModeStatic).
	// Should contain: [Server Cert, Intermediate CA(s), Root CA].
	StaticCertPEM []byte

	// StaticKeyPEM is the PEM-encoded private key (for TLSModeStatic).
	StaticKeyPEM []byte

	// MixedCACertPEM is the PEM-encoded CA certificate chain (for TLSModeMixed).
	// Should contain: [Intermediate CA, Root CA] or [Root CA].
	MixedCACertPEM []byte

	// MixedCAKeyPEM is the PEM-encoded CA private key (for TLSModeMixed).
	MixedCAKeyPEM []byte

	// AutoDNSNames are DNS names for auto-generated certificates (for TLSModeAuto and TLSModeMixed).
	AutoDNSNames []string

	// AutoIPAddresses are IP addresses for auto-generated certificates (for TLSModeAuto and TLSModeMixed).
	AutoIPAddresses []string

	// AutoValidityDays is the certificate validity period in days (for TLSModeAuto and TLSModeMixed).
	// Default: 365 days (1 year).
	AutoValidityDays int
}

// TLSMaterial holds the runtime TLS configuration and certificate pools.
type TLSMaterial struct {
	// Config is the tls.Config for HTTPS servers.
	Config *tls.Config

	// RootCAPool is the certificate pool for root CAs (for client certificate validation).
	RootCAPool *x509.CertPool

	// IntermediateCAPool is the certificate pool for intermediate CAs (for chain building).
	IntermediateCAPool *x509.CertPool
}
