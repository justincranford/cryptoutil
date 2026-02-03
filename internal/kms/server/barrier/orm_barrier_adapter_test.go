// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration

package barrier

import (
	"context"
	"fmt"
	"os"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilKmsServerRepositoryOrm "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	testSettings      = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("orm_barrier_adapter_test")
	testCtx           = context.Background()
	testTelemetry     *cryptoutilSharedTelemetry.TelemetryService
	testJWKGenService *cryptoutilSharedCryptoJose.JWKGenService
	testTemplateCore  *cryptoutilAppsTemplateServiceServerApplication.Core
	testGormDB        *gorm.DB
	testOrmRepo       *cryptoutilKmsServerRepositoryOrm.OrmRepository
	testRepoAdapter   *OrmRepositoryAdapter
)

func TestMain(m *testing.M) {
	var rc int

	func() {
		// Start template Core which provides GORM directly with proper migrations.
		var err error

		testTemplateCore, err = cryptoutilAppsTemplateServiceServerApplication.StartCore(testCtx, testSettings)
		if err != nil {
			panic(fmt.Sprintf("failed to start template core: %v", err))
		}

		defer func() {
			if testTemplateCore.ShutdownDBContainer != nil {
				testTemplateCore.ShutdownDBContainer()
			}
			testTemplateCore.Basic.Shutdown()
		}()

		testTelemetry = testTemplateCore.Basic.TelemetryService
		testJWKGenService = testTemplateCore.Basic.JWKGenService
		testGormDB = testTemplateCore.DB

		// Apply template migrations (1001-1005 for barrier tables, sessions, etc.).
		sqlDB, err := testGormDB.DB()
		if err != nil {
			panic(fmt.Sprintf("failed to get sql.DB from GORM: %v", err))
		}

		err = cryptoutilAppsTemplateServiceServerRepository.ApplyMigrationsFromFS(
			sqlDB,
			cryptoutilAppsTemplateServiceServerRepository.MigrationsFS,
			"migrations",
			"sqlite",
		)
		if err != nil {
			panic(fmt.Sprintf("failed to apply template migrations: %v", err))
		}

		// Create OrmRepository using template Core's GORM.
		testOrmRepo, err = cryptoutilKmsServerRepositoryOrm.NewOrmRepository(testCtx, testTelemetry, testGormDB, testJWKGenService, false)
		if err != nil {
			panic(fmt.Sprintf("failed to create OrmRepository: %v", err))
		}

		defer testOrmRepo.Shutdown()

		// Create adapter.
		testRepoAdapter = NewOrmRepositoryAdapter(testOrmRepo)
		defer testRepoAdapter.Shutdown()

		rc = m.Run()
	}()
	os.Exit(rc)
}

func cleanupDatabase(t *testing.T) {
	t.Helper()
	require.NoError(t, testGormDB.Exec("DELETE FROM barrier_root_keys").Error)
	require.NoError(t, testGormDB.Exec("DELETE FROM barrier_intermediate_keys").Error)
	require.NoError(t, testGormDB.Exec("DELETE FROM barrier_content_keys").Error)
}

// TestOrmRepositoryAdapter_ImplementsInterface verifies interface compliance at compile time.
func TestOrmRepositoryAdapter_ImplementsInterface(t *testing.T) {
	t.Parallel()

	// These are compile-time checks, the test just documents the assertion.
	var _ cryptoutilAppsTemplateServiceServerBarrier.Repository = (*OrmRepositoryAdapter)(nil)
	var _ cryptoutilAppsTemplateServiceServerBarrier.Transaction = (*OrmTransactionAdapter)(nil)
}

// TestOrmRepositoryAdapter_WithTransaction tests the transaction wrapper.
func TestOrmRepositoryAdapter_WithTransaction(t *testing.T) {
	t.Cleanup(func() { cleanupDatabase(t) })

	executed := false
	err := testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		executed = true
		require.NotNil(t, tx.Context())
		return nil
	})

	require.NoError(t, err)
	require.True(t, executed, "transaction function should have been executed")
}

