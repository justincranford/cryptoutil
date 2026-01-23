// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

var (
	testDB               *gorm.DB
	testSQLDB            *sql.DB // Keep reference to prevent GC - in-memory SQLite requires open connection
	testBarrierService   *cryptoutilTemplateBarrier.BarrierService
	testJWKGenService    *cryptoutilJose.JWKGenService
	testTelemetryService *cryptoutilTelemetry.TelemetryService
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup: Create shared heavyweight resources ONCE.
	dbID, _ := googleUuid.NewV7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	var err error

	testSQLDB, err = sql.Open("sqlite", dsn)
	if err != nil {
		panic("TestMain: failed to open SQLite: " + err.Error())
	}

	// Configure SQLite for concurrent operations.
	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		panic("TestMain: failed to enable WAL: " + err.Error())
	}

	if _, err := testSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
		panic("TestMain: failed to set busy timeout: " + err.Error())
	}

	testSQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	testSQLDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	testDB, err = gorm.Open(sqlite.Dialector{Conn: testSQLDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic("TestMain: failed to create GORM DB: " + err.Error())
	}

	// Create barrier tables.
	if err := createBarrierTables(testSQLDB); err != nil {
		panic("TestMain: failed to create tables: " + err.Error())
	}

	// Initialize telemetry.
	testTelemetryService, err = cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	if err != nil {
		panic("TestMain: failed to create telemetry: " + err.Error())
	}
	defer testTelemetryService.Shutdown()

	// Initialize JWK gen service.
	testJWKGenService, err = cryptoutilJose.NewJWKGenService(ctx, testTelemetryService, false)
	if err != nil {
		panic("TestMain: failed to create JWK gen service: " + err.Error())
	}
	defer testJWKGenService.Shutdown()

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := testJWKGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	if err != nil {
		panic("TestMain: failed to generate unseal JWK: " + err.Error())
	}

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	if err != nil {
		panic("TestMain: failed to create unseal service: " + err.Error())
	}
	defer unsealService.Shutdown()

	// Create barrier repository.
	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(testDB)
	if err != nil {
		panic("TestMain: failed to create barrier repository: " + err.Error())
	}
	defer barrierRepo.Shutdown()

	// Create barrier service.
	testBarrierService, err = cryptoutilTemplateBarrier.NewBarrierService(
		ctx,
		testTelemetryService,
		testJWKGenService,
		barrierRepo,
		unsealService,
	)
	if err != nil {
		panic("TestMain: failed to create barrier service: " + err.Error())
	}

	defer testBarrierService.Shutdown()
	defer func() {
		if closeErr := testSQLDB.Close(); closeErr != nil {
			panic("TestMain: failed to close test SQL DB: " + closeErr.Error())
		}
	}()

	// Run all tests - defer statements execute cleanup AFTER m.Run() completes.
	exitCode := m.Run()

	// Exit with test result code.
	os.Exit(exitCode)
}

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

// TestBarrierService_EncryptDecrypt_Success tests successful encryption and decryption.
func TestBarrierService_EncryptDecrypt_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("test data for encryption")

	// Encrypt.
	ciphertext, err := testBarrierService.EncryptContentWithContext(ctx, plaintext)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotEmpty(t, ciphertext)
	require.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")

	// Decrypt.
	decrypted, err := testBarrierService.DecryptContentWithContext(ctx, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

// TestBarrierService_EncryptDecrypt_MultipleRounds tests multiple encryption/decryption cycles.
func TestBarrierService_EncryptDecrypt_MultipleRounds(t *testing.T) {
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
			ciphertext, err := testBarrierService.EncryptContentWithContext(ctx, tt.plaintext)
			require.NoError(t, err)
			require.NotNil(t, ciphertext)
			require.NotEqual(t, tt.plaintext, ciphertext)

			// Decrypt.
			decrypted, err := testBarrierService.DecryptContentWithContext(ctx, ciphertext)
			require.NoError(t, err)
			require.Equal(t, tt.plaintext, decrypted)
		})
	}
}

// TestBarrierService_EncryptDecrypt_EmptyData tests that empty data returns an error.
func TestBarrierService_EncryptDecrypt_EmptyData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("")

	// Encrypt empty data should fail with validation error.
	_, err := testBarrierService.EncryptContentWithContext(ctx, plaintext)
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwks can't be empty")
}

// TestBarrierService_EncryptBytesWithContext_AliasSuccess tests the alias method for encryption.
func TestBarrierService_EncryptBytesWithContext_AliasSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	plaintext := []byte("test data for encryption alias")

	// Encrypt using alias method.
	ciphertext, err := testBarrierService.EncryptBytesWithContext(ctx, plaintext)
	require.NoError(t, err)
	require.NotNil(t, ciphertext)
	require.NotEmpty(t, ciphertext)
	require.NotEqual(t, plaintext, ciphertext, "Ciphertext should differ from plaintext")

	// Decrypt using alias method.
	decrypted, err := testBarrierService.DecryptBytesWithContext(ctx, ciphertext)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted, "Decrypted data should match original plaintext")
}

