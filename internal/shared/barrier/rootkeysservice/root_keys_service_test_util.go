// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// RequireNewForTest creates a new RootKeysService for testing and panics on error.
func RequireNewForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create rootKeysService")

	return rootKeysService
}
