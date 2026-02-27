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
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

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
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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
				telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
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

	telemetrySvc, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true).ToTelemetrySettings())
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	// Create DB but DON'T initialize any root keys
	dbUUID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
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
