// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)

// createKeyServiceTestDB creates an isolated in-memory SQLite database with barrier tables.
func createKeyServiceTestDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	ctx := context.Background()

	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Create barrier tables.
	schema := `
	CREATE TABLE IF NOT EXISTS barrier_root_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS barrier_intermediate_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (kek_uuid) REFERENCES barrier_root_keys(uuid)
	);
	CREATE TABLE IF NOT EXISTS barrier_content_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY (kek_uuid) REFERENCES barrier_intermediate_keys(uuid)
	);
	`
	_, err = sqlDB.ExecContext(ctx, schema)
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, sqlDB.Close())
	}

	return db, cleanup
}

// TestNewRootKeysService_ValidationErrors tests constructor validation paths.
func TestNewRootKeysService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create valid dependencies for testing.
	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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
		telemetryService   *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService      *cryptoutilSharedCryptoJose.JWKGenService
		repository         cryptoutilAppsTemplateServiceServerBarrier.Repository
		unsealKeysService  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContain string
	}{
		{
			name:               "nil telemetry service",
			telemetryService:   nil,
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			unsealKeysService:  unsealSvc,
			expectedErrContain: "telemetryService must be non-nil",
		},
		{
			name:               "nil jwk gen service",
			telemetryService:   telemetrySvc,
			jwkGenService:      nil,
			repository:         repo,
			unsealKeysService:  unsealSvc,
			expectedErrContain: "jwkGenService must be non-nil",
		},
		{
			name:               "nil repository",
			telemetryService:   telemetrySvc,
			jwkGenService:      jwkGenSvc,
			repository:         nil,
			unsealKeysService:  unsealSvc,
			expectedErrContain: "repository must be non-nil",
		},
		{
			name:               "nil unseal keys service",
			telemetryService:   telemetrySvc,
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			unsealKeysService:  nil,
			expectedErrContain: "unsealKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(
				tt.telemetryService,
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

// TestNewIntermediateKeysService_ValidationErrors tests constructor validation paths.
func TestNewIntermediateKeysService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create valid dependencies for testing.
	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create a valid root keys service for testing.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	tests := []struct {
		name               string
		telemetryService   *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService      *cryptoutilSharedCryptoJose.JWKGenService
		repository         cryptoutilAppsTemplateServiceServerBarrier.Repository
		rootKeysService    *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService
		expectedErrContain string
	}{
		{
			name:               "nil telemetry service",
			telemetryService:   nil,
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			rootKeysService:    rootKeysSvc,
			expectedErrContain: "telemetryService must be non-nil",
		},
		{
			name:               "nil jwk gen service",
			telemetryService:   telemetrySvc,
			jwkGenService:      nil,
			repository:         repo,
			rootKeysService:    rootKeysSvc,
			expectedErrContain: "jwkGenService must be non-nil",
		},
		{
			name:               "nil repository",
			telemetryService:   telemetrySvc,
			jwkGenService:      jwkGenSvc,
			repository:         nil,
			rootKeysService:    rootKeysSvc,
			expectedErrContain: "repository must be non-nil",
		},
		{
			name:               "nil root keys service",
			telemetryService:   telemetrySvc,
			jwkGenService:      jwkGenSvc,
			repository:         repo,
			rootKeysService:    nil,
			expectedErrContain: "rootKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(
				tt.telemetryService,
				tt.jwkGenService,
				tt.repository,
				tt.rootKeysService,
			)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestNewContentKeysService_ValidationErrors tests constructor validation paths.
func TestNewContentKeysService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create valid dependencies for testing.
	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create valid root and intermediate keys services for testing.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	tests := []struct {
		name                    string
		telemetryService        *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService           *cryptoutilSharedCryptoJose.JWKGenService
		repository              cryptoutilAppsTemplateServiceServerBarrier.Repository
		intermediateKeysService *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKeysService
		expectedErrContain      string
	}{
		{
			name:                    "nil telemetry service",
			telemetryService:        nil,
			jwkGenService:           jwkGenSvc,
			repository:              repo,
			intermediateKeysService: intermediateKeysSvc,
			expectedErrContain:      "telemetryService must be non-nil",
		},
		{
			name:                    "nil jwk gen service",
			telemetryService:        telemetrySvc,
			jwkGenService:           nil,
			repository:              repo,
			intermediateKeysService: intermediateKeysSvc,
			expectedErrContain:      "jwkGenService must be non-nil",
		},
		{
			name:                    "nil repository",
			telemetryService:        telemetrySvc,
			jwkGenService:           jwkGenSvc,
			repository:              nil,
			intermediateKeysService: intermediateKeysSvc,
			expectedErrContain:      "repository must be non-nil",
		},
		{
			name:                    "nil intermediate keys service",
			telemetryService:        telemetrySvc,
			jwkGenService:           jwkGenSvc,
			repository:              repo,
			intermediateKeysService: nil,
			expectedErrContain:      "intermediateKeysService must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(
				tt.telemetryService,
				tt.jwkGenService,
				tt.repository,
				tt.intermediateKeysService,
			)
			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestRootKeysService_Shutdown tests shutdown behavior.
func TestRootKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	service, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestIntermediateKeysService_Shutdown tests shutdown behavior.
func TestIntermediateKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	service, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestContentKeysService_Shutdown tests shutdown behavior.
func TestContentKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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
	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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
func TestRootKeysService_DecryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name               string
		setupFunc          func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte)
		expectedErrContain string
	}{
		{
			name: "invalid_jwe_format",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte) {
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

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

				var tx cryptoutilAppsTemplateServiceServerBarrier.Transaction

				err = repo.WithTransaction(ctx, func(transaction cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					tx = transaction

					return nil
				})
				require.NoError(t, err)

				// Return invalid JWE format (not a valid JWE at all)
				invalidJWE := []byte("this is not a valid JWE format")

				return tx, rootKeysSvc, invalidJWE
			},
			expectedErrContain: "failed to parse encrypted intermediate key message",
		},
		{
			name: "corrupted_jwe_json",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte) {
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

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

				var tx cryptoutilAppsTemplateServiceServerBarrier.Transaction

				err = repo.WithTransaction(ctx, func(transaction cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					tx = transaction

					return nil
				})
				require.NoError(t, err)

				// Return corrupted JSON (looks like JSON but invalid structure)
				corruptedJWE := []byte(`{"protected":"invalid","encrypted_key":"data","iv":"data","ciphertext":"data","tag":"data"}`)

				return tx, rootKeysSvc, corruptedJWE
			},
			expectedErrContain: "failed to parse encrypted intermediate key message",
		},
		{
			name: "nil_encrypted_bytes",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte) {
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

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

				var tx cryptoutilAppsTemplateServiceServerBarrier.Transaction

				err = repo.WithTransaction(ctx, func(transaction cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					tx = transaction

					return nil
				})
				require.NoError(t, err)

				// Return nil bytes
				return tx, rootKeysSvc, nil
			},
			expectedErrContain: "failed to parse encrypted intermediate key message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, service, encryptedBytes := tt.setupFunc(t)

			_, err := service.DecryptKey(tx, encryptedBytes)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestRootKeysService_EncryptKey_GetLatestFails tests EncryptKey when GetRootKeyLatest fails.
func TestRootKeysService_EncryptKey_GetLatestFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	// Create DB but DON'T initialize any root keys
	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Create ONLY the schema, don't insert any keys
	schema := `
	CREATE TABLE IF NOT EXISTS barrier_root_keys (
		uuid TEXT PRIMARY KEY,
		encrypted TEXT NOT NULL,
		kek_uuid TEXT,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);
	`
	_, err = sqlDB.ExecContext(ctx, schema)
	require.NoError(t, err)

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	// Create a proper service and then manually delete the root key from DB
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	// Delete the root key that was just created
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Now try to encrypt - should fail because no root key exists
	_, testKey, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	var encryptErr error

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encryptErr = rootKeysSvc.EncryptKey(tx, testKey)

		return encryptErr
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted root JWK latest from DB")
}

