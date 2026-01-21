// Copyright (c) 2025 Justin Cranford
//
//

// Package application provides server application lifecycle management components.
package application

import (
	"context"
	"fmt"

	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// ApplicationBasic encapsulates basic service infrastructure (telemetry, unseal, JWK generation).
// This is the foundation layer used by ApplicationCore.
type ApplicationBasic struct {
	TelemetryService  *cryptoutilTelemetry.TelemetryService
	UnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
	JWKGenService     *cryptoutilJose.JWKGenService
	Settings          *cryptoutilConfig.ServiceTemplateServerSettings
}

// StartApplicationBasic initializes basic service infrastructure.
// This includes telemetry, unseal keys, and JWK generation services.
func StartApplicationBasic(ctx context.Context, settings *cryptoutilConfig.ServiceTemplateServerSettings) (*ApplicationBasic, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx cannot be nil")
	} else if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	app := &ApplicationBasic{Settings: settings}

	// Initialize telemetry service.
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	app.TelemetryService = telemetryService

	// Initialize unseal keys service.
	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, app.TelemetryService, settings)
	if err != nil {
		app.TelemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		app.Shutdown()

		return nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	app.UnsealKeysService = unsealKeysService

	// Initialize JWK Generation Service.
	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, settings.VerboseMode)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		app.Shutdown()

		return nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}

	app.JWKGenService = jwkGenService

	return app, nil
}

// Shutdown gracefully shuts down all basic services (LIFO order).
func (a *ApplicationBasic) Shutdown() {
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
