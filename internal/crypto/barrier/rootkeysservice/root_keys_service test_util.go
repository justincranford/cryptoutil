package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/crypto/barrier/unsealkeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, ormRepository, unsealKeysService, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")
	return rootKeysService
}
