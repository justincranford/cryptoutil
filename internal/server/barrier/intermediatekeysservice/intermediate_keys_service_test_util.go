package intermediatekeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilRootKeysService "cryptoutil/internal/server/barrier/rootkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
)

func RequireNewForTest(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService) *IntermediateKeysService {
	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, jwkGenService, ormRepository, rootKeysService)
	cryptoutilAppErr.RequireNoError(err, "failed to create intermediateKeysService")

	return intermediateKeysService
}

// Helper functions for validation tests - allow passing nil values.
func toTelemetryService(v any) *cryptoutilTelemetry.TelemetryService {
	if v == nil {
		return nil
	}

	service, _ := v.(*cryptoutilTelemetry.TelemetryService)

	return service
}

func toJWKGenService(v any) *cryptoutilJose.JWKGenService {
	if v == nil {
		return nil
	}

	service, _ := v.(*cryptoutilJose.JWKGenService)

	return service
}

func toOrmRepository(v any) *cryptoutilOrmRepository.OrmRepository {
	if v == nil {
		return nil
	}

	repo, _ := v.(*cryptoutilOrmRepository.OrmRepository)

	return repo
}

func toRootKeysService(v any) *cryptoutilRootKeysService.RootKeysService {
	if v == nil {
		return nil
	}

	service, _ := v.(*cryptoutilRootKeysService.RootKeysService)

	return service
}
