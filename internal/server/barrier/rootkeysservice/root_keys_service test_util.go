package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, uuidV7KeyGenPool *cryptoutilKeygen.KeyGenPool, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, ormRepository, unsealKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")
	return rootKeysService
}
