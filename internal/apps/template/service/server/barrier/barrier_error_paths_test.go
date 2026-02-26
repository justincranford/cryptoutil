// Copyright (c) 2025 Justin Cranford
//

package barrier

import (
"context"
"database/sql"
"fmt"
"testing"

cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
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

// createBarrierNoTablesDB creates an in-memory SQLite DB with NO barrier tables.
func createBarrierNoTablesDB(t *testing.T) *gorm.DB {
t.Helper()

dbUUID, err := googleUuid.NewV7()
require.NoError(t, err)

dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())

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

telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
require.NoError(t, err)

t.Cleanup(func() { telemetryService.Shutdown() })

jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
require.NoError(t, err)

t.Cleanup(func() { jwkGenService.Shutdown() })

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

// TestRotateRootKey_GenerateJWKFails verifies the error path when rotationGenerateRootJWEJWKFn fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateRootKey_GenerateJWKFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationGenerateRootJWEJWKFn
rotationGenerateRootJWEJWKFn = func(_ *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
return nil, nil, nil, nil, nil, errMockServiceFailure
}

defer func() { rotationGenerateRootJWEJWKFn = originalFn }()

mockRepo := newMockServiceRepository()
mockRepo.tx.rootKey = &RootKey{UUID: googleUuid.New(), Encrypted: "dummy", KEKUUID: googleUuid.UUID{}}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateRootKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to generate root JWK")
}

// TestRotateRootKey_UnsealEncryptFails verifies the error path when rotationUnsealEncryptKeyFn fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateRootKey_UnsealEncryptFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationUnsealEncryptKeyFn
rotationUnsealEncryptKeyFn = func(_ cryptoutilUnsealKeysService.UnsealKeysService, _ joseJwk.Key) ([]byte, error) {
return nil, errMockServiceFailure
}

defer func() { rotationUnsealEncryptKeyFn = originalFn }()

mockRepo := newMockServiceRepository()
mockRepo.tx.rootKey = &RootKey{UUID: googleUuid.New(), Encrypted: "dummy", KEKUUID: googleUuid.UUID{}}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateRootKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to encrypt root key")
}

// TestRotateIntermediateKey_GenerateJWKFails verifies the error path when rotationGenerateIntermediateJWEJWKFn fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateIntermediateKey_GenerateJWKFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationGenerateIntermediateJWEJWKFn
rotationGenerateIntermediateJWEJWKFn = func(_ *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
return nil, nil, nil, nil, nil, errMockServiceFailure
}

defer func() { rotationGenerateIntermediateJWEJWKFn = originalFn }()

rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
require.NoError(t, err)

mockRepo := newMockServiceRepository()
mockRepo.tx.intermediateKey = &IntermediateKey{UUID: googleUuid.New(), Encrypted: "dummy", KEKUUID: googleUuid.New()}
mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateIntermediateKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to generate intermediate JWK")
}

// TestRotateIntermediateKey_EncryptFails verifies the error path when rotationEncryptKeyFn fails
// during intermediate key rotation. Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateIntermediateKey_EncryptFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationEncryptKeyFn
rotationEncryptKeyFn = func(_ []joseJwk.Key, _ joseJwk.Key) (*joseJwe.Message, []byte, error) {
return nil, nil, errMockServiceFailure
}

defer func() { rotationEncryptKeyFn = originalFn }()

rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
require.NoError(t, err)

mockRepo := newMockServiceRepository()
mockRepo.tx.intermediateKey = &IntermediateKey{UUID: googleUuid.New(), Encrypted: "dummy", KEKUUID: googleUuid.New()}
mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateIntermediateKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to encrypt intermediate key")
}

// TestRotateContentKey_GenerateJWKFails verifies the error path when rotationGenerateContentJWEJWKFn fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateContentKey_GenerateJWKFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationGenerateContentJWEJWKFn
rotationGenerateContentJWEJWKFn = func(_ *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
return nil, nil, nil, nil, nil, errMockServiceFailure
}

defer func() { rotationGenerateContentJWEJWKFn = originalFn }()

rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
require.NoError(t, err)

_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
require.NoError(t, err)

mockRepo := newMockServiceRepository()
mockRepo.tx.intermediateKey = &IntermediateKey{UUID: googleUuid.New(), Encrypted: string(encryptedIntermediateKey), KEKUUID: *rootKeyUUID}
mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateContentKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to generate content JWK")
}

// TestRotateContentKey_EncryptFails verifies the error path when rotationEncryptKeyFn fails
// during content key rotation. Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestRotateContentKey_EncryptFails(t *testing.T) {
jwkGenService, unsealService := setupRotationServiceTestHelper(t)

originalFn := rotationEncryptKeyFn
rotationEncryptKeyFn = func(_ []joseJwk.Key, _ joseJwk.Key) (*joseJwe.Message, []byte, error) {
return nil, nil, errMockServiceFailure
}

defer func() { rotationEncryptKeyFn = originalFn }()

rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

_, clearIntermediateJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
require.NoError(t, err)

encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
require.NoError(t, err)

_, encryptedIntermediateKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{clearRootJWK}, clearIntermediateJWK)
require.NoError(t, err)

mockRepo := newMockServiceRepository()
mockRepo.tx.intermediateKey = &IntermediateKey{UUID: googleUuid.New(), Encrypted: string(encryptedIntermediateKey), KEKUUID: *rootKeyUUID}
mockRepo.tx.rootKey = &RootKey{UUID: *rootKeyUUID, Encrypted: string(encryptedRootKey)}

rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
require.NoError(t, err)

result, err := rotationService.RotateContentKey(context.Background(), "test")
require.Error(t, err)
require.Nil(t, result)
require.Contains(t, err.Error(), "failed to encrypt content key")
}

// TestIntermediateKeyInit_GenerateJWKFails verifies the error path when intermediateGenerateJWEJWKFn
// fails during intermediate key service init. The mock repo returns ErrNoIntermediateKeyFound so that
// the code attempts to generate the first intermediate key, which then fails.
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestIntermediateKeyInit_GenerateJWKFails(t *testing.T) {
telemetryService, jwkGenService, unsealService := setupBarrierErrorTestHelper(t)

originalFn := intermediateGenerateJWEJWKFn
intermediateGenerateJWEJWKFn = func(_ *cryptoutilSharedCryptoJose.JWKGenService) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
return nil, nil, nil, nil, nil, errMockServiceFailure
}

defer func() { intermediateGenerateJWEJWKFn = originalFn }()

mockRepo := newMockServiceRepository()
mockRepo.tx.getIntermediateKeyLatestErr = ErrNoIntermediateKeyFound

_, err := NewService(context.Background(), telemetryService, jwkGenService, mockRepo, unsealService)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to create intermediate keys service")
}
