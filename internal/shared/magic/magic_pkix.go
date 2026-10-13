// Copyright (c) 2025-2026 Justin Cranford.
//
//

package magic

import "time"

// Common duration constants.
const (
	// Days1 represents a 24-hour period.
	Days1 = 24 * time.Hour
	// Days30 represents a 30-day period.
	Days30 = 30 * Days1
	// Days365 represents a 365-day period.
	Days365 = 365 * Days1
)

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
	TLSMaxCACertDuration = TLSMaxValidityCACertYears * Days365

	// TLSDefaultValidityCACertYears - Years for CA certificates.
	TLSDefaultValidityCACertYears = 10
	// TLSDefaultMaxCACertDuration - Maximum duration for CA certificates (10 years).
	TLSDefaultMaxCACertDuration = TLSDefaultValidityCACertYears * Days365

	// TLSMaxValidityEndEntityDays - Maximum days for server end-entity certificate.
	TLSMaxValidityEndEntityDays = 398
	// TLSMaxSubscriberCertDuration - Maximum duration for subscriber certificates (398 days).
	TLSMaxSubscriberCertDuration = TLSMaxValidityEndEntityDays * Days1

	// TLSDefaultValidityEndEntityDays - Days for server end-entity certificate.
	TLSDefaultValidityEndEntityDays = 397
	// TLSDefaultSubscriberCertDuration - Default duration for subscriber certificates (397 days).
	TLSDefaultSubscriberCertDuration = TLSDefaultValidityEndEntityDays * Days1

	// TLSDefaultValidityEndEntityDaysWithRandomizationBuffer is the safe end-entity cert validity
	// period when NotBefore randomization is enabled. CertificateRandomizationNotBeforeMinutes (120)
	// can shift NotBefore backward by up to 2 hours, extending effective validity past the nominal
	// period. Using 396 instead of 397 ensures the actual validity never exceeds
	// TLSMaxValidityEndEntityDays (398) even with maximum randomization applied.
	// Derivation: TLSDefaultValidityEndEntityDays (397) - 1 day buffer for 120-minute randomization.
	TLSDefaultValidityEndEntityDaysWithRandomizationBuffer = TLSDefaultValidityEndEntityDays - 1 // 396

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

	// TLSServerKeyPairsNeeded is the number of ECDSA P-256 key pairs required to generate one TLS
	// server subject: 1 for the end-entity certificate + 1 for the issuing CA certificate.
	TLSServerKeyPairsNeeded = 2
)

// PKI and ASN.1 protocol constants.
const (
	// DaysPerYear is the number of days in the standard non-leap certificate year.
	DaysPerYear = 365
	// Days7 is a seven-day duration.
	Days7 = 7 * Days1
	// ECDSAP256KeySizeBits is the bit size of the NIST P-256 curve.
	ECDSAP256KeySizeBits = 256
	// OIDArcNIST is the second arc of NIST object identifiers.
	OIDArcNIST = 16
	// OIDArcNISTHashAlgorithm is the hash-algorithm arc under NIST algorithm identifiers.
	OIDArcNISTHashAlgorithm = 3
	// PKIMaxTrackedEnrollments bounds the number of enrollment requests retained in memory.
	PKIMaxTrackedEnrollments = 1000
	// PKISerialNumberRandomBytes is the number of random bytes used for TSA serial numbers.
	PKISerialNumberRandomBytes = 8
	// PKITestOIDArc14 is a fixture arc value in PKI ASN.1 identifier tests.
	PKITestOIDArc14 = 14
	// PKITestOIDArc3 is a fixture arc value in PKI ASN.1 identifier tests.
	PKITestOIDArc3 = OIDArcNISTHashAlgorithm
	// PKIProfileIDTLSServer is the default TLS server certificate profile identifier.
	PKIProfileIDTLSServer = "tls-server"
	// PKIProfileIDTLSClient is the default TLS client certificate profile identifier.
	PKIProfileIDTLSClient = "tls-client"
	// TLSName is the canonical lowercase text label for Transport Layer Security.
	TLSName = "tls"

	// PKIProfileCategoryTLS is the category for TLS certificate profiles.
	PKIProfileCategoryTLS = TLSName
	// TLSTestCACertValidity15Years is a 15-year CA certificate test duration in days.
	TLSTestCACertValidity15Years = 15
)