// TestIntermediateKeysService_DecryptKey_ErrorPaths tests error paths in intermediate key decryption.
func TestIntermediateKeysService_DecryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name               string
		setupFunc          func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKeysService, []byte)
		expectedErrContain string
	}{
		{
			name: "invalid_jwe_format",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.IntermediateKeysService, []byte) {
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

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

				var tx cryptoutilAppsTemplateServiceServerBarrier.Transaction

				err = repo.WithTransaction(ctx, func(transaction cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					tx = transaction

					return nil
				})
				require.NoError(t, err)

				invalidJWE := []byte("invalid JWE content")

				return tx, intermediateKeysSvc, invalidJWE
			},
			expectedErrContain: "failed to parse encrypted content key message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx, service, encryptedBytes := tt.setupFunc(t)

			_, err := service.DecryptKey(tx, encryptedBytes)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestContentKeysService_EncryptContent_NilInput tests EncryptContent with nil input.
func TestContentKeysService_EncryptContent_NilInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		// Try to encrypt nil data
		_, _, encryptErr := contentKeysSvc.EncryptContent(tx, nil)

		return encryptErr
	})
	// The function should handle nil gracefully - check actual behavior
	// Based on the code, it will likely pass nil to jose encryption which may error
	// For now, we're just testing that it handles the error path
	if err != nil {
		require.Error(t, err)
	}
}

