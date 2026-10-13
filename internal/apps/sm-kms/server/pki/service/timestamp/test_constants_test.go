// Copyright (c) 2025-2026 Justin Cranford.
package timestamp

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	testLiteralAddInfoNotAvailable    = cryptoutilSharedMagic.RFC3161FailureNameAddInfoNotAvailable
	testLiteralBadAlg                 = cryptoutilSharedMagic.RFC3161FailureNameBadAlg
	testLiteralBadDataFormat          = cryptoutilSharedMagic.RFC3161FailureNameBadDataFormat
	testLiteralBadRequest             = cryptoutilSharedMagic.RFC3161FailureNameBadRequest
	testLiteralGranted                = cryptoutilSharedMagic.RFC3161StatusNameGranted
	testLiteralGrantedWithMods        = cryptoutilSharedMagic.RFC3161StatusNameGrantedWithMods
	testLiteralRejection              = cryptoutilSharedMagic.RFC3161StatusNameRejection
	testLiteralRevocationNotification = cryptoutilSharedMagic.RFC3161StatusNameRevocationNotification
	testLiteralRevocationWarning      = cryptoutilSharedMagic.RFC3161StatusNameRevocationWarning
	testLiteralSystemFailure          = cryptoutilSharedMagic.RFC3161FailureNameSystemFailure
	testLiteralTimeNotAvailable       = cryptoutilSharedMagic.RFC3161FailureNameTimeNotAvailable
	testLiteralUnacceptedExtension    = cryptoutilSharedMagic.RFC3161FailureNameUnacceptedExtension
	testLiteralUnacceptedPolicy       = cryptoutilSharedMagic.RFC3161FailureNameUnacceptedPolicy
	testLiteralWaiting                = cryptoutilSharedMagic.StringWaiting
)
