// Copyright (c) 2025 Justin Cranford
//
//

package intermediatekeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilJose "cryptoutil/internal/jose/crypto"
	cryptoutilRootKeysService "cryptoutil/internal/kms/server/barrier/rootkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
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

	service, ok := v.(*cryptoutilTelemetry.TelemetryService)
	if !ok {
		return nil
	}

	return service
}

func toJWKGenService(v any) *cryptoutilJose.JWKGenService {
	if v == nil {
		return nil
	}

	service, ok := v.(*cryptoutilJose.JWKGenService)
	if !ok {
		return nil
	}

	return service
}

func toOrmRepository(v any) *cryptoutilOrmRepository.OrmRepository {
	if v == nil {
		return nil
	}

	repo, ok := v.(*cryptoutilOrmRepository.OrmRepository)
	if !ok {
		return nil
	}

	return repo
}

func toRootKeysService(v any) *cryptoutilRootKeysService.RootKeysService {
	if v == nil {
		return nil
	}

	service, ok := v.(*cryptoutilRootKeysService.RootKeysService)
	if !ok {
		return nil
	}

	return service
}