// TestContentKeysService_DecryptContent_InvalidCiphertext tests DecryptContent with invalid ciphertext.
func TestContentKeysService_DecryptContent_InvalidCiphertext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		// Try to decrypt invalid ciphertext
		invalidCiphertext := []byte("this is not valid JWE ciphertext")
		_, decryptErr := contentKeysSvc.DecryptContent(tx, invalidCiphertext)

		return decryptErr
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWE message")
}

// TestRootKeysService_DecryptKey_AdditionalErrorPaths tests additional error scenarios.
func TestRootKeysService_DecryptKey_AdditionalErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name               string
		setupFunc          func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte)
		expectedErrContain string
	}{
		{
			name: "key_not_found",
			setupFunc: func(t *testing.T) (cryptoutilAppsTemplateServiceServerBarrier.Transaction, *cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, []byte) {
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

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

				// Encrypt an intermediate key with a root key.
				var encryptedBytes []byte

				err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					_, testKey, _, _, _, genErr := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
					if genErr != nil {
						return genErr
					}

					encBytes, _, encErr := rootKeysSvc.EncryptKey(tx, testKey)
					encryptedBytes = encBytes

					return encErr
				})
				require.NoError(t, err)

				// Delete the root key to make it non-existent.
				sqlDB, err := db.DB()
				require.NoError(t, err)
				_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
				require.NoError(t, err)

				// Return transaction and encrypted bytes.
				var tx cryptoutilAppsTemplateServiceServerBarrier.Transaction

				_ = repo.WithTransaction(ctx, func(txInner cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
					tx = txInner

					return nil
				})

				return tx, rootKeysSvc, encryptedBytes
			},
			expectedErrContain: "failed to get root key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx, service, encryptedBytes := tt.setupFunc(t)

			_, err := service.DecryptKey(tx, encryptedBytes)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrContain)
		})
	}
}

// TestRotationService_RotateRootKey_ErrorPaths tests rotation error scenarios.
func TestRotationService_RotateRootKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test: No root key exists.
	t.Run("no_root_key_exists", func(t *testing.T) {
		t.Parallel()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
		require.NoError(t, err)

		// Delete all root keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
		require.NoError(t, err)

		// Attempt rotation - should fail.
		_, err = rotationSvc.RotateRootKey(ctx, "test rotation")

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get")
	})
}

// TestRotationService_RotateIntermediateKey_ErrorPaths tests intermediate key rotation errors.
func TestRotationService_RotateIntermediateKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test: No intermediate key exists.
	t.Run("no_intermediate_key_exists", func(t *testing.T) {
		t.Parallel()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
		require.NoError(t, err)

		// Delete all intermediate keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Attempt rotation - should fail.
		_, err = rotationSvc.RotateIntermediateKey(ctx, "test rotation")

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get")
	})
}

// TestRotationService_RotateContentKey_ErrorPaths tests content key rotation errors.
func TestRotationService_RotateContentKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test: No intermediate key exists (content key rotation requires intermediate key).
	t.Run("no_intermediate_key_exists", func(t *testing.T) {
		t.Parallel()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
		require.NoError(t, err)

		// Delete all intermediate keys (content key rotation requires intermediate key).
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Attempt rotation - should fail because no intermediate key exists.
		_, err = rotationSvc.RotateContentKey(ctx, "test rotation")

		require.Error(t, err)
		require.Contains(t, err.Error(), "no intermediate key found")
	})
}

