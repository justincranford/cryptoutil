// Copyright (c) 2025-2026 Justin Cranford.

// Package test_help_barrier provides barrier and unseal key fixture composition helpers
// for integration and E2E test suites that need encryption-at-rest (barrier layer) support.
// It handles barrier service setup, unseal key derivation, and elastic key ring initialization.
//
// Consumed by:
//   - test_orch_integration: optional fixture for barrier-heavy tests
//   - Repository test suites: barrier service fixtures
//   - API test suites: barrier-protected resource fixtures
package test_help_barrier

import (
	"context"
	"fmt"
	"testing"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"gorm.io/gorm"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilAppsFrameworkServiceServerBarrier "cryptoutil/internal/apps-framework/service/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps-framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

type barrierDeps struct {
	newTelemetryServiceFn  func(context.Context, *cryptoutilSharedTelemetry.TelemetrySettings) (*cryptoutilSharedTelemetry.TelemetryService, error)
	newJWKGenServiceFn     func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, bool) (*cryptoutilSharedCryptoJose.JWKGenService, error)
	generateUnsealJWKFn    func(*cryptoutilSharedCryptoJose.JWKGenService) (joseJwk.Key, error)
	newUnsealKeysServiceFn func([]joseJwk.Key) (cryptoutilUnsealKeysService.UnsealKeysService, error)
	newBarrierRepoFn       func(*gorm.DB) (*cryptoutilAppsFrameworkServiceServerBarrier.GormRepository, error)
	newBarrierServiceFn    func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, *cryptoutilSharedCryptoJose.JWKGenService, cryptoutilAppsFrameworkServiceServerBarrier.Repository, cryptoutilUnsealKeysService.UnsealKeysService) (*cryptoutilAppsFrameworkServiceServerBarrier.Service, error)
	autoMigrateBarrierFn   func(*gorm.DB) error
}

func defaultBarrierDeps() barrierDeps {
	return barrierDeps{
		newTelemetryServiceFn: cryptoutilSharedTelemetry.NewTelemetryService,
		newJWKGenServiceFn:    cryptoutilSharedCryptoJose.NewJWKGenService,
		generateUnsealJWKFn: func(jwkGenService *cryptoutilSharedCryptoJose.JWKGenService) (joseJwk.Key, error) {
			_, testUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)

			wrappedErr := wrapGenerateJWEJWKError(err)

			return testUnsealJWK, wrappedErr
		},
		newUnsealKeysServiceFn: cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple,
		newBarrierRepoFn:       cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository,
		newBarrierServiceFn:    cryptoutilAppsFrameworkServiceServerBarrier.NewService,
		autoMigrateBarrierFn: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&cryptoutilAppsFrameworkServiceServerBarrier.RootKey{},
				&cryptoutilAppsFrameworkServiceServerBarrier.IntermediateKey{},
				&cryptoutilAppsFrameworkServiceServerBarrier.ContentKey{},
			)
		},
	}
}

func wrapGenerateJWEJWKError(err error) error {
	if err != nil {
		return fmt.Errorf("GenerateJWEJWK: %w", err)
	}

	return nil
}

// NewTestBarrierService builds a barrier service fixture using the provided GORM DB.
//
// This helper creates a telemetry service, JWK generator, unseal key service, and
// barrier repository wired into a ready-to-use barrier service instance.
func NewTestBarrierService(t *testing.T, db *gorm.DB) *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	t.Helper()

	barrierService, err := newTestBarrierService(context.Background(), t, db)
	if err != nil {
		panic(fmt.Sprintf("test_help_barrier: create barrier service: %v", err))
	}

	return barrierService
}

func newTestBarrierService(ctx context.Context, t *testing.T, db *gorm.DB) (*cryptoutilAppsFrameworkServiceServerBarrier.Service, error) {
	return newTestBarrierServiceWithDeps(ctx, t, db, defaultBarrierDeps())
}

func newTestBarrierServiceWithDeps(ctx context.Context, t *testing.T, db *gorm.DB, deps barrierDeps) (*cryptoutilAppsFrameworkServiceServerBarrier.Service, error) {
	t.Helper()

	if db == nil {
		return nil, fmt.Errorf("db must be non-nil")
	}

	if err := deps.autoMigrateBarrierFn(db); err != nil {
		return nil, fmt.Errorf("migrate barrier tables: %w", err)
	}

	telemetrySettings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetryService, err := deps.newTelemetryServiceFn(ctx, telemetrySettings.ToTelemetrySettings())
	if err != nil {
		return nil, fmt.Errorf("create telemetry service: %w", err)
	}

	t.Cleanup(func() {
		telemetryService.Shutdown()
	})

	jwkGenService, err := deps.newJWKGenServiceFn(ctx, telemetryService, false)
	if err != nil {
		return nil, fmt.Errorf("create JWK generation service: %w", err)
	}

	t.Cleanup(func() {
		jwkGenService.Shutdown()
	})

	testUnsealJWK, err := deps.generateUnsealJWKFn(jwkGenService)
	if err != nil {
		return nil, fmt.Errorf("generate unseal JWK: %w", err)
	}

	unsealKeysService, err := deps.newUnsealKeysServiceFn([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		return nil, fmt.Errorf("create unseal keys service: %w", err)
	}

	t.Cleanup(func() {
		unsealKeysService.Shutdown()
	})

	barrierRepo, err := deps.newBarrierRepoFn(db)
	if err != nil {
		return nil, fmt.Errorf("create barrier repository: %w", err)
	}

	t.Cleanup(func() {
		barrierRepo.Shutdown()
	})

	barrierService, err := deps.newBarrierServiceFn(ctx, telemetryService, jwkGenService, barrierRepo, unsealKeysService)
	if err != nil {
		return nil, fmt.Errorf("create barrier service: %w", err)
	}

	t.Cleanup(func() {
		barrierService.Shutdown()
	})

	return barrierService, nil
}
