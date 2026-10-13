// Copyright (c) 2025-2026 Justin Cranford.
package revocation

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	testLiteralTestOrg                     = "Test Org"
	testLiteralAaCompromise                = cryptoutilSharedMagic.RFC5280RevocationReasonNameAACompromise
	testLiteralAffiliationChanged          = cryptoutilSharedMagic.RFC5280RevocationReasonNameAffiliationChanged
	testLiteralCaCompromise                = cryptoutilSharedMagic.RFC5280RevocationReasonNameCACompromise
	testLiteralCertificateHold             = cryptoutilSharedMagic.RFC5280RevocationReasonNameCertificateHold
	testLiteralCessationOfOperation        = cryptoutilSharedMagic.RFC5280RevocationReasonNameCessationOfOperation
	testLiteralConfigIsRequired            = cryptoutilSharedMagic.PKIErrorConfigRequired
	testLiteralCryptoProviderIsRequired    = cryptoutilSharedMagic.PKIErrorCryptoProviderRequired
	testLiteralIssuerCertificateIsRequired = cryptoutilSharedMagic.PKIErrorIssuerCertificateRequired
	testLiteralKeyCompromise               = cryptoutilSharedMagic.RFC5280RevocationReasonNameKeyCompromise
	testLiteralNilConfig                   = "nil-config"
	testLiteralNilIssuer                   = "nil-issuer"
	testLiteralNilProvider                 = "nil-provider"
	testLiteralPrivilegeWithdrawn          = cryptoutilSharedMagic.RFC5280RevocationReasonNamePrivilegeWithdrawn
	testLiteralRemoveFromCRL               = cryptoutilSharedMagic.RFC5280RevocationReasonNameRemoveFromCRL
	testLiteralSuperseded                  = cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded
	testLiteralTestExampleCom              = "test.example.com"
	testLiteralUnspecified                 = cryptoutilSharedMagic.RFC5280RevocationReasonNameUnspecified
	testLiteralValidConfig                 = "valid-config"
)