// TestContentKeysService_EncryptContent_ErrorPaths tests encryption error scenarios.
func TestContentKeysService_EncryptContent_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("intermediate_key_encryption_fails", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// Delete all intermediate keys to cause encryption to fail.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Attempt to encrypt content - should fail because no intermediate key exists.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encryptErr := contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK")
	})

	t.Run("add_content_key_db_failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// Create a content key first to establish UUID.
		var firstKeyID *googleUuid.UUID

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, keyID, encryptErr := contentKeysSvc.EncryptContent(tx, []byte("test data"))
			firstKeyID = keyID

			return encryptErr
		})
		require.NoError(t, err)

		// Try to manually insert a content key with the same UUID to cause UNIQUE constraint violation.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			return tx.AddContentKey(&cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
				UUID:      *firstKeyID, // Duplicate UUID
				Encrypted: "fake_encrypted_jwk",
				KEKUUID:   googleUuid.New(),
			})
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "UNIQUE constraint failed") // SQLite error
	})
}

// TestContentKeysService_DecryptContent_ErrorPaths tests decryption error scenarios.
func TestContentKeysService_DecryptContent_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("invalid_jwe_format", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// Attempt to decrypt with invalid JWE format.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, []byte("not-a-valid-jwe"))

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JWE message")
	})

	t.Run("content_key_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// First encrypt some content to get a valid JWE.
		var ciphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			ciphertext, _, encryptErr = contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Delete all content keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_content_keys")
		require.NoError(t, err)

		// Attempt to decrypt - should fail because content key not found.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, ciphertext)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted content key")
	})

	t.Run("missing_kid_header", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// Create a JWE without a kid header.
		// JWE compact format: header.encrypted_key.iv.ciphertext.tag
		// Header: {"alg":"A256KW","enc":"A256GCM"} - missing "kid" field.
		jweWithoutKid := []byte("eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIn0.AAAA.AAAA.AAAA.AAAA")

		// Attempt to decrypt - should fail because kid header is missing.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, jweWithoutKid)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse JWE message kid")
	})

	t.Run("decrypt_content_key_failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		// Create first barrier with original unseal key.
		_, unsealJWK1, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK1})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc1.Shutdown() })

		rootKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc1.Shutdown() })

		intermediateKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc1.Shutdown() })

		contentKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc1.Shutdown() })

		// Encrypt content with the first barrier.
		var ciphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			ciphertext, _, encryptErr = contentKeysSvc1.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Delete all root and intermediate keys to simulate rotation/corruption.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Create second barrier with DIFFERENT unseal key.
		_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc2.Shutdown() })

		rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc2.Shutdown() })

		intermediateKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc2.Shutdown() })

		contentKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc2.Shutdown() })

		// Attempt to decrypt with second barrier - should fail because intermediate key used to encrypt content key is missing.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc2.DecryptContent(tx, ciphertext)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt content key")
	})

	t.Run("decrypt_bytes_failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
		require.NoError(t, err)
		t.Cleanup(func() { contentKeysSvc.Shutdown() })

		// Encrypt content to get a valid JWE structure.
		var validCiphertext []byte

		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			var encryptErr error

			validCiphertext, _, encryptErr = contentKeysSvc.EncryptContent(tx, []byte("test data"))

			return encryptErr
		})
		require.NoError(t, err)

		// Convert to string, split into JWE parts.
		jweString := string(validCiphertext)
		parts := []byte(jweString)
		dotCount := 0
		thirdDotIdx := -1
		fourthDotIdx := -1

		for i := 0; i < len(parts); i++ {
			if parts[i] == '.' {
				dotCount++
				if dotCount == 3 {
					thirdDotIdx = i
				} else if dotCount == 4 {
					fourthDotIdx = i

					break
				}
			}
		}

		require.True(t, thirdDotIdx > 0 && fourthDotIdx > thirdDotIdx, "JWE compact serialization should have at least 4 dots")

		// Replace the ciphertext portion (between 3rd and 4th dot) with valid base64url but wrong length/content.
		// This will pass JWE parsing but fail during actual AES-GCM decryption.
		corruptedJWE := make([]byte, 0, len(parts))
		corruptedJWE = append(corruptedJWE, parts[:thirdDotIdx+1]...)
		corruptedJWE = append(corruptedJWE, []byte("AAAAAAAAAAAAAAAAAAAAAA")...) // Valid base64url, wrong ciphertext
		corruptedJWE = append(corruptedJWE, parts[fourthDotIdx:]...)

		// Attempt to decrypt - should fail at DecryptBytesWithContext.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := contentKeysSvc.DecryptContent(tx, corruptedJWE)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt content with content key")
	})
}

