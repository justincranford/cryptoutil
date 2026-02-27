// Copyright (c) 2025 Justin Cranford
//

package barrier

import (
	"context"
	"errors"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
)

// errMockServiceFailure is a standard error for mock service failures in rotation service tests.
var errMockServiceFailure = errors.New("mock service failure")

// mockServiceTransaction implements Transaction interface for rotation service testing.
type mockServiceTransaction struct {
	ctx                                context.Context
	rootKey                            *RootKey
	intermediateKey                    *IntermediateKey
	contentKey                         *ContentKey
	getRootKeyLatestErr                error
	getRootKeyLatestReturnsNil         bool
	getRootKeyErr                      error
	addRootKeyErr                      error
	getIntermediateKeyLatestErr        error
	getIntermediateKeyLatestReturnsNil bool
	getIntermediateKeyErr              error
	addIntermediateKeyErr              error
	getContentKeyErr                   error
	addContentKeyErr                   error
}

func (m *mockServiceTransaction) Context() context.Context {
	return m.ctx
}

func (m *mockServiceTransaction) GetRootKeyLatest() (*RootKey, error) {
	if m.getRootKeyLatestErr != nil {
		return nil, m.getRootKeyLatestErr
	}

	if m.getRootKeyLatestReturnsNil {
		return nil, nil
	}

	return m.rootKey, nil
}

func (m *mockServiceTransaction) GetRootKey(_ *googleUuid.UUID) (*RootKey, error) {
	if m.getRootKeyErr != nil {
		return nil, m.getRootKeyErr
	}

	return m.rootKey, nil
}

func (m *mockServiceTransaction) AddRootKey(_ *RootKey) error {
	return m.addRootKeyErr
}

func (m *mockServiceTransaction) GetIntermediateKeyLatest() (*IntermediateKey, error) {
	if m.getIntermediateKeyLatestErr != nil {
		return nil, m.getIntermediateKeyLatestErr
	}

	if m.getIntermediateKeyLatestReturnsNil {
		return nil, nil
	}

	return m.intermediateKey, nil
}

func (m *mockServiceTransaction) GetIntermediateKey(_ *googleUuid.UUID) (*IntermediateKey, error) {
	if m.getIntermediateKeyErr != nil {
		return nil, m.getIntermediateKeyErr
	}

	return m.intermediateKey, nil
}

func (m *mockServiceTransaction) AddIntermediateKey(_ *IntermediateKey) error {
	return m.addIntermediateKeyErr
}

func (m *mockServiceTransaction) GetContentKey(_ *googleUuid.UUID) (*ContentKey, error) {
	if m.getContentKeyErr != nil {
		return nil, m.getContentKeyErr
	}

	return m.contentKey, nil
}

func (m *mockServiceTransaction) AddContentKey(_ *ContentKey) error {
	return m.addContentKeyErr
}

// mockServiceRepository implements Repository interface for rotation service testing.
type mockServiceRepository struct {
	tx             *mockServiceTransaction
	withTxErr      error
	shouldCallTxFn bool
	shutdownCalled bool
}

func (m *mockServiceRepository) WithTransaction(ctx context.Context, fn func(tx Transaction) error) error {
	if m.withTxErr != nil {
		return m.withTxErr
	}

	if m.shouldCallTxFn && m.tx != nil {
		m.tx.ctx = ctx

		return fn(m.tx)
	}

	return nil
}

func (m *mockServiceRepository) Shutdown() {
	m.shutdownCalled = true
}

func newMockServiceRepository() *mockServiceRepository {
	return &mockServiceRepository{
		tx:             &mockServiceTransaction{},
		shouldCallTxFn: true,
	}
}

// setupRotationServiceTestHelper creates test dependencies for rotation service testing.
func setupRotationServiceTestHelper(t *testing.T) (*cryptoutilSharedCryptoJose.JWKGenService, cryptoutilUnsealKeysService.UnsealKeysService) {
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

	return jwkGenService, unsealService
}

// TestRotateRootKey_NoRootKeyFound tests error when no root key exists.
func TestRotateRootKey_NoRootKeyFound(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.getRootKeyLatestReturnsNil = true

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateRootKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no root key found")
}

// TestRotateRootKey_AddRootKeyFailure tests error when adding new root key fails.
func TestRotateRootKey_AddRootKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.rootKey = &RootKey{
		UUID:      googleUuid.New(),
		Encrypted: "dummy-encrypted-key",
	}
	mockRepo.tx.addRootKeyErr = errMockServiceFailure

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateRootKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to store new root key")
}

// TestRotateIntermediateKey_NoIntermediateKeyFound tests error when no intermediate key exists.
func TestRotateIntermediateKey_NoIntermediateKeyFound(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.getIntermediateKeyLatestReturnsNil = true

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no intermediate key found")
}

// TestRotateIntermediateKey_NoRootKeyFound tests error when root key not found during intermediate rotation.
func TestRotateIntermediateKey_NoRootKeyFound(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: "dummy-encrypted-key",
		KEKUUID:   googleUuid.New(),
	}
	mockRepo.tx.getRootKeyLatestReturnsNil = true

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "no root key found")
}

// TestRotateIntermediateKey_DecryptRootKeyFailure tests error when decrypting root key fails.
func TestRotateIntermediateKey_DecryptRootKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Create a different unseal key to encrypt the root key (so decryption will fail)
	_, differentUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Generate root key and encrypt with the different unseal key
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	// Encrypt root key with different unseal key (not the one in unsealService)
	_, encryptedRootKey, err := cryptoutilSharedCryptoJose.EncryptKey([]joseJwk.Key{differentUnsealJWK}, clearRootJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: "dummy-encrypted-key",
		KEKUUID:   googleUuid.New(),
	}
	mockRepo.tx.rootKey = &RootKey{
		UUID:      *rootKeyUUID,
		Encrypted: string(encryptedRootKey), // Encrypted with wrong unseal key
	}

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to decrypt root key")
}

// TestRotateIntermediateKey_AddIntermediateKeyFailure tests error when adding new intermediate key fails.
func TestRotateIntermediateKey_AddIntermediateKeyFailure(t *testing.T) {
	t.Parallel()

	jwkGenService, unsealService := setupRotationServiceTestHelper(t)

	// Generate a valid encrypted root key
	rootKeyUUID, clearRootJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	encryptedRootKey, err := unsealService.EncryptKey(clearRootJWK)
	require.NoError(t, err)

	mockRepo := newMockServiceRepository()
	mockRepo.tx.intermediateKey = &IntermediateKey{
		UUID:      googleUuid.New(),
		Encrypted: "dummy-encrypted-key",
		KEKUUID:   googleUuid.New(),
	}
	mockRepo.tx.rootKey = &RootKey{
		UUID:      *rootKeyUUID,
		Encrypted: string(encryptedRootKey),
	}
	mockRepo.tx.addIntermediateKeyErr = errMockServiceFailure

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	ctx := context.Background()
	result, err := rotationService.RotateIntermediateKey(ctx, "test rotation")
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to store new intermediate key")
}

// TestRotateContentKey_NoIntermediateKeyFound tests error when no intermediate key exists.
