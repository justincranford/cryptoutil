// Copyright (c) 2025 Justin Cranford

//go:build ignore
// +build ignore

// TODO(v7-phase5): This test file is temporarily disabled because it imports
// cryptoutil/internal/kms/server/repository/sqlrepository which no longer exists.
// This will be fixed during Phase 5 (KMS Barrier Migration) when shared/barrier
// is merged INTO the template barrier.

package intermediatekeysservice

import (
	"context"
	"testing"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

func TestNewIntermediateKeysService_ValidationErrors(t *testing.T) {
	tests := []struct {
		name              string
		telemetryService  any
		jwkGenService     any
		ormRepository     any
		rootKeysService   any
		expectedErrString string
	}{
		{
			name:              "nil telemetryService",
			telemetryService:  nil,
			jwkGenService:     testJWKGenService,
			ormRepository:     testOrmRepository,
			rootKeysService:   testRootKeysService,
			expectedErrString: "telemetryService must be non-nil",
		},
		{
			name:              "nil jwkGenService",
			telemetryService:  testTelemetryService,
			jwkGenService:     nil,
			ormRepository:     testOrmRepository,
			rootKeysService:   testRootKeysService,
			expectedErrString: "jwkGenService must be non-nil",
		},
		{
			name:              "nil ormRepository",
			telemetryService:  testTelemetryService,
			jwkGenService:     testJWKGenService,
			ormRepository:     nil,
			rootKeysService:   testRootKeysService,
			expectedErrString: "ormRepository must be non-nil",
		},
		{
			name:              "nil rootKeysService",
			telemetryService:  testTelemetryService,
			jwkGenService:     testJWKGenService,
			ormRepository:     testOrmRepository,
			rootKeysService:   nil,
			expectedErrString: "rootKeysService must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewIntermediateKeysService(
				toTelemetryService(tc.telemetryService),
				toJWKGenService(tc.jwkGenService),
				toOrmRepository(tc.ormRepository),
				toRootKeysService(tc.rootKeysService),
			)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tc.expectedErrString)
		})
	}
}

func TestIntermediateKeysService_EncryptKey_Success(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)

	defer intermediateKeysService.Shutdown()

	_, clearContentKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)
	require.NotNil(t, clearContentKey)

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedContentKeyBytes, intermediateKeyKidUUID, encErr := intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
		require.NoError(t, encErr)
		require.NotNil(t, encryptedContentKeyBytes)
		require.NotNil(t, intermediateKeyKidUUID)
		require.NotEmpty(t, encryptedContentKeyBytes)
		require.NotEqual(t, "", intermediateKeyKidUUID.String())

		return nil
	})
	require.NoError(t, err)
}

func TestIntermediateKeysService_DecryptKey_Success(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)

	defer intermediateKeysService.Shutdown()

	_, clearContentKey, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)
	require.NotNil(t, clearContentKey)

	var encryptedContentKeyBytes []byte

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var encErr error

		encryptedContentKeyBytes, _, encErr = intermediateKeysService.EncryptKey(sqlTransaction, clearContentKey)
		require.NoError(t, encErr)

		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedContentKeyBytes)

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentKey, decErr := intermediateKeysService.DecryptKey(sqlTransaction, encryptedContentKeyBytes)
		require.NoError(t, decErr)
		require.NotNil(t, decryptedContentKey)
		require.ElementsMatch(t, clearContentKey.Keys(), decryptedContentKey.Keys())

		return nil
	})
	require.NoError(t, err)
}

func TestIntermediateKeysService_DecryptKey_InvalidEncryptedData(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)

	defer intermediateKeysService.Shutdown()

	invalidEncryptedData := []byte("invalid-encrypted-content")

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentKey, decErr := intermediateKeysService.DecryptKey(sqlTransaction, invalidEncryptedData)
		require.Error(t, decErr)
		require.Nil(t, decryptedContentKey)
		require.Contains(t, decErr.Error(), "failed to parse encrypted content key message")

		return nil
	})
	require.NoError(t, err)
}

func TestIntermediateKeysService_Shutdown(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)
	require.NotNil(t, intermediateKeysService.telemetryService)
	require.NotNil(t, intermediateKeysService.jwkGenService)
	require.NotNil(t, intermediateKeysService.ormRepository)
	require.NotNil(t, intermediateKeysService.rootKeysService)

	intermediateKeysService.Shutdown()

	require.Nil(t, intermediateKeysService.telemetryService)
	require.Nil(t, intermediateKeysService.jwkGenService)
	require.Nil(t, intermediateKeysService.ormRepository)
	require.Nil(t, intermediateKeysService.rootKeysService)
}

func TestIntermediateKeysService_RoundTrip(t *testing.T) {
	intermediateKeysService, err := NewIntermediateKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testRootKeysService)
	require.NoError(t, err)
	require.NotNil(t, intermediateKeysService)

	defer intermediateKeysService.Shutdown()

	tests := []struct {
		name string
	}{
		{
			name: "A256GCM with dir",
		},
		{
			name: "A128GCM with dir",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var clearContentKey any

			var err error

			if tc.name == "A256GCM with dir" {
				_, clearContentKey, _, _, _, err = testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
			} else {
				_, clearContentKey, _, _, _, err = testJWKGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA128GCM, &cryptoutilSharedCryptoJose.AlgDir)
			}

			require.NoError(t, err)
			require.NotNil(t, clearContentKey)

			var encryptedContentKeyBytes []byte

			err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
				var encErr error

				jwkKey, ok := clearContentKey.(joseJwk.Key)
				require.True(t, ok)

				encryptedContentKeyBytes, _, encErr = intermediateKeysService.EncryptKey(sqlTransaction, jwkKey)
				require.NoError(t, encErr)

				return nil
			})
			require.NoError(t, err)
			require.NotNil(t, encryptedContentKeyBytes)

			err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
				decryptedContentKey, decErr := intermediateKeysService.DecryptKey(sqlTransaction, encryptedContentKeyBytes)
				require.NoError(t, decErr)
				require.NotNil(t, decryptedContentKey)

				jwkKey, ok := clearContentKey.(joseJwk.Key)
				require.True(t, ok)
				require.ElementsMatch(t, jwkKey.Keys(), decryptedContentKey.Keys())

				return nil
			})
			require.NoError(t, err)
		})
	}
}
