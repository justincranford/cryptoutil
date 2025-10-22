// Package magic provides commonly used magic numbers and values as named constants.
// This file contains TLS and certificate-related constants.
package magic

import "time"

// TLS certificate validity periods.
const (
	// TLSMaxValidityCACertYears - Maximum years for CA certificates.
	TLSMaxValidityCACertYears = 25
	// TLSMaxCACertDuration - Maximum duration for CA certificates (25 years).
	TLSMaxCACertDuration = TLSMaxValidityCACertYears * 365 * 24 * time.Hour

	// TLSDefaultValidityCACertYears - Years for CA certificates.
	TLSDefaultValidityCACertYears = 10
	// TLSDefaultMaxCACertDuration - Maximum duration for CA certificates (25 years).
	TLSDefaultMaxCACertDuration = TLSMaxValidityCACertYears * 365 * 24 * time.Hour

	// TLSMaxValidityEndEntityDays - Maximum days for server end-entity certificate.
	TLSMaxValidityEndEntityDays = 398
	// TLSMaxSubscriberCertDuration - Maximum duration for subscriber certificates (398 days).
	TLSMaxSubscriberCertDuration = TLSMaxValidityEndEntityDays * 24 * time.Hour

	// TLSDefaultValidityEndEntityDays - Days for server end-entity certificate.
	TLSDefaultValidityEndEntityDays = 397
	// TLSDefaultSubscriberCertDuration - Maximum duration for subscriber certificates (398 days).
	TLSDefaultSubscriberCertDuration = TLSDefaultValidityEndEntityDays * 24 * time.Hour
)
