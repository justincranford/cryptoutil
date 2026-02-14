// Copyright (c) 2025 Justin Cranford

// Package magic contains magic constants for PKI CA service.
package magic

// PKI CA service constants.
const (
	// OTLPServicePKICA is the OTLP service name for pki-ca telemetry.
	OTLPServicePKICA = "pki-ca"

	// PKICAServicePort is the default public API port for pki-ca service.
	// Port range for PKI CA: 8100-8199 (100-port block).
	PKICAServicePort = 8100
)
