// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "time"

// TLS certificate validity periods.
const (
	// Serial number bit sizes for cryptographic range.
	MinSerialNumberBits = 64
	MaxSerialNumberBits = 159

	// HoursPerDay - Number of hours in a day.
	HoursPerDay = 24

	// CertificateRandomizationNotBeforeMinutes - Certificate validity randomization range in minutes.
	CertificateRandomizationNotBeforeMinutes = 120

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

	// Test certificate validity durations.
	TLSTestCACertValidity20Years        = 20
	TLSTestCACertValidity5Years         = 5
	TLSTestEndEntityCertValidity396Days = 396
	TLSTestEndEntityCertValidity30Days  = 30
	TLSTestEndEntityCertValidity1Year   = 365
	TLSTestEndEntityCertValidity1Day    = 1

	// TLS test self-signed certificate serial number bits (128 bits for testing).
	TLSSelfSignedCertSerialNumberBits = 128

	// StringPEMTypePKCS8PrivateKey - PKCS8 private key PEM type.
	StringPEMTypePKCS8PrivateKey = "PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypePKIXPublicKey - PKIX public key PEM type.
	StringPEMTypePKIXPublicKey = "PUBLIC KEY"
	// StringPEMTypeRSAPrivateKey - RSA private key PEM type.
	StringPEMTypeRSAPrivateKey = "RSA PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypeRSAPublicKey - RSA public key PEM type.
	StringPEMTypeRSAPublicKey = "RSA PUBLIC KEY"
	// StringPEMTypeECPrivateKey - EC private key PEM type.
	StringPEMTypeECPrivateKey = "EC PRIVATE KEY" // pragma: allowlist secret
	// StringPEMTypeCertificate - Certificate PEM type.
	StringPEMTypeCertificate = "CERTIFICATE"
	// StringPEMTypeCSR - Certificate signing request PEM type.
	StringPEMTypeCSR = "CERTIFICATE REQUEST"
	// StringPEMTypeSecretKey - Secret key PEM type.
	StringPEMTypeSecretKey = "SECRET KEY" // pragma: allowlist secret
)

// PKI certificate generation constants.
const (
	// ISO3166Alpha2CountryCodeLength - ISO 3166-1 alpha-2 country code length (2 characters).
	ISO3166Alpha2CountryCodeLength = 2

	// PKICASerialNumberBits - Default serial number bit length for CA-issued certificates.
	// CA/Browser Forum Baseline Requirements: minimum 64 bits, recommended 128 bits of entropy.
	PKICASerialNumberBits = 128

	// DefaultTLSAutoCAChainTiers - Default number of CA chain tiers for auto-generated TLS certificates.
	// Tier layout: Root CA + Intermediate CA + End Entity certificate = 3 tiers.
	DefaultTLSAutoCAChainTiers = 3
)
