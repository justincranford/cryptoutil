// Package magic provides commonly used magic numbers and values as named constants.
// This file contains TLS and certificate-related constants.
package magic

import "time"

// TLS certificate validity periods.
const (
	// TLSValidityCACertYears - Years for CA certificates.
	TLSValidityCACertYears = 10
	// TLSValidityEndEntityDays - Days for server end-entity certificate.
	TLSValidityEndEntityDays = 397
	// TLSMaxSubscriberCertDuration - Maximum duration for subscriber certificates (398 days).
	TLSMaxSubscriberCertDuration = 398 * 24 * time.Hour
	// TLSMaxCACertDuration - Maximum duration for CA certificates (25 years).
	TLSMaxCACertDuration = 25 * 365 * 24 * time.Hour
)

// TLS server configuration.
const (
	// TLSKeyPairsNeeded - Number of keypairs requested for server TLS.
	TLSKeyPairsNeeded = 2
)
