// Package magic provides commonly used magic numbers and values as named constants.
// This file contains TLS and certificate-related constants.
package magic

// TLS certificate validity periods.
const (
	// TLSValidityCACertYears - Years for CA certificates.
	TLSValidityCACertYears = 10
	// TLSValidityEndEntityDays - Days for server end-entity certificate.
	TLSValidityEndEntityDays = 397
)

// TLS server configuration.
const (
	// TLSKeyPairsNeeded - Number of keypairs requested for server TLS.
	TLSKeyPairsNeeded = 2
)
