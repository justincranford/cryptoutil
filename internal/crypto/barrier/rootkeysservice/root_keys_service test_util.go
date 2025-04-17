package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealRepository cryptoutilUnsealRepository.UnsealRepository) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, ormRepository, unsealRepository)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")
	return rootKeysService
}