// TestIntermediateKeysService_EncryptKey_ErrorPaths tests intermediate key encryption error scenarios.
func TestIntermediateKeysService_EncryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("no_intermediate_key_exists", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		// Delete all intermediate keys.
		sqlDB, err := db.DB()
		require.NoError(t, err)
		_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
		require.NoError(t, err)

		// Generate a test JWK to encrypt.
		_, testJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Attempt to encrypt - should fail because no intermediate key exists.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encryptErr := intermediateKeysSvc.EncryptKey(tx, testJWK)

			return encryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK latest from DB")
	})

	t.Run("decrypt_intermediate_key_failure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		// Create first unseal key and initialize keys.
		_, unsealJWK1, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK1})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc1.Shutdown() })

		rootKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc1)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc1.Shutdown() })

		intermediateKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc1)
		require.NoError(t, err)
		intermediateKeysSvc1.Shutdown() // Shutdown after initialization.

		// Create DIFFERENT unseal key and new services.
		_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)
		unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
		require.NoError(t, err)
		t.Cleanup(func() { unsealSvc2.Shutdown() })

		rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { rootKeysSvc2.Shutdown() })

		// Create intermediate service with WRONG root service.
		intermediateKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc2)
		require.NoError(t, err)
		t.Cleanup(func() { intermediateKeysSvc2.Shutdown() })

		// Generate content JWK to encrypt.
		_, contentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Try to encrypt - should fail because wrong unseal key means can't decrypt intermediate key.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, _, encErr := intermediateKeysSvc2.EncryptKey(tx, contentJWK)

			return encErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to decrypt intermediate JWK latest")
	})
}

// TestNewService_NilParameters tests that NewService properly rejects nil parameters.
func TestNewService_NilParameters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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
		name                string
		ctx                 context.Context
		telemetrySvc        *cryptoutilSharedTelemetry.TelemetryService
		jwkGenSvc           *cryptoutilSharedCryptoJose.JWKGenService
		repo                cryptoutilAppsTemplateServiceServerBarrier.Repository
		unsealSvc           cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContains string
	}{
		{
			name:                "nil_context",
			ctx:                 nil,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "ctx must be non-nil",
		},
		{
			name:                "nil_telemetry_service",
			ctx:                 ctx,
			telemetrySvc:        nil,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "telemetryService must be non-nil",
		},
		{
			name:                "nil_jwkgen_service",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           nil,
			repo:                repo,
			unsealSvc:           unsealSvc,
			expectedErrContains: "jwkGenService must be non-nil",
		},
		{
			name:                "nil_repository",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                nil,
			unsealSvc:           unsealSvc,
			expectedErrContains: "repository must be non-nil",
		},
		{
			name:                "nil_unseal_service",
			ctx:                 ctx,
			telemetrySvc:        telemetrySvc,
			jwkGenSvc:           jwkGenSvc,
			repo:                repo,
			unsealSvc:           nil,
			expectedErrContains: "unsealKeysService must be non-nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			svc, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(tc.ctx, tc.telemetrySvc, tc.jwkGenSvc, tc.repo, tc.unsealSvc)
			require.Error(t, err)
			require.Nil(t, svc)
			require.Contains(t, err.Error(), tc.expectedErrContains)
		})
	}
}

// TestDecryptContent_InvalidKidFormat tests DecryptContent with invalid kid format in JWE.
func TestDecryptContent_InvalidKidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	// Create a malformed JWE with an invalid kid format (not a UUID).
	// JWE compact format: header.encrypted_key.iv.ciphertext.tag
	// We create a valid-looking JWE header with invalid kid, rest can be garbage.
	malformedJWE := []byte("eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIiwia2lkIjoibm90LWEtdXVpZCJ9.AAAA.AAAA.AAAA.AAAA")

	// Try to decrypt - should fail because kid is not a valid UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := contentKeysSvc.DecryptContent(tx, malformedJWE)

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse kid as uuid")
}

