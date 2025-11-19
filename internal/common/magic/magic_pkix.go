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
