// Copyright (c) 2025 Justin Cranford

// Package magic contains magic constants for SM KMS service.
package magic

// SM KMS service constants.
const (
	// OTLPServiceSMKMS is the OTLP service name for sm-kms telemetry.
	OTLPServiceSMKMS = "sm-kms"

	// KMSServiceID is the canonical service identifier for sm-kms.
	KMSServiceID = OTLPServiceSMKMS

	// SMProductName is the product name component of the SM product.
	SMProductName = "sm"

	// KMSServiceName is the service name component of the sm-kms service.
	KMSServiceName = "kms"

	// KMSServicePort is the default public API port for sm-kms service.
	// Port range for SM KMS: 8000-8099 (100-port block).
	KMSServicePort = 8000
)
