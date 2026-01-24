// Copyright (c) 2025 Justin Cranford

package contentkeysservice

import (
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilIntermediateKeysService "cryptoutil/internal/shared/barrier/intermediatekeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

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

func toIntermediateKeysService(v any) *cryptoutilIntermediateKeysService.IntermediateKeysService {
	if v == nil {
		return nil
	}

	service, ok := v.(*cryptoutilIntermediateKeysService.IntermediateKeysService)
	if !ok {
		return nil
	}

	return service
}
