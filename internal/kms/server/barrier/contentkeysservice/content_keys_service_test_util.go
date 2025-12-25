// Copyright (c) 2025 Justin Cranford

package contentkeysservice

import (
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilIntermediateKeysService "cryptoutil/internal/kms/server/barrier/intermediatekeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
)

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
