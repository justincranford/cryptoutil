// Copyright (c) 2025 Justin Cranford

//go:build ignore
// +build ignore

// TODO(v7-phase5): This test file is temporarily disabled because it imports
// cryptoutil/internal/kms/server/repository/sqlrepository which no longer exists.
// This will be fixed during Phase 5 (KMS Barrier Migration) when shared/barrier
// is merged INTO the template barrier.

package contentkeysservice

import (
	"context"
	"testing"

	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"

	"github.com/stretchr/testify/require"
)

func TestNewContentKeysService_ValidationErrors(t *testing.T) {
	tests := []struct {
		name                    string
		telemetryService        any
		jwkGenService           any
		ormRepository           any
		intermediateKeysService any
		expectedErrString       string
	}{
		{
			name:                    "nil telemetryService",
			telemetryService:        nil,
			jwkGenService:           testJWKGenService,
			ormRepository:           testOrmRepository,
			intermediateKeysService: testIntermediateKeysService,
			expectedErrString:       "telemetryService must be non-nil",
		},
		{
			name:                    "nil jwkGenService",
			telemetryService:        testTelemetryService,
			jwkGenService:           nil,
			ormRepository:           testOrmRepository,
			intermediateKeysService: testIntermediateKeysService,
			expectedErrString:       "jwkGenService must be non-nil",
		},
		{
			name:                    "nil ormRepository",
			telemetryService:        testTelemetryService,
			jwkGenService:           testJWKGenService,
			ormRepository:           nil,
			intermediateKeysService: testIntermediateKeysService,
			expectedErrString:       "ormRepository must be non-nil",
		},
		{
			name:                    "nil intermediateKeysService",
			telemetryService:        testTelemetryService,
			jwkGenService:           testJWKGenService,
			ormRepository:           testOrmRepository,
			intermediateKeysService: nil,
			expectedErrString:       "intermediateKeysService must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service, err := NewContentKeysService(
				toTelemetryService(tc.telemetryService),
				toJWKGenService(tc.jwkGenService),
				toOrmRepository(tc.ormRepository),
				toIntermediateKeysService(tc.intermediateKeysService),
			)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tc.expectedErrString)
		})
	}
}

func TestContentKeysService_EncryptContent_Success(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)

	defer contentKeysService.Shutdown()

	clearBytes := []byte("Test data for encryption")

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		encryptedContentBytes, contentKeyKidUUID, encErr := contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		require.NoError(t, encErr)
		require.NotNil(t, encryptedContentBytes)
		require.NotNil(t, contentKeyKidUUID)
		require.NotEmpty(t, encryptedContentBytes)
		require.NotEqual(t, "", contentKeyKidUUID.String())

		return nil
	})
	require.NoError(t, err)
}

func TestContentKeysService_DecryptContent_Success(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)

	defer contentKeysService.Shutdown()

	clearBytes := []byte("Test data for round-trip encryption")

	var encryptedContentBytes []byte

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var encErr error

		encryptedContentBytes, _, encErr = contentKeysService.EncryptContent(sqlTransaction, clearBytes)
		require.NoError(t, encErr)

		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, encryptedContentBytes)

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentBytes, decErr := contentKeysService.DecryptContent(sqlTransaction, encryptedContentBytes)
		require.NoError(t, decErr)
		require.NotNil(t, decryptedContentBytes)
		require.Equal(t, clearBytes, decryptedContentBytes)

		return nil
	})
	require.NoError(t, err)
}

func TestContentKeysService_DecryptContent_InvalidEncryptedData(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)

	defer contentKeysService.Shutdown()

	invalidEncryptedData := []byte("invalid-encrypted-content")

	err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		decryptedContentBytes, decErr := contentKeysService.DecryptContent(sqlTransaction, invalidEncryptedData)
		require.Error(t, decErr)
		require.Nil(t, decryptedContentBytes)
		require.Contains(t, decErr.Error(), "failed to parse JWE message")

		return nil
	})
	require.NoError(t, err)
}

func TestContentKeysService_Shutdown(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)
	require.NotNil(t, contentKeysService.telemetryService)
	require.NotNil(t, contentKeysService.jwkGenService)
	require.NotNil(t, contentKeysService.ormRepository)
	require.NotNil(t, contentKeysService.intermediateKeysService)

	contentKeysService.Shutdown()

	require.Nil(t, contentKeysService.telemetryService)
	// jwkGenService is NOT set to nil by Shutdown() - it's a shared resource
	require.Nil(t, contentKeysService.ormRepository)
	require.Nil(t, contentKeysService.intermediateKeysService)
}

func TestContentKeysService_RoundTrip(t *testing.T) {
	contentKeysService, err := NewContentKeysService(testTelemetryService, testJWKGenService, testOrmRepository, testIntermediateKeysService)
	require.NoError(t, err)
	require.NotNil(t, contentKeysService)

	defer contentKeysService.Shutdown()

	tests := []struct {
		name  string
		bytes []byte
	}{
		{
			name:  "Small data",
			bytes: []byte("Hello World"),
		},
		{
			name:  "Medium data",
			bytes: []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."),
		},
		{
			name:  "Binary data",
			bytes: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var encryptedContentBytes []byte

			err := testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
				var encErr error

				encryptedContentBytes, _, encErr = contentKeysService.EncryptContent(sqlTransaction, tc.bytes)
				require.NoError(t, encErr)

				return nil
			})
			require.NoError(t, err)
			require.NotNil(t, encryptedContentBytes)

			err = testOrmRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
				decryptedContentBytes, decErr := contentKeysService.DecryptContent(sqlTransaction, encryptedContentBytes)
				require.NoError(t, decErr)
				require.NotNil(t, decryptedContentBytes)
				require.Equal(t, tc.bytes, decryptedContentBytes)

				return nil
			})
			require.NoError(t, err)
		})
	}
}