// TestOrmTransactionAdapter_RootKeyOperations tests root key CRUD through adapter.
func TestOrmTransactionAdapter_RootKeyOperations(t *testing.T) {
	t.Cleanup(func() { cleanupDatabase(t) })

	var addedKeyUUID googleUuid.UUID

	// Test AddRootKey and GetRootKeyLatest.
	err := testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		addedKeyUUID = googleUuid.New()
		key := &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
			UUID:      addedKeyUUID,
			Encrypted: "encrypted-root-key-data",
			KEKUUID:   googleUuid.UUID{},
		}
		require.NoError(t, tx.AddRootKey(key))

		// Verify we can get the latest.
		latestKey, latestErr := tx.GetRootKeyLatest()
		require.NoError(t, latestErr)
		require.NotNil(t, latestKey)
		require.Equal(t, addedKeyUUID, latestKey.UUID)
		require.Equal(t, "encrypted-root-key-data", latestKey.Encrypted)

		return nil
	})
	require.NoError(t, err)

	// Test GetRootKey by UUID.
	err = testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		retrievedKey, getErr := tx.GetRootKey(&addedKeyUUID)
		require.NoError(t, getErr)
		require.NotNil(t, retrievedKey)
		require.Equal(t, addedKeyUUID, retrievedKey.UUID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransactionAdapter_IntermediateKeyOperations tests intermediate key CRUD through adapter.
func TestOrmTransactionAdapter_IntermediateKeyOperations(t *testing.T) {
	t.Cleanup(func() { cleanupDatabase(t) })

	var addedKeyUUID googleUuid.UUID
	parentUUID := googleUuid.New()

	// Test AddIntermediateKey and GetIntermediateKeyLatest.
	err := testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		addedKeyUUID = googleUuid.New()
		key := &cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{
			UUID:      addedKeyUUID,
			Encrypted: "encrypted-intermediate-key-data",
			KEKUUID:   parentUUID,
		}
		require.NoError(t, tx.AddIntermediateKey(key))

		// Verify we can get the latest.
		latestKey, latestErr := tx.GetIntermediateKeyLatest()
		require.NoError(t, latestErr)
		require.NotNil(t, latestKey)
		require.Equal(t, addedKeyUUID, latestKey.UUID)
		require.Equal(t, "encrypted-intermediate-key-data", latestKey.Encrypted)
		require.Equal(t, parentUUID, latestKey.KEKUUID)

		return nil
	})
	require.NoError(t, err)

	// Test GetIntermediateKey by UUID.
	err = testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		retrievedKey, getErr := tx.GetIntermediateKey(&addedKeyUUID)
		require.NoError(t, getErr)
		require.NotNil(t, retrievedKey)
		require.Equal(t, addedKeyUUID, retrievedKey.UUID)

		return nil
	})
	require.NoError(t, err)
}

// TestOrmTransactionAdapter_ContentKeyOperations tests content key CRUD through adapter.
func TestOrmTransactionAdapter_ContentKeyOperations(t *testing.T) {
	t.Cleanup(func() { cleanupDatabase(t) })

	var addedKeyUUID googleUuid.UUID
	parentUUID := googleUuid.New()

	// Test AddContentKey and GetContentKey.
	err := testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		addedKeyUUID = googleUuid.New()
		key := &cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
			UUID:      addedKeyUUID,
			Encrypted: "encrypted-content-key-data",
			KEKUUID:   parentUUID,
		}
		require.NoError(t, tx.AddContentKey(key))

		return nil
	})
	require.NoError(t, err)

	// Test GetContentKey by UUID.
	err = testRepoAdapter.WithTransaction(testCtx, func(tx cryptoutilAppsTemplateServiceServerBarrier.Transaction) error {
		retrievedKey, getErr := tx.GetContentKey(&addedKeyUUID)
		require.NoError(t, getErr)
		require.NotNil(t, retrievedKey)
		require.Equal(t, addedKeyUUID, retrievedKey.UUID)
		require.Equal(t, "encrypted-content-key-data", retrievedKey.Encrypted)
		require.Equal(t, parentUUID, retrievedKey.KEKUUID)

		return nil
	})
	require.NoError(t, err)
}

