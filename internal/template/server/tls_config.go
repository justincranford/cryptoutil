// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	cryptoutilConfig "cryptoutil/internal/shared/config"
)

// TLSGeneratedSettings holds configuration for TLS certificate provisioning.
type TLSGeneratedSettings struct {
	// Mode determines certificate provisioning strategy.
	Mode cryptoutilConfig.TLSMode

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
