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

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
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
	sqlDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
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
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	tests := []struct {
		name               string
		telemetryService   *cryptoutilTelemetry.TelemetryService
		jwkGenService      *cryptoutilJose.JWKGenService
		repository         cryptoutilTemplateBarrier.BarrierRepository
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

			service, err := cryptoutilTemplateBarrier.NewRootKeysService(
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
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	// Create a valid root keys service for testing.
	rootKeysSvc, err := cryptoutilTemplateBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	tests := []struct {
		name               string
		telemetryService   *cryptoutilTelemetry.TelemetryService
		jwkGenService      *cryptoutilJose.JWKGenService
		repository         cryptoutilTemplateBarrier.BarrierRepository
		rootKeysService    *cryptoutilTemplateBarrier.RootKeysService
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

			service, err := cryptoutilTemplateBarrier.NewIntermediateKeysService(
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
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	// Create valid root and intermediate keys services for testing.
	rootKeysSvc, err := cryptoutilTemplateBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilTemplateBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	tests := []struct {
		name                    string
		telemetryService        *cryptoutilTelemetry.TelemetryService
		jwkGenService           *cryptoutilJose.JWKGenService
		repository              cryptoutilTemplateBarrier.BarrierRepository
		intermediateKeysService *cryptoutilTemplateBarrier.IntermediateKeysService
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

			service, err := cryptoutilTemplateBarrier.NewContentKeysService(
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

	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	service, err := cryptoutilTemplateBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestIntermediateKeysService_Shutdown tests shutdown behavior.
func TestIntermediateKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilTemplateBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	service, err := cryptoutilTemplateBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)

	// Shutdown should not panic and can be called multiple times.
	service.Shutdown()
	service.Shutdown()
}

// TestContentKeysService_Shutdown tests shutdown behavior.
func TestContentKeysService_Shutdown(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	rootKeysSvc, err := cryptoutilTemplateBarrier.NewRootKeysService(telemetrySvc, jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	t.Cleanup(func() { rootKeysSvc.Shutdown() })

	intermediateKeysSvc, err := cryptoutilTemplateBarrier.NewIntermediateKeysService(telemetrySvc, jwkGenSvc, repo, rootKeysSvc)
	require.NoError(t, err)
	t.Cleanup(func() { intermediateKeysSvc.Shutdown() })

	service, err := cryptoutilTemplateBarrier.NewContentKeysService(telemetrySvc, jwkGenSvc, repo, intermediateKeysSvc)
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
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	tests := []struct {
		name               string
		jwkGenService      *cryptoutilJose.JWKGenService
		repository         cryptoutilTemplateBarrier.BarrierRepository
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

			service, err := cryptoutilTemplateBarrier.NewRotationService(
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

	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)
	unsealSvc, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealSvc.Shutdown() })

	service, err := cryptoutilTemplateBarrier.NewRotationService(jwkGenSvc, repo, unsealSvc)
	require.NoError(t, err)
	require.NotNil(t, service)
}

// TestGormBarrierRepository_AddRootKey_NilKey tests AddRootKey with nil key.
func TestGormBarrierRepository_AddRootKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddRootKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormBarrierRepository_AddIntermediateKey_NilKey tests AddIntermediateKey with nil key.
func TestGormBarrierRepository_AddIntermediateKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddIntermediateKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormBarrierRepository_AddContentKey_NilKey tests AddContentKey with nil key.
func TestGormBarrierRepository_AddContentKey_NilKey(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
		return tx.AddContentKey(nil)
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "key must be non-nil")
}

// TestGormBarrierRepository_GetRootKey_NilUUID tests GetRootKey with nil UUID.
func TestGormBarrierRepository_GetRootKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var rootKey *cryptoutilTemplateBarrier.BarrierRootKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
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

// TestGormBarrierRepository_GetIntermediateKey_NilUUID tests GetIntermediateKey with nil UUID.
func TestGormBarrierRepository_GetIntermediateKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var intermediateKey *cryptoutilTemplateBarrier.BarrierIntermediateKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
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

// TestGormBarrierRepository_GetContentKey_NilUUID tests GetContentKey with nil UUID.
func TestGormBarrierRepository_GetContentKey_NilUUID(t *testing.T) {
	t.Parallel()

	db, cleanup := createKeyServiceTestDB(t)
	defer cleanup()

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Shutdown() })

	var contentKey *cryptoutilTemplateBarrier.BarrierContentKey

	err = repo.WithTransaction(context.Background(), func(tx cryptoutilTemplateBarrier.BarrierTransaction) error {
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
