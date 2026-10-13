// Copyright (c) 2025-2026 Justin Cranford.
//

package barrier

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	rotationTestReason       = "test rotation"
	testCaseRotationSucceeds = "rotation succeeds"
)

// createBarrierNoTablesDB creates an in-memory SQLite DB with NO barrier tables.
func createBarrierNoTablesDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbUUID.String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	db, err := gorm.Open(gormSqlite.Dialector{Conn: sqlDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	t.Cleanup(func() { _ = sqlDB.Close() })

	return db
}

// createBarrierRootOnlyDB creates an in-memory SQLite DB with only the barrier_root_keys table.
func createBarrierRootOnlyDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := createBarrierNoTablesDB(t)

	err := db.AutoMigrate(&RootKey{})
	require.NoError(t, err)

	return db
}

// mockFailingUnsealService implements UnsealKeysService where EncryptKey always fails.
type mockFailingUnsealService struct{}

func (m *mockFailingUnsealService) EncryptKey(_ joseJwk.Key) ([]byte, error) {
	return nil, errMockServiceFailure
}

func (m *mockFailingUnsealService) DecryptKey(_ []byte) (joseJwk.Key, error) {
	return nil, errMockServiceFailure
}

func (m *mockFailingUnsealService) EncryptData(_ []byte) ([]byte, error) {
	return nil, errMockServiceFailure
}

func (m *mockFailingUnsealService) DecryptData(_ []byte) ([]byte, error) {
	return nil, errMockServiceFailure
}

func (m *mockFailingUnsealService) Shutdown() {}

// setupBarrierErrorTestHelper creates telemetry, jwkGen, and unseal service for barrier error path tests.
func setupBarrierErrorTestHelper(t *testing.T) (*cryptoutilSharedTelemetry.TelemetryService, *cryptoutilSharedCryptoJose.JWKGenService, cryptoutilUnsealKeysService.UnsealKeysService) {
	t.Helper()

	ctx := context.Background()

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)

	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := newBarrierTestJWKGenService(ctx, telemetryService)
	require.NoError(t, err)

	t.Cleanup(func() { shutdownBarrierTestJWKGenService(jwkGenService) })

	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	t.Cleanup(func() { unsealService.Shutdown() })

	return telemetryService, jwkGenService, unsealService
}

// TestNewBarrierService_RootKeyInitFails verifies that NewService fails gracefully when
// GetRootKeyLatest returns a SQL error (no barrier_root_keys table). Covers the GORM repository
// SQL error path, root_keys_service init failure paths, and barrier_service root keys error path.
func TestNewBarrierService_RootKeyInitFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	db := createBarrierNoTablesDB(t)
	gormRepo, err := NewGormRepository(db)
	require.NoError(t, err)

	_, err = NewService(ctx, telemetryService, jwkGenService, gormRepo, unsealService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create root keys service")
}

// TestNewBarrierService_IntermediateKeyInitFails verifies that NewService fails gracefully when
// GetIntermediateKeyLatest returns a SQL error (only root table exists, no intermediate table).
// Covers the GORM repository SQL error intermediate path and intermediate keys service cascade error.
func TestNewBarrierService_IntermediateKeyInitFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	db := createBarrierRootOnlyDB(t)
	gormRepo, err := NewGormRepository(db)
	require.NoError(t, err)

	_, err = NewService(ctx, telemetryService, jwkGenService, gormRepo, unsealService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create intermediate keys service")
}

// TestRootKeysService_UnsealEncryptFails verifies that initializeFirstRootJWK fails when
// the unseal service EncryptKey call fails. Uses a DB with the root keys table (empty) so that
// root key creation is attempted, and a mock unseal service that always fails EncryptKey.
func TestRootKeysService_UnsealEncryptFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	telemetryService, jwkGenService, _ := setupBarrierErrorTestHelper(t)

	db := createBarrierRootOnlyDB(t)
	gormRepo, err := NewGormRepository(db)
	require.NoError(t, err)

	_, err = NewService(ctx, telemetryService, jwkGenService, gormRepo, &mockFailingUnsealService{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create root keys service")
}

func TestRotateRootKey_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*RotationService)
		wantErr   string
	}{
		{name: testCaseRotationSucceeds},
		{
			name: "root JWK generation error",
			configure: func(service *RotationService) {
				service.generateRootJWEJWKFn = func(*cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
					return nil, nil, nil, nil, nil, errMockServiceFailure
				}
			},
			wantErr: "failed to generate root JWK",
		},
		{
			name: "unseal encryption error",
			configure: func(service *RotationService) {
				service.encryptUnsealKeyFn = func(cryptoutilUnsealKeysService.UnsealKeysService, joseJwk.Key) ([]byte, error) {
					return nil, errMockServiceFailure
				}
			},
			wantErr: "failed to encrypt root key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwkGenService, unsealService := setupRotationServiceTestHelper(t)
			oldRootKeyUUID, err := googleUuid.NewV7()
			require.NoError(t, err)

			mockRepo := newMockServiceRepository()
			mockRepo.tx.rootKey = &RootKey{UUID: oldRootKeyUUID, Encrypted: testBarrierEncryptedValue}

			rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
			require.NoError(t, err)

			if tc.configure != nil {
				tc.configure(rotationService)
			}

			result, err := rotationService.RotateRootKey(context.Background(), rotationTestReason)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, oldRootKeyUUID, result.OldKeyUUID)
			require.NotEqual(t, oldRootKeyUUID, result.NewKeyUUID)
		})
	}
}

