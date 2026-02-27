// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"fmt"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

func TestContentKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	service, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestNewRotationService_ValidationErrors tests constructor validation paths.
func TestNewRotationService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create valid dependencies for testing.
	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
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
		repository         cryptoutilAppsTemplateServiceServerBarrier.Repository
		unsealKeysService  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContain string
	}{
		{
			name:               "nil jwk gen service",
			jwkGenService:      nil,
			repository:         repo,
			unsealKeysService:  unsealSvc,
			expectedErrContain: "jwkGenService must be non-nil",
		},
		{
			name:               "nil repository",
			jwkGenService:      jwkGenSvc,
			repository:         nil,
			unsealKeysService:  unsealSvc,
			expectedErrContain: "repository must be non-nil",
		},
		{
			name:               "nil unseal keys service",
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			unsealKeysService:  nil,
			expectedErrContain: "unsealKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(
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

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	service, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, service)
}

// TestGormRepository_AddRootKey_NilKey tests AddRootKey with nil key.
func TestGormRepository_AddRootKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var rootKey *cryptoutilAppsTemplateServiceServerBarrier.RootKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var intermediateKey *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var contentKey *cryptoutilAppsTemplateServiceServerBarrier.ContentKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
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
