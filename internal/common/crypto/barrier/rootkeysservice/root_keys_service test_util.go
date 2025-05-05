package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilUnsealKeysService "cryptoutil/internal/common/crypto/barrier/unsealkeysservice"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, ormRepository, unsealKeysService, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")
	return rootKeysService
}
