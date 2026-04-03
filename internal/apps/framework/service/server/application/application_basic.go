// Copyright (c) 2025 Justin Cranford
//
//

// Package application provides server application lifecycle management components.
package application

import (
	"context"
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/framework/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Basic encapsulates basic service infrastructure (telemetry, unseal, JWK generation).
// This is the foundation layer used by Core.
type Basic struct {
	TelemetryService  *cryptoutilSharedTelemetry.TelemetryService
	UnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
	JWKGenService     *cryptoutilSharedCryptoJose.JWKGenService
	Settings          *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
}

// StartBasic initializes basic service infrastructure.
// This includes telemetry, unseal keys, and JWK generation services.
func StartBasic(ctx context.Context, settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Basic, error) {
	return startBasicInternal(ctx, settings, cryptoutilSharedCryptoJose.NewJWKGenService)
}

func startBasicInternal(
	ctx context.Context,
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	newJWKGenServiceFn func(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, devMode bool) (*cryptoutilSharedCryptoJose.JWKGenService, error),
) (*Basic, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	app := &Basic{Settings: settings}

	// Initialize telemetry service.
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings.ToTelemetrySettings())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	app.TelemetryService = telemetryService

	// Initialize unseal keys service.
	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, app.TelemetryService, settings)
	if err != nil {
		app.TelemetryService.Slogger.Error("failed to create unseal repository", cryptoutilSharedMagic.StringError, err)
		app.Shutdown()

		return nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	app.UnsealKeysService = unsealKeysService

	// Initialize JWK Generation Service.
	jwkGenService, err := newJWKGenServiceFn(ctx, telemetryService, settings.VerboseMode)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", cryptoutilSharedMagic.StringError, err)
		app.Shutdown()

		return nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}

	app.JWKGenService = jwkGenService

	return app, nil
}

// Shutdown gracefully shuts down all basic services (LIFO order).
func (a *Basic) Shutdown() {
	if a.TelemetryService != nil {
		a.TelemetryService.Slogger.Debug("stopping application basic")
	}

	if a.JWKGenService != nil {
		a.JWKGenService.Shutdown()
	}

	if a.UnsealKeysService != nil {
		a.UnsealKeysService.Shutdown()
	}

	if a.TelemetryService != nil {
		a.TelemetryService.Shutdown()
	}
}
