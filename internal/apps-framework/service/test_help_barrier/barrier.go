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

// NewTestBarrierService builds a barrier service fixture using the provided GORM DB.
//
// This helper creates a telemetry service, JWK generator, unseal key service, and
// barrier repository wired into a ready-to-use barrier service instance.
func NewTestBarrierService(t *testing.T, db *gorm.DB) *cryptoutilAppsFrameworkServiceServerBarrier.Service {
	t.Helper()

	if db == nil {
		t.Fatal("test_help_barrier: db must be non-nil")
	}

	ctx := context.Background()

	telemetrySettings := cryptoutilAppsFrameworkServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, telemetrySettings.ToTelemetrySettings())
	if err != nil {
		t.Fatalf("test_help_barrier: create telemetry service: %v", err)
	}

	t.Cleanup(func() {
		telemetryService.Shutdown()
	})

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	if err != nil {
		t.Fatalf("test_help_barrier: create JWK generation service: %v", err)
	}

	t.Cleanup(func() {
		jwkGenService.Shutdown()
	})

	_, testUnsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	if err != nil {
		t.Fatalf("test_help_barrier: generate unseal JWK: %v", err)
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{testUnsealJWK})
	if err != nil {
		t.Fatalf("test_help_barrier: create unseal keys service: %v", err)
	}

	t.Cleanup(func() {
		unsealKeysService.Shutdown()
	})

	barrierRepo, err := cryptoutilAppsFrameworkServiceServerBarrier.NewGormRepository(db)
	if err != nil {
		t.Fatalf("test_help_barrier: create barrier repository: %v", err)
	}

	t.Cleanup(func() {
		barrierRepo.Shutdown()
	})

	barrierService, err := cryptoutilAppsFrameworkServiceServerBarrier.NewService(ctx, telemetryService, jwkGenService, barrierRepo, unsealKeysService)
	if err != nil {
		t.Fatalf("test_help_barrier: create barrier service: %v", err)
	}

	t.Cleanup(func() {
		barrierService.Shutdown()
	})

	return barrierService
}