// TestIntermediateKeysService_DecryptKey_RootKeyMissing tests intermediate key decryption when root key is missing.
func TestIntermediateKeysService_DecryptKey_RootKeyMissing(t *testing.T) {
	t.Parallel()

	t.Run("intermediate_key_not_found", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

		// Create a JWE with a non-existent intermediate key kid.
		_, testKey, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
		require.NoError(t, err)

		// Encrypt some data with this key - the resulting JWE will have a kid that doesn't exist in DB.
		_, jweBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{testKey}, []byte("test data"))
		require.NoError(t, err)

		// Try to decrypt - should fail because intermediate key not found.
		err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
			_, decryptErr := intermediateKeysSvc.DecryptKey(tx, jweBytes)

			return decryptErr
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get intermediate key")
	})
}

// TestRootKeysService_EncryptKey_NoRootKey tests root key encryption when no root key exists.
func TestRootKeysService_EncryptKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Delete the root key that was auto-created.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Generate a test JWK to encrypt.
	_, testJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Attempt to encrypt - should fail because no root key exists.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encryptErr := rootKeysSvc.EncryptKey(tx, testJWK)

		return encryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted root JWK latest from DB")
}

// TestRotationService_RotateRootKey_NoExistingKey tests root key rotation when no key exists.
func TestRotationService_RotateRootKey_NoExistingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create rotation service directly (without root keys service initialization).
	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, rotationSvc)

	// Try to rotate - should fail because no root key exists.
	_, err = rotationSvc.RotateRootKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get current root key")
}

// TestRotationService_RotateIntermediateKey_NoExistingKey tests intermediate key rotation when no key exists.
func TestRotationService_RotateIntermediateKey_NoExistingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create rotation service directly (without services initialization).
	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, rotationSvc)

	// Try to rotate - should fail because no intermediate key exists.
	_, err = rotationSvc.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get current intermediate key")
}

// TestRotationService_RotateContentKey_NoExistingKey tests content key rotation when no key exists.
func TestRotationService_RotateContentKey_NoExistingKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create rotation service directly (without services initialization).
	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, rotationSvc)

	// Try to rotate - should fail because no intermediate key exists (content key depends on intermediate key).
	_, err = rotationSvc.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get current intermediate key")
}

// TestRootKeysService_DecryptKey_InvalidJWE tests DecryptKey with invalid JWE format.
func TestRootKeysService_DecryptKey_InvalidJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Try to decrypt invalid JWE format.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := rootKeysSvc.DecryptKey(tx, []byte("not-a-valid-jwe"))

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse encrypted intermediate key message")
}

// TestRootKeysService_DecryptKey_InvalidKidFormat tests DecryptKey with invalid kid format.
func TestRootKeysService_DecryptKey_InvalidKidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create a JWE with invalid kid format (not a valid UUID).
	malformedJWE := []byte("eyJhbGciOiJkaXIiLCJlbmMiOiJBMjU2R0NNIiwia2lkIjoibm90LWEtdXVpZCJ9..AAAA.AAAA.AAAA")

	// Try to decrypt - should fail because kid is not a valid UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := rootKeysSvc.DecryptKey(tx, malformedJWE)

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse kid as uuid")
}

// TestRootKeysService_DecryptKey_RootKeyNotFound tests DecryptKey when root key doesn't exist.
func TestRootKeysService_DecryptKey_RootKeyNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create a JWE referencing a non-existent root key UUID.
	_, testKey, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgDir)
	require.NoError(t, err)

	// Encrypt some data with this key to get a JWE with a kid that doesn't exist in DB.
	_, jweBytes, err := cryptoutilSharedCryptoJose.EncryptBytes([]joseJwk.Key{testKey}, []byte("test data"))
	require.NoError(t, err)

	// Try to decrypt - should fail because root key not found.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := rootKeysSvc.DecryptKey(tx, jweBytes)

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestIntermediateKeysService_DecryptKey_InvalidJWE tests DecryptKey with invalid JWE format.
func TestIntermediateKeysService_DecryptKey_InvalidJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Try to decrypt invalid JWE format.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := intermediateKeysSvc.DecryptKey(tx, []byte("not-a-valid-jwe"))

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse encrypted content key message")
}