// TestConversionFunctions_NilHandling tests that conversion functions handle nil safely.
func TestConversionFunctions_NilHandling(t *testing.T) {
	t.Parallel()

	require.Nil(t, convertOrmRootKeyToBarrier(nil))
	require.Nil(t, convertBarrierRootKeyToOrm(nil))
	require.Nil(t, convertOrmIntermediateKeyToBarrier(nil))
	require.Nil(t, convertBarrierIntermediateKeyToOrm(nil))
	require.Nil(t, convertOrmContentKeyToBarrier(nil))
	require.Nil(t, convertBarrierContentKeyToOrm(nil))
}

// TestConversionFunctions_RoundTrip tests that conversion preserves all fields.
func TestConversionFunctions_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("RootKey round-trip", func(t *testing.T) {
		t.Parallel()
		original := &cryptoutilAppsTemplateServiceServerBarrier.RootKey{
			UUID:      googleUuid.New(),
			Encrypted: "test-encrypted",
			KEKUUID:   googleUuid.New(),
			CreatedAt: 1234567890,
			UpdatedAt: 1234567899,
		}
		ormKey := convertBarrierRootKeyToOrm(original)
		roundTrip := convertOrmRootKeyToBarrier(ormKey)
		require.Equal(t, original.UUID, roundTrip.UUID)
		require.Equal(t, original.Encrypted, roundTrip.Encrypted)
		require.Equal(t, original.KEKUUID, roundTrip.KEKUUID)
		require.Equal(t, original.CreatedAt, roundTrip.CreatedAt)
		require.Equal(t, original.UpdatedAt, roundTrip.UpdatedAt)
	})

	t.Run("IntermediateKey round-trip", func(t *testing.T) {
		t.Parallel()
		original := &cryptoutilAppsTemplateServiceServerBarrier.IntermediateKey{
			UUID:      googleUuid.New(),
			Encrypted: "test-encrypted",
			KEKUUID:   googleUuid.New(),
			CreatedAt: 1234567890,
			UpdatedAt: 1234567899,
		}
		ormKey := convertBarrierIntermediateKeyToOrm(original)
		roundTrip := convertOrmIntermediateKeyToBarrier(ormKey)
		require.Equal(t, original.UUID, roundTrip.UUID)
		require.Equal(t, original.Encrypted, roundTrip.Encrypted)
		require.Equal(t, original.KEKUUID, roundTrip.KEKUUID)
		require.Equal(t, original.CreatedAt, roundTrip.CreatedAt)
		require.Equal(t, original.UpdatedAt, roundTrip.UpdatedAt)
	})

	t.Run("ContentKey round-trip", func(t *testing.T) {
		t.Parallel()
		original := &cryptoutilAppsTemplateServiceServerBarrier.ContentKey{
			UUID:      googleUuid.New(),
			Encrypted: "test-encrypted",
			KEKUUID:   googleUuid.New(),
			CreatedAt: 1234567890,
			UpdatedAt: 1234567899,
		}
		ormKey := convertBarrierContentKeyToOrm(original)
		roundTrip := convertOrmContentKeyToBarrier(ormKey)
		require.Equal(t, original.UUID, roundTrip.UUID)
		require.Equal(t, original.Encrypted, roundTrip.Encrypted)
		require.Equal(t, original.KEKUUID, roundTrip.KEKUUID)
		require.Equal(t, original.CreatedAt, roundTrip.CreatedAt)
		require.Equal(t, original.UpdatedAt, roundTrip.UpdatedAt)
	})
}
