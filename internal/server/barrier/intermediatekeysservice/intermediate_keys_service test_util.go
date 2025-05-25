package intermediatekeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService, uuidV7KeyGenPool *cryptoutilKeygen.KeyGenPool, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *IntermediateKeysService {
	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService, uuidV7KeyGenPool, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create intermediateKeysService")
	return intermediateKeysService
}
