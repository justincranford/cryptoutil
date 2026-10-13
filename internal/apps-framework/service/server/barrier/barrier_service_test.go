// Copyright (c) 2025-2026 Justin Cranford.
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

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

const (
	testErrorContextMustBeNonNil                          = "ctx must be non-nil"
	testCaseNilTelemetryService                           = "nil telemetry service"
	testErrorTelemetryServiceMustBeNonNil                 = "telemetryService must be non-nil"
	testCaseNilJWKGenService                              = "nil jwk gen service"
	testErrorJWKGenServiceMustBeNonNil                    = "jwkGenService must be non-nil"
	testCaseNilRepository                                 = "nil repository"
	testErrorRepositoryMustBeNonNil                       = "repository must be non-nil"
	testCaseNilUnsealKeysService                          = "nil unseal keys service"
	testErrorUnsealKeysServiceMustBeNonNil                = "unsealKeysService must be non-nil"
	testCaseInvalidJWEFormat                              = "invalid_jwe_format"
	testErrorFailedToParseEncryptedIntermediateKeyMessage = "failed to parse encrypted intermediate key message"
	testEncryptedRootKey1                                 = "encrypted_root_key_1"
	testEncryptedRootKey2                                 = "encrypted_root_key_2"
	testEncryptedRootKey                                  = "encrypted_root_key"
	testEncryptedIntermediateKey1                         = "encrypted_intermediate_key_1"
	testEncryptedIntermediateKey2                         = "encrypted_intermediate_key_2"
	testEncryptedContentKey1                              = "encrypted_content_key_1"
	testEncryptedContentKey2                              = "encrypted_content_key_2"
)

var (
	testDB               *gorm.DB
	testSQLDB            *sql.DB // Keep reference to prevent GC - in-memory SQLite requires open connection
	testService          *cryptoutilAppsFrameworkServiceServerBarrier.Service
	testJWKGenService    *cryptoutilSharedCryptoJose.JWKGenService
	testTelemetryService *cryptoutilSharedTelemetry.TelemetryService
)

// createBarrierTables creates the barrier encryption tables for testing.
func createBarrierTables(db *sql.DB) error {
	ctx := context.Background()

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

	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to create barrier tables: %w", err)
	}

	return nil
}

// TestService_EncryptDecrypt_Success tests successful encryption and decryption.
func TestService_EncryptDecrypt_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("test data for encryption")

	// Encrypt.
	ciphertext, err := testService.EncryptContentWithContext(ctx, plaintext)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotEmpty(t, ciphertext)
	require.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")

	// Decrypt.
	decrypted, err := testService.DecryptContentWithContext(ctx, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

// TestService_EncryptDecrypt_MultipleRounds tests multiple encryption/decryption cycles.
func TestService_EncryptDecrypt_MultipleRounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "short text",
			plaintext: []byte("short"),
		},
		{
			name:      "medium text",
			plaintext: []byte("This is a medium length plaintext for testing barrier encryption"),
		},
		{
			name:      "long text",
			plaintext: []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris."),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name:      "unicode text",
			plaintext: []byte("Hello 世界 🌍"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Encrypt.
			ciphertext, err := testService.EncryptContentWithContext(ctx, tt.plaintext)
			require.NoError(t, err)
			require.NotNil(t, ciphertext)
			require.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt.
			decrypted, err := testService.DecryptContentWithContext(ctx, ciphertext)
			require.NoError(t, err)
			require.Equal(t, tt.plaintext, decrypted)
		})
	}
}

// TestService_EncryptDecrypt_EmptyData tests that empty data returns an error.
func TestService_EncryptDecrypt_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("")

	// Encrypt empty data should fail with validation error.
	_, err := testService.EncryptContentWithContext(ctx, plaintext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwks can't be empty")
}

// TestService_EncryptBytesWithContext_AliasSuccess tests the alias method for encryption.
func TestService_EncryptBytesWithContext_AliasSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("test data for encryption alias")

	// Encrypt using alias method.
	ciphertext, err := testService.EncryptBytesWithContext(ctx, plaintext)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotEmpty(t, ciphertext)
	require.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")

	// Decrypt using alias method.
	decrypted, err := testService.DecryptBytesWithContext(ctx, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

// TestService_DecryptInvalidCiphertext tests decryption with invalid ciphertext.
func TestService_DecryptInvalidCiphertext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{
			name:       "garbage data",
			ciphertext: []byte("not a valid JWE"),
		},
		{
			name:       "empty data",
			ciphertext: []byte(""),
		},
		{
			name:       "malformed JSON",
			ciphertext: []byte("{invalid json}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := testService.DecryptContentWithContext(ctx, tt.ciphertext)
			require.Error(t, err, "Decryption should fail for invalid ciphertext")
		})
	}
}

// TestService_Shutdown tests service shutdown behavior.
func TestService_Shutdown(t *testing.T) {
	t.Parallel()
	// NOTE: Cannot run parallel - creates isolated database but takes exclusive test time.
	ctx := context.Background()

	// Create isolated in-memory database for this test.
	dbUUID, _ := googleUuid.NewV7()
	dsn := fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbUUID.String())
	shutdownSQLDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, shutdownSQLDB.Close())
	}()

	_, err = shutdownSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = shutdownSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	shutdownSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	shutdownSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	shutdownSQLDB.SetConnMaxLifetime(0)

	shutdownDB, err := gorm.Open(sqlite.Dialector{Conn: shutdownSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	err = createBarrierTables(shutdownSQLDB)
	require.NoError(t, err)

	// Create isolated barrier service for shutdown testing.
	telemetrySvc, err := newBarrierTestTelemetryService(t)
	require.NoError(t, err)

	jwkGenSvc, err := newBarrierTestJWKGenService(t, telemetrySvc)
	require.NoError(t, err)

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(shutdownDB)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	service, err := cryptoutilAppsFrameworkServiceServerBarrier.NewService(
		ctx,
		telemetrySvc,
		jwkGenSvc,
		barrierRepo,
		unsealService,
	)
	require.NoError(t, err)

	// Verify service works before shutdown.
	plaintext := []byte("test before shutdown")
	ciphertext, err := service.EncryptContentWithContext(ctx, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, ciphertext)

	// Shutdown service.
	service.Shutdown()

	// Verify operations fail after shutdown.
	_, err = service.EncryptContentWithContext(ctx, plaintext)
	require.Error(t, err, "Encryption should fail after shutdown")
	require.Contains(t, err.Error(), "barrier service is closed")

	_, err = service.DecryptContentWithContext(ctx, ciphertext)
	require.Error(t, err, "Decryption should fail after shutdown")
	require.Contains(t, err.Error(), "barrier service is closed")

	// Verify multiple shutdowns are safe (idempotent).
	service.Shutdown()
	service.Shutdown()
}

// TestService_ConcurrentEncryption tests concurrent encryption operations.
