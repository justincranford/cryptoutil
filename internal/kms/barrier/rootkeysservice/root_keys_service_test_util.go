// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilJose "cryptoutil/internal/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/kms/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/kms/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")

	return rootKeysService
}