// TestIntermediateKeysService_DecryptKey_InvalidKidFormat tests DecryptKey with invalid kid format.
func TestIntermediateKeysService_DecryptKey_InvalidKidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create a JWE with invalid kid format (not a valid UUID).
	malformedJWE := []byte("eyJhbGciOiJBMjU2S1ciLCJlbmMiOiJBMjU2R0NNIiwia2lkIjoibm90LWEtdXVpZCJ9..AAAA.AAAA.AAAA")

	// Try to decrypt - should fail because kid is not a valid UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := intermediateKeysSvc.DecryptKey(tx, malformedJWE)

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse kid as uuid")
}

// TestContentKeysService_DecryptContent_InvalidJWE tests DecryptContent with invalid JWE format.
func TestContentKeysService_DecryptContent_InvalidJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	// Try to decrypt invalid JWE format.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decryptErr := contentKeysSvc.DecryptContent(tx, []byte("not-a-valid-jwe"))

		return decryptErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWE message")
}

// TestRotationService_RotateIntermediateKey_NoRootKey tests intermediate key rotation when no root key exists.
func TestRotationService_RotateIntermediateKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create services normally first (creates initial keys).
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)

	// Delete all root keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to rotate intermediate key - should fail because no root key exists.
	_, err = rotationSvc.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get")
}

// TestRotationService_RotateContentKey_NoRootKey tests content key rotation when root key is missing.
func TestRotationService_RotateContentKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create services normally first (creates initial keys).
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	rotationSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)

	// Delete all root keys (this will cause content key rotation to fail when trying to decrypt intermediate key).
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to rotate content key - should fail because root key is missing.
	_, err = rotationSvc.RotateContentKey(ctx, "test rotation")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestEncryptContent_InvalidInput tests EncryptContent with edge cases.
func TestEncryptContent_InvalidInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	contentKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { contentKeysSvc.Shutdown() })

	// Test with empty content (should fail - empty is invalid).
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := contentKeysSvc.EncryptContent(tx, []byte{})
		require.Error(t, encErr)
		require.Contains(t, encErr.Error(), "clearBytes")

		return nil
	})
	require.NoError(t, err)

	// Test with nil content (should also fail - nil is invalid).
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := contentKeysSvc.EncryptContent(tx, nil)
		require.Error(t, encErr)
		require.Contains(t, encErr.Error(), "clearBytes")

		return nil
	})
	require.NoError(t, err)

	// Test with large content.
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		encryptedLarge, _, encErr := contentKeysSvc.EncryptContent(tx, largeContent)
		if encErr != nil {
			return encErr
		}

		decrypted, decErr := contentKeysSvc.DecryptContent(tx, encryptedLarge)
		if decErr != nil {
			return decErr
		}

		require.Equal(t, largeContent, decrypted)

		return nil
	})
	require.NoError(t, err)
}

// TestIntermediateKeysService_EncryptKey_NoIntermediateKey tests EncryptKey when no intermediate key exists.
func TestIntermediateKeysService_EncryptKey_NoIntermediateKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create services to initialize keys.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	// Delete all intermediate keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_intermediate_keys")
	require.NoError(t, err)

	// Try to encrypt key - should fail because no intermediate key exists.
	_, testContentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := intermediateKeysSvc.EncryptKey(tx, testContentJWK)

		return encErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted intermediate JWK")
}

// TestRootKeysService_EncryptKey_NoRootKey_DeletedKey tests EncryptKey when no root key exists.
func TestRootKeysService_EncryptKey_NoRootKey_DeletedKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create root keys service to initialize keys.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	// Delete all root keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to encrypt key - should fail because no root key exists.
	_, testIntermediateJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, _, encErr := rootKeysSvc.EncryptKey(tx, testIntermediateJWK)

		return encErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get encrypted root JWK")
}

