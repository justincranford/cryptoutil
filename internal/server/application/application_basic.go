package application

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
)

type ServerApplicationBasic struct {
	TelemetryService  *cryptoutilTelemetry.TelemetryService
	UnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
	JWKGenService     *cryptoutilJose.JWKGenService
}

func StartServerApplicationBasic(ctx context.Context, settings *cryptoutilConfig.Settings) (*ServerApplicationBasic, error) {
	serverApplicationBasic := &ServerApplicationBasic{}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}
	serverApplicationBasic.TelemetryService = telemetryService

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, serverApplicationBasic.TelemetryService, settings)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		serverApplicationBasic.Shutdown()
		return nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}
	serverApplicationBasic.UnsealKeysService = unsealKeysService

	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		serverApplicationBasic.Shutdown()
		return nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}
	serverApplicationBasic.JWKGenService = jwkGenService

	return serverApplicationBasic, nil
}

func (c *ServerApplicationBasic) Shutdown() func() {
	return func() {
		if c.TelemetryService != nil {
			c.TelemetryService.Slogger.Debug("stopping server basic")
		}
		if c.JWKGenService != nil {
			c.JWKGenService.Shutdown()
		}
		if c.UnsealKeysService != nil {
			c.UnsealKeysService.Shutdown()
		}
		if c.TelemetryService != nil {
			c.TelemetryService.Shutdown()
		}
	}
}
