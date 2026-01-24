// Copyright (c) 2025 Justin Cranford
//
//

package intermediatekeysservice

import (
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilRootKeysService "cryptoutil/internal/shared/barrier/rootkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// RequireNewForTest creates a new IntermediateKeysService for testing and panics on error.
func RequireNewForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeysService *cryptoutilRootKeysService.RootKeysService) *IntermediateKeysService {
	intermediateKeysService, err := NewIntermediateKeysService(telemetryService, jwkGenService, ormRepository, rootKeysService)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create intermediateKeysService")

	return intermediateKeysService
}

// Helper functions for validation tests - allow passing nil values.
func toTelemetryService(v any) *cryptoutilSharedTelemetry.TelemetryService {
	if v == nil {
		return nil
	}

	service, ok := v.(*cryptoutilSharedTelemetry.TelemetryService)
	if !ok {
		return nil
	}

	return service
}

func toJWKGenService(v any) *cryptoutilSharedCryptoJose.JWKGenService {
	if v == nil {
		return nil
	}

	service, ok := v.(*cryptoutilSharedCryptoJose.JWKGenService)
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