// RFC5280 certificate revocation reason codes.
const (
	RFC5280RevocationReasonUnspecified          = 0
	RFC5280RevocationReasonKeyCompromise        = 1
	RFC5280RevocationReasonCACompromise         = 2
	RFC5280RevocationReasonAffiliationChanged   = 3
	RFC5280RevocationReasonSuperseded           = 4
	RFC5280RevocationReasonCessationOfOperation = 5
	RFC5280RevocationReasonCertificateHold      = 6
	RFC5280RevocationReasonRemoveFromCRL        = 8
	RFC5280RevocationReasonPrivilegeWithdrawn   = 9
	RFC5280RevocationReasonAACompromise         = 10
)

// RFC5280 certificate revocation reason names.
const (
	RFC5280RevocationReasonNameUnspecified          = "unspecified"
	RFC5280RevocationReasonNameKeyCompromise        = "keyCompromise"
	RFC5280RevocationReasonNameCACompromise         = "caCompromise"
	RFC5280RevocationReasonNameAffiliationChanged   = "affiliationChanged"
	RFC5280RevocationReasonNameSuperseded           = "superseded"
	RFC5280RevocationReasonNameCessationOfOperation = "cessationOfOperation"
	RFC5280RevocationReasonNameCertificateHold      = "certificateHold"
	RFC5280RevocationReasonNameRemoveFromCRL        = "removeFromCRL"
	RFC5280RevocationReasonNamePrivilegeWithdrawn   = "privilegeWithdrawn"
	RFC5280RevocationReasonNameAACompromise         = "aaCompromise"
)

// RFC3161 timestamp status codes.
const (
	RFC3161StatusGranted                = 0
	RFC3161StatusGrantedWithMods        = 1
	RFC3161StatusRejection              = 2
	RFC3161StatusWaiting                = 3
	RFC3161StatusRevocationWarning      = 4
	RFC3161StatusRevocationNotification = 5
)

// RFC3161 timestamp status names.
const (
	RFC3161StatusNameGranted                = "granted"
	RFC3161StatusNameGrantedWithMods        = "grantedWithMods"
	RFC3161StatusNameRejection              = "rejection"
	RFC3161StatusNameRevocationWarning      = "revocationWarning"
	RFC3161StatusNameRevocationNotification = "revocationNotification"
)

// RFC3161 timestamp failure codes.
const (
	RFC3161FailureBadAlg              = 0
	RFC3161FailureBadRequest          = 2
	RFC3161FailureBadDataFormat       = 5
	RFC3161FailureTimeNotAvailable    = 14
	RFC3161FailureUnacceptedPolicy    = 15
	RFC3161FailureUnacceptedExtension = 16
	RFC3161FailureAddInfoNotAvailable = 17
	RFC3161FailureSystemFailure       = 25
)

// RFC3161 timestamp failure names.
const (
	RFC3161FailureNameBadAlg              = "badAlg"
	RFC3161FailureNameBadRequest          = "badRequest"
	RFC3161FailureNameBadDataFormat       = "badDataFormat"
	RFC3161FailureNameTimeNotAvailable    = "timeNotAvailable"
	RFC3161FailureNameUnacceptedPolicy    = "unacceptedPolicy"
	RFC3161FailureNameUnacceptedExtension = "unacceptedExtension"
	RFC3161FailureNameAddInfoNotAvailable = "addInfoNotAvailable"
	RFC3161FailureNameSystemFailure       = "systemFailure"
)
