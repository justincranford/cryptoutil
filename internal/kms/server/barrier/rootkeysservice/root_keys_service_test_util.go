// Copyright (c) 2025 Justin Cranford
//
//

package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilJose "cryptoutil/internal/jose"
	cryptoutilUnsealKeysService "cryptoutil/internal/kms/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")

	return rootKeysService
}
