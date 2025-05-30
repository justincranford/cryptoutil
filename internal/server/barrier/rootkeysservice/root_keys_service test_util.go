package rootkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, uuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID], aes256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, ormRepository, unsealKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create rootKeysService")
	return rootKeysService
}
