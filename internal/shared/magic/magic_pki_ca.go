// Copyright (c) 2025-2026 Justin Cranford.
// Package magic provides shared constants for the CA package.
package magic

import "time"

// BackdateBuffer allows slight backdating to handle clock skew.
const BackdateBuffer = 1 * time.Minute

// HexBase is the base for hex encoding serial numbers.
const HexBase = 16

// SerialNumberLength is 20 bytes (160 bits) per CA/Browser Forum requirements.
const SerialNumberLength = 20

// DefaultPageLimit is the default number of items per page for pagination.
const DefaultPageLimit = 20

// BitsPerByte is the number of bits per byte (for key size calculations).
const BitsPerByte = 8

// CA_ITEMS is the sm-kms PKI table name for CAItem persistence model.
const CA_ITEMS = "ca_items"

const (
	PKICertificatesTable             = "pki_certificates"
	PKIAuthoritiesTable              = "pki_authorities"
	PKICRLSequencesTable             = "pki_crl_sequences"
	PKIDefaultAuthorityID            = "sm-kms-pki-development-authority"
	PKICertificateStatusIndex        = "idx_pki_certificates_tenant_status"
	PKICertificateSerialIndex        = "idx_pki_certificates_tenant_serial"
	PKICertificateRequestLegacyIndex = "idx_pki_certificates_tenant_request"
	PKICertificateRequestIndex       = "idx_pki_certificates_tenant_request_unique"
	PKIPrivateKeyEnvelopeVersionV1   = 1
	PKIErrorRemoveFromCRLUnsupported = "removeFromCRL is unsupported because only complete CRLs are generated"
)

const (
	// PKICRLFormatDER selects DER-encoded CRL output.
	PKICRLFormatDER = "der"
	// PKICRLFormatPEM selects PEM-encoded CRL output.
	PKICRLFormatPEM = "pem"
	// PKIHeaderContentDisposition identifies a downloadable PKI response.
	PKIHeaderContentDisposition = "Content-Disposition"
	// PKIHeaderContentTransferEncoding identifies the EST body transfer encoding.
	PKIHeaderContentTransferEncoding = "Content-Transfer-Encoding"
	// PKIContentTransferEncodingBase64 identifies base64-encoded EST bodies.
	PKIContentTransferEncodingBase64 = "base64"
	// PKIContentTypeCRLDER identifies DER-encoded CRL output.
	PKIContentTypeCRLDER = "application/pkix-crl"
	// PKIContentTypeCRLPEM identifies PEM-encoded CRL output.
	PKIContentTypeCRLPEM = "application/x-pem-file"
	// PKIContentTypeOCSPResponse identifies OCSP response payloads.
	PKIContentTypeOCSPResponse = "application/ocsp-response"
	// PKIContentTypePKCS7CertsOnly identifies EST CA certificate responses.
	PKIContentTypePKCS7CertsOnly = "application/pkcs7-mime; smime-type=certs-only"
	// PKIContentTypeTimestampReply identifies RFC 3161 timestamp responses.
	PKIContentTypeTimestampReply = "application/timestamp-reply"
	// PKIMessageFailedToParseRequestBody describes malformed JSON request bodies.
	PKIMessageFailedToParseRequestBody = "failed to parse request body"
)

// PKI CA hierarchy and status labels.
const (
	ROOT         = "root"
	INTERMEDIATE = "intermediate"
	ISSUING      = "issuing"
	ACTIVE       = "active"
	PENDING      = "pending"
	SUSPENDED    = "suspended"
	REVOKED      = "revoked"
	EXPIRED      = "expired"
)

// PKI CA test labels and curve aliases.
const (
	TEST_CA_NAME = "test-ca"
	TEST_CA_CN   = "Test CA"
	P256         = "P-256"
	P384         = "P-384"
	P521         = "P-521"
)

// PKI threat model identifiers and classifications.
const (
	PKIAssetIDRootCAKey               = "ASSET-001"
	PKIAssetIDIntermediateCAKey       = "ASSET-002"
	PKIAssetIDCertificateDatabase     = "ASSET-003"
	PKIAssetIDAuditLogs               = "ASSET-004"
	PKIThreatIDSpoofing               = "THREAT-S-001"
	PKIThreatIDTampering              = "THREAT-T-001"
	PKIThreatIDRepudiation            = "THREAT-R-001"
	PKIThreatIDInformationDisclosure  = "THREAT-I-001"
	PKIThreatIDDenialOfService        = "THREAT-D-001"
	PKIThreatIDElevationOfPrivilege   = "THREAT-E-001"
	PKIThreatTypeCryptographicKey     = "cryptographic_key"
	PKIThreatStatusMitigated          = "mitigated"
	PKIThreatStatusImplemented        = "implemented"
	PKIControlTypeTechnical           = "technical"
	PKIThreatMitigationAuditLogging   = "Audit logging"
	PKIThreatMitigationAccessControls = "Access controls"
)

// PKI compliance requirement text and identifiers.
const (
	PKIComplianceTitleSerialNumber             = "Serial Number"
	PKIComplianceDescriptionCABasicConstraints = "CA certificates MUST have Basic Constraints"
	CABFBaselineSectionCertificateExtensions   = "7.1.2.4"
	RFC5280SectionSerialNumber                 = "4.1.2.2"
	RFC5280SectionValidity                     = "4.1.2.5"
	CABFBaselineIDBasicConstraintsExtension    = "BR-7.1.2.4-BC"
	PKIErrorConfigRequired                     = "config is required"
	PKIErrorIssuerCertificateRequired          = "issuer certificate is required"
	PKIErrorCryptoProviderRequired             = "crypto provider is required"
)

// File permission constants.
const (
	DirPermissions     = 0o755
	FilePermissions    = 0o644
	KeyFilePermissions = 0o600
)