func TestRotateIntermediateKey_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*RotationService)
		wantErr   string
	}{
		{name: testCaseRotationSucceeds},
		{
			name: "intermediate JWK generation error",
			configure: func(service *RotationService) {
				service.generateIntermediateJWEJWKFn = func(*cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
					return nil, nil, nil, nil, nil, errMockServiceFailure
				}
			},
			wantErr: "failed to generate intermediate JWK",
		},
		{
			name: "intermediate encryption error",
			configure: func(service *RotationService) {
				service.encryptKeyFn = func([]joseJwk.Key, joseJwk.Key) (*joseJwe.Message, []byte, error) {
					return nil, nil, errMockServiceFailure
				}
			},
			wantErr: "failed to encrypt intermediate key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwkGenService, unsealService := setupRotationServiceTestHelper(t)
			rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
			require.NoError(t, err)
			encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
			require.NoError(t, err)

			oldIntermediateUUID, err := googleUuid.NewV7()
			require.NoError(t, err)

			mockRepo := newMockServiceRepository()
			mockRepo.tx.intermediateKey = &IntermediateKey{UUID: oldIntermediateUUID, Encrypted: testBarrierEncryptedValue, KEKUUID: *rootKeyUUID}
			mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

			rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
			require.NoError(t, err)

			if tc.configure != nil {
				tc.configure(rotationService)
			}

			result, err := rotationService.RotateIntermediateKey(context.Background(), rotationTestReason)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, oldIntermediateUUID, result.OldKeyUUID)
			require.NotEqual(t, oldIntermediateUUID, result.NewKeyUUID)
		})
	}
}

func TestRotateContentKey_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*RotationService)
		wantErr   string
	}{
		{name: testCaseRotationSucceeds},
		{
			name: "content JWK generation error",
			configure: func(service *RotationService) {
				service.generateContentJWEJWKFn = func(*cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
					return nil, nil, nil, nil, nil, errMockServiceFailure
				}
			},
			wantErr: "failed to generate content JWK",
		},
		{
			name: "content encryption error",
			configure: func(service *RotationService) {
				service.encryptKeyFn = func([]joseJwk.Key, joseJwk.Key) (*joseJwe.Message, []byte, error) {
					return nil, nil, errMockServiceFailure
				}
			},
			wantErr: "failed to encrypt content key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jwkGenService, unsealService := setupRotationServiceTestHelper(t)
			rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
			require.NoError(t, err)
			intermediateUUID, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
			require.NoError(t, err)
			encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
			require.NoError(t, err)
			_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
			require.NoError(t, err)

			mockRepo := newMockServiceRepository()
			mockRepo.tx.intermediateKey = &IntermediateKey{UUID: *intermediateUUID, Encrypted: string(encryptedIntermediateKey), KEKUUID: *rootKeyUUID}
			mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

			rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
			require.NoError(t, err)

			if tc.configure != nil {
				tc.configure(rotationService)
			}

			result, err := rotationService.RotateContentKey(context.Background(), rotationTestReason)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotEqual(t, googleUuid.UUID{}, result.NewKeyUUID)
		})
	}
}

// TestIntermediateKeyInit_GenerateJWKFails verifies the error path when generateJWEJWKFn
// fails during intermediate key service init. The mock repo returns ErrNoIntermediateKeyFound so that
// the code attempts to generate the first intermediate key, which then fails.
func TestIntermediateKeyInit_GenerateJWKFails(t *testing.T) {
	t.Parallel()

	telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

	stubGenerateJWEJWK := func(_ *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
		return nil, nil, nil, nil, nil, errMockServiceFailure
	}

	mockRepo := newMockServiceRepository()
	mockRepo.tx.getIntermediateKeyLatestErr = ErrNoIntermediateKeyFound

	_, err := newServiceInternal(context.Background(), telemetryService, jwkGenService, mockRepo, unsealService, stubGenerateJWEJWK)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create intermediate keys service")
}
