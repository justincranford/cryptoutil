// Copyright (c) 2025-2026 Justin Cranford.

package test_help_barrier

import (
	"context"
	"errors"
	"testing"

	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewTestBarrierService_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		withDB    bool
		wantError string
	}{
		{name: "nil db returns error", withDB: false, wantError: "db must be non-nil"},
		{name: "in-memory db creates barrier service", withDB: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var db *gorm.DB
			if tc.withDB {
				db = cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
			}

			svc, err := newTestBarrierService(context.Background(), t, db)
			if tc.wantError != "" {
				require.Nil(t, svc)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, svc)

			ciphertext, encErr := svc.EncryptContentWithContext(context.Background(), []byte("hello"))
			require.NoError(t, encErr)
			require.NotEmpty(t, ciphertext)

			plaintext, decErr := svc.DecryptContentWithContext(context.Background(), ciphertext)
			require.NoError(t, decErr)
			require.Equal(t, []byte("hello"), plaintext)
		})
	}
}

func TestNewTestBarrierService_WrapperSuccess(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
	svc := NewTestBarrierService(t, db)
	require.NotNil(t, svc)
}

func TestNewTestBarrierService_InjectedErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		override  func(barrierDeps) barrierDeps
		wantError string
	}{
		{
			name: "repository error",
			override: func(deps barrierDeps) barrierDeps {
				deps.newBarrierRepoFn = func(*gorm.DB) (*cryptoutilAppsFrameworkServiceServerBarrier.GormRepository, error) {
					return nil, errors.New("injected barrier repository failure")
				}

				return deps
			},
			wantError: "create barrier repository",
		},
		{
			name: "migration error",
			override: func(deps barrierDeps) barrierDeps {
				deps.autoMigrateBarrierFn = func(*gorm.DB) error { return errors.New("injected migration failure") }

				return deps
			},
			wantError: "migrate barrier tables",
		},
		{
			name: "jwk gen error",
			override: func(deps barrierDeps) barrierDeps {
				deps.newJWKGenServiceFn = func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, bool) (*cryptoutilSharedCryptoJose.JWKGenService, error) {
					return nil, errors.New("injected jwk generation failure")
				}

				return deps
			},
			wantError: "create JWK generation service",
		},
		{
			name: "telemetry error",
			override: func(deps barrierDeps) barrierDeps {
				deps.newTelemetryServiceFn = func(context.Context, *cryptoutilSharedTelemetry.TelemetrySettings) (*cryptoutilSharedTelemetry.TelemetryService, error) {
					return nil, errors.New("injected telemetry failure")
				}

				return deps
			},
			wantError: "create telemetry service",
		},
		{
			name: "generate unseal jwk error",
			override: func(deps barrierDeps) barrierDeps {
				deps.generateUnsealJWKFn = func(*cryptoutilSharedCryptoJose.JWKGenService) (joseJwk.Key, error) {
					return nil, errors.New("injected unseal jwk failure")
				}

				return deps
			},
			wantError: "generate unseal JWK",
		},
		{
			name: "unseal service error",
			override: func(deps barrierDeps) barrierDeps {
				deps.newUnsealKeysServiceFn = func([]joseJwk.Key) (cryptoutilUnsealKeysService.UnsealKeysService, error) {
					return nil, errors.New("injected unseal service failure")
				}

				return deps
			},
			wantError: "create unseal keys service",
		},
		{
			name: "barrier service error",
			override: func(deps barrierDeps) barrierDeps {
				deps.newBarrierServiceFn = func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, *cryptoutilSharedCryptoJose.JWKGenService, cryptoutilAppsFrameworkServiceServerBarrier.Repository, cryptoutilUnsealKeysService.UnsealKeysService) (*cryptoutilAppsFrameworkServiceServerBarrier.Service, error) {
					return nil, errors.New("injected barrier service failure")
				}

				return deps
			},
			wantError: "create barrier service",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
			deps := tc.override(defaultBarrierDeps())

			svc, err := newTestBarrierServiceWithDeps(context.Background(), t, db, deps)
			require.Nil(t, svc)
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantError)
		})
	}
}

func TestNewTestBarrierService_WrapperPanicsOnError(t *testing.T) {
	t.Parallel()

	db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
	deps := defaultBarrierDeps()
	deps.autoMigrateBarrierFn = func(*gorm.DB) error { return errors.New("injected panic path") }

	require.Panics(t, func() {
		_, err := newTestBarrierServiceWithDeps(context.Background(), t, db, deps)
		if err != nil {
			panic(err)
		}
	})
}

func TestWrapGenerateJWEJWKError_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		wantErr string
	}{
		{name: "nil error", err: nil},
		{name: "wrapped error", err: errors.New("raw generate error"), wantErr: "GenerateJWEJWK"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wrappedErr := wrapGenerateJWEJWKError(tc.err)
			if tc.wantErr == "" {
				require.NoError(t, wrappedErr)

				return
			}

			require.Error(t, wrappedErr)
			require.ErrorContains(t, wrappedErr, tc.wantErr)
		})
	}
}
