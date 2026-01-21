// Copyright (c) 2025 Justin Cranford

// Package tls_generator provides TLS certificate generation and provisioning utilities.
package tls_generator

// TLSGeneratedSettings holds configuration for TLS certificate provisioning.
type TLSGeneratedSettings struct {
	// StaticCertPEM is the PEM-encoded certificate chain (server certificate followed by intermediates and root).
	// Should contain: [Server Cert, Intermediate CA(s), Root CA].
	StaticCertPEM []byte

	// StaticKeyPEM is the PEM-encoded private key (for server certificate).
	StaticKeyPEM []byte

	// MixedCACertPEM is the PEM-encoded CA certificate chain (for staging/QA where CA signs server certs).
	// Should contain: [Intermediate CA(s), Root CA] or [Root CA].
	MixedCACertPEM []byte

	// MixedCAKeyPEM is the PEM-encoded CA private key (for staging/QA where CA signs server certs).
	MixedCAKeyPEM []byte
}
