// Copyright (c) 2025-2026 Justin Cranford.
//

package barrier_test

import (
	"context"
	"fmt"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

func TestContentKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	telemetrySvc, err := newBarrierTestTelemetryService(t)
	require.NoError(t, err)

	jwkGenSvc, err := newBarrierTestJWKGenService(t, telemetrySvc)
	require.NoError(t, err)

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilAppsFrameworkServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsFrameworkServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	service, err := cryptoutilAppsFrameworkServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestNewRotationService_ValidationErrors tests constructor validation paths.
func TestNewRotationService_ValidationErrors(t *testing.T) {
	t.Parallel()

	// Create valid dependencies for testing.
	jwkGenSvc, err := newBarrierTestJWKGenServiceWithTelemetry(t)
	require.NoError(t, err)

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	tests := []struct {
		name               string
		jwkGenService      *cryptoutilSharedCryptoJose.JWKGenService
		repository         cryptoutilAppsFrameworkServiceServerBarrier.Repository
		unsealKeysService  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContain string
	}{
		{
			name:               testCaseNilJWKGenService,
			jwkGenService:      nil,
			repository:         repo,
			unsealKeysService:  unsealSvc,
			expectedErrContain: testErrorJWKGenServiceMustBeNonNil,
		},
		{
			name:               testCaseNilRepository,
			jwkGenService:      jwkGenSvc,
			repository:         nil,
			unsealKeysService:  unsealSvc,
			expectedErrContain: testErrorRepositoryMustBeNonNil,
		},
		{
			name:               testCaseNilUnsealKeysService,
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			unsealKeysService:  nil,
			expectedErrContain: testErrorUnsealKeysServiceMustBeNonNil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService(
				tt.jwkGenService,
				tt.repository,
				tt.unsealKeysService,
			)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestNewRotationService_Success tests successful construction.
func TestNewRotationService_Success(t *testing.T) {
	t.Parallel()

	jwkGenSvc, err := newBarrierTestJWKGenServiceWithTelemetry(t)
	require.NoError(t, err)

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	service, err := cryptoutilAppsFrameworkServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, service)
}

// TestGormRepository_AddRootKey_NilKey tests AddRootKey with nil key.
func TestGormRepository_AddRootKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormRepository_AddIntermediateKey_NilKey tests AddIntermediateKey with nil key.
func TestGormRepository_AddIntermediateKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormRepository_AddContentKey_NilKey tests AddContentKey with nil key.
func TestGormRepository_AddContentKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormRepository_GetRootKey_NilUUID tests GetRootKey with nil UUID.
func TestGormRepository_GetRootKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var rootKey *cryptoutilAppsFrameworkServiceServerBarrier.RootKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		var getErr error

		rootKey, getErr = tx.GetRootKey(nil)
		if getErr != nil {
			return fmt.Errorf("GetRootKey error: %w", getErr)
		}

		return nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
	require.Nil(t, rootKey)
}

// TestGormRepository_GetIntermediateKey_NilUUID tests GetIntermediateKey with nil UUID.
func TestGormRepository_GetIntermediateKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var intermediateKey *cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		var getErr error

		intermediateKey, getErr = tx.GetIntermediateKey(nil)
		if getErr != nil {
			return fmt.Errorf("GetIntermediateKey error: %w", getErr)
		}

		return nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
	require.Nil(t, intermediateKey)
}

// TestGormRepository_GetContentKey_NilUUID tests GetContentKey with nil UUID.
func TestGormRepository_GetContentKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var contentKey *cryptoutilAppsFrameworkServiceServerBarrier.ContentKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsFrameworkServiceServerBarrier.Transaction) error {
		var getErr error

		contentKey, getErr = tx.GetContentKey(nil)
		if getErr != nil {
			return fmt.Errorf("GetContentKey error: %w", getErr)
		}

		return nil
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "uuid must be non-nil")
	require.Nil(t, contentKey)
}

// TestRootKeysService_DecryptKey_ErrorPaths tests error paths in root key decryption.
