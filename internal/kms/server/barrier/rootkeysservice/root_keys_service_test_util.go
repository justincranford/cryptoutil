// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilJose "cryptoutil/internal/jose/crypto"
	cryptoutilUnsealKeysService "cryptoutil/internal/kms/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")

	return rootKeysService
}
