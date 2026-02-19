// Copyright (c) 2025 Justin Cranford
//

package barrier_test

import (
	"context"
	"database/sql"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"
)

func TestService_ConcurrentEncryption(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	const numGoroutines = 10

	// Launch multiple concurrent encryption operations.
	results := make(chan []byte, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			plaintext := []byte("concurrent test data " + string(rune(id)))

			ciphertext, err := testService.EncryptContentWithContext(ctx, plaintext)
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

// TestNewService_ValidationErrors tests constructor validation paths.
func TestNewService_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name               string
		ctx                context.Context
		telemetryService   *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService      *cryptoutilSharedCryptoJose.JWKGenService
		repository         cryptoutilAppsTemplateServiceServerBarrier.Repository
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

			service, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(
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

// TestNewService_NilUnsealService tests nil unseal service validation.
func TestNewService_NilUnsealService(t *testing.T) {
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
	validSQLDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	validSQLDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	validSQLDB.SetConnMaxLifetime(0)

	validDB, err := gorm.Open(sqlite.Dialector{Conn: validSQLDB}, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	require.NoError(t, createBarrierTables(validSQLDB))

	repo, err := cryptoutilAppsTemplateServiceServerBarrier.NewGormRepository(validDB)
	require.NoError(t, err)

	defer repo.Shutdown()

	service, err := cryptoutilAppsTemplateServiceServerBarrier.NewService(
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
