// Copyright (c) 2025 Justin Cranford

//go:build ignore
// +build ignore

// TODO(v7-phase5): This test utility file is temporarily disabled because it's
// only used by disabled comprehensive tests. This will be fixed during Phase 5
// (KMS Barrier Migration) when shared/barrier is merged INTO the template barrier.

package rootkeysservice

import (
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// RequireNewForTest creates a new RootKeysService for testing and panics on error.
func RequireNewForTest(telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService) *RootKeysService {
	rootKeysService, err := NewRootKeysService(telemetryService, jwkGenService, ormRepository, unsealKeysService)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create rootKeysService")

	return rootKeysService
}