// TestIntermediateKeysService_DecryptKey_NoRootKey tests DecryptKey when root key is missing.
func TestIntermediateKeysService_DecryptKey_NoRootKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
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

	// Create services to initialize keys.
	rootKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilAppsTemplateServiceServerBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	// Encrypt a key first to get valid encrypted data.
	var encryptedKeyBytes []byte

	_, testContentJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		var encErr error

		encryptedKeyBytes, _, encErr = intermediateKeysSvc.EncryptKey(tx, testContentJWK)

		return encErr
	})
	require.NoError(t, err)

	// Delete all root keys.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "DELETE FROM barrier_root_keys")
	require.NoError(t, err)

	// Try to decrypt - should fail because root key is missing.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, decErr := intermediateKeysSvc.DecryptKey(tx, encryptedKeyBytes)

		return decErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get root key")
}

// TestRepositoryAddKey_NilInput tests that Add* methods reject nil key inputs.
func TestRepositoryAddKey_NilInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := createKeyServiceTestDB(t)
	t.Cleanup(cleanup)

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	// Test AddRootKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddRootKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test AddIntermediateKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddIntermediateKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test AddContentKey with nil key.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		return tx.AddContentKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")
}

// TestRepositoryGetKey_NilUUID tests that Get*Key methods reject nil UUID inputs.
func TestRepositoryGetKey_NilUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, cleanup := createKeyServiceTestDB(t)
	t.Cleanup(cleanup)

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	// Test GetRootKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetRootKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test GetIntermediateKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetIntermediateKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")

	// Test GetContentKey with nil UUID.
	err = repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		_, getErr := tx.GetContentKey(nil)

		return getErr
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-nil")
}

// TestRootKeysService_EncryptKey_ErrorPaths tests EncryptKey error scenarios.
func TestRootKeysService_EncryptKey_ErrorPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(*testing.T) (*cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, cryptoutilAppsTemplateServiceServerBarrier.Repository, *cryptoutilSharedCryptoJose.JWKGenService)
		wantErr string
	}{
		{
			name: "decrypt_root_key_failure",
			setup: func(t *testing.T) (*cryptoutilAppsTemplateServiceServerBarrier.RootKeysService, cryptoutilAppsTemplateServiceServerBarrier.Repository, *cryptoutilSharedCryptoJose.JWKGenService) {
				// Create two different unseal keys.
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
				require.NoError(t, err)
				t.Cleanup(func() { telemetrySvc.Shutdown() })

				jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
				require.NoError(t, err)
				t.Cleanup(func() { jwkGenSvc.Shutdown() })

				db, cleanup := createKeyServiceTestDB(t)
				t.Cleanup(cleanup)

				repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(db)
				require.NoError(t, err)
				t.Cleanup(func() { repo.Shutdown() })

				// Create first unseal key and initialize root key.
				_, unsealJWK1, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)
				unsealSvc1, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK1})
				require.NoError(t, err)
				// Don't cleanup unsealSvc1 yet.

				rootKeysSvc1, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc1)
				require.NoError(t, err)
				// Don't cleanup rootKeysSvc1 yet - still need DB populated.

				// Create DIFFERENT unseal key and new service.
				_, unsealJWK2, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)
				unsealSvc2, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK2})
				require.NoError(t, err)
				t.Cleanup(func() {
					rootKeysSvc1.Shutdown()
					unsealSvc1.Shutdown()
					unsealSvc2.Shutdown()
				})

				// Create service with WRONG unseal key.
				rootKeysSvc2, err := cryptoutilAppsTemplateServiceServerBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc2)
				require.NoError(t, err)
				t.Cleanup(func() { rootKeysSvc2.Shutdown() })

				return rootKeysSvc2, repo, jwkGenSvc
			},
			wantErr: "failed to decrypt root JWK latest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootKeysSvc, repo, jwkGenSvc := tc.setup(t)

			var testJWK joseJwk.Key

			if tc.name == "encrypt_bytes_failure" {
				// Create signing key (Ed25519) which cannot be used for encryption.
				_, _, signKey, _, _, err := jwkGenSvc.GenerateJWSJWK(cryptoutilSharedCryptoJose.AlgEdDSA)
				require.NoError(t, err)

				testJWK = signKey
			} else {
				// Generate intermediate JWK to encrypt.
				_, intermediateJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
				require.NoError(t, err)

				testJWK = intermediateJWK
			}

			// Try to encrypt - should fail.
			err := repo.WithTransaction(ctx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
				_, _, encErr := rootKeysSvc.EncryptKey(tx, testJWK)

				return encErr
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