// TestBarrierService_DecryptInvalidCiphertext tests decryption with invalid ciphertext.
func TestBarrierService_DecryptInvalidCiphertext(t *testing.T) {
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

			_, err := testBarrierService.DecryptContentWithContext(ctx, tt.ciphertext)
			require.Error(t, err, "Decryption should fail for invalid ciphertext")
		})
	}
}

// TestBarrierService_Shutdown tests service shutdown behavior.
func TestBarrierService_Shutdown(t *testing.T) {
	// NOTE: Cannot run parallel - creates isolated database but takes exclusive test time.
	ctx := context.Background()

	// Create isolated in-memory database for this test.
	dbUUID, _ := googleUuid.NewV7()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbUUID.String())
	shutdownSQLDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, shutdownSQLDB.Close())
	}()

	_, err = shutdownSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = shutdownSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	shutdownSQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	shutdownSQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	shutdownSQLDB.SetConnMaxLifetime(0)

	shutdownDB, err := gorm.Open(sqlite.Dialector{Conn: shutdownSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	err = createBarrierTables(shutdownSQLDB)
	require.NoError(t, err)

	// Create isolated barrier service for shutdown testing.
	telemetrySvc, err := cryptoutilTelemetry.NewTelemetryService(ctx, cryptoutilConfig.NewTestConfig(cryptoutilMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetrySvc.Shutdown() })

	jwkGenSvc, err := cryptoutilJose.NewJWKGenService(ctx, telemetrySvc, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenSvc.Shutdown() })

	_, unsealJWK, _, _, _, err := jwkGenSvc.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	barrierRepo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(shutdownDB)
	require.NoError(t, err)
	t.Cleanup(func() { barrierRepo.Shutdown() })

	service, err := cryptoutilTemplateBarrier.NewBarrierService(
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

// TestBarrierService_ConcurrentEncryption tests concurrent encryption operations.
func TestBarrierService_ConcurrentEncryption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const numGoroutines = 10

	// Launch multiple concurrent encryption operations.
	results := make(chan []byte, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			plaintext := []byte("concurrent test data " + string(rune(id)))

			ciphertext, err := testBarrierService.EncryptContentWithContext(ctx, plaintext)
			if err != nil {
				errors <- err

				return
			}

			results <- ciphertext
		}(i)
	}

	// Collect results.
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-errors:
			require.NoError(t, err, "Concurrent encryption should not fail")
		case ciphertext := <-results:
			require.NotEmpty(t, ciphertext, "Ciphertext should not be empty")
		}
	}
}

// TestNewBarrierService_ValidationErrors tests constructor validation paths.
func TestNewBarrierService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name               string
		ctx                context.Context
		telemetryService   *cryptoutilTelemetry.TelemetryService
		jwkGenService      *cryptoutilJose.JWKGenService
		repository         cryptoutilTemplateBarrier.BarrierRepository
		unsealKeysService  cryptoutilUnsealKeysService.UnsealKeysService
		expectedErrContain string
	}{
		{
			name:               "nil context",
			ctx:                nil,
			telemetryService:   testTelemetryService,
			jwkGenService:      testJWKGenService,
			repository:         nil, // We'll use a mock.
			unsealKeysService:  nil, // We'll use a mock.
			expectedErrContain: "ctx must be non-nil",
		},
		{
			name:               "nil telemetry service",
			ctx:                ctx,
			telemetryService:   nil,
			jwkGenService:      testJWKGenService,
			repository:         nil,
			unsealKeysService:  nil,
			expectedErrContain: "telemetryService must be non-nil",
		},
		{
			name:               "nil jwk gen service",
			ctx:                ctx,
			telemetryService:   testTelemetryService,
			jwkGenService:      nil,
			repository:         nil,
			unsealKeysService:  nil,
			expectedErrContain: "jwkGenService must be non-nil",
		},
		{
			name:               "nil repository",
			ctx:                ctx,
			telemetryService:   testTelemetryService,
			jwkGenService:      testJWKGenService,
			repository:         nil,
			unsealKeysService:  nil,
			expectedErrContain: "repository must be non-nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service, err := cryptoutilTemplateBarrier.NewBarrierService(
				tt.ctx,
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

// TestNewBarrierService_NilUnsealService tests nil unseal service validation.
func TestNewBarrierService_NilUnsealService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a valid repository for this test.
	dbID, _ := googleUuid.NewV7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"
	validSQLDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, validSQLDB.Close())
	}()

	_, err = validSQLDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = validSQLDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)
	validSQLDB.SetMaxOpenConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	validSQLDB.SetMaxIdleConns(cryptoutilMagic.SQLiteMaxOpenConnections)
	validSQLDB.SetConnMaxLifetime(0)

	validDB, err := gorm.Open(sqlite.Dialector{Conn: validSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	require.NoError(t, createBarrierTables(validSQLDB))

	repo, err := cryptoutilTemplateBarrier.NewGormBarrierRepository(validDB)
	require.NoError(t, err)
	defer repo.Shutdown()

	service, err := cryptoutilTemplateBarrier.NewBarrierService(
		ctx,
		testTelemetryService,
		testJWKGenService,
		repo,
		nil, // nil unseal service
	)
	require.Error(t, err)
	require.Nil(t, service)
	require.Contains(t, err.Error(), "unsealKeysService must be non-nil")
}
