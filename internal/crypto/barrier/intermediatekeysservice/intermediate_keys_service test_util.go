package intermediatekeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilRootKeysService "cryptoutil/internal/crypto/barrier/rootkeysservice"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) *IntermediateKeysService {
	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, ormRepository, rootKeysService, aes256KeyGenPool)
	cryptoutilAppErr.RequireNoError(err, "failed to create intermediateKeysService")
	return intermediateKeysService
}
