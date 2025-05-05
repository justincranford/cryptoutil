package intermediatekeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *IntermediateKeysService {
	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create intermediateKeysService")
	return intermediateKeysService
}
