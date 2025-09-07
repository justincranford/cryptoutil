package application

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	telemetryService "cryptoutil/internal/common/telemetry"
)

type ServerApplicationBasic struct {
	TelemetryService *telemetryService.TelemetryService
	JwkGenService    *cryptoutilJose.JwkGenService
}

func StartServerApplicationBasic(ctx context.Context, settings *cryptoutilConfig.Settings) (*ServerApplicationBasic, error) {
	serverApplicationBasic := &ServerApplicationBasic{}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}
	serverApplicationBasic.TelemetryService = telemetryService

	jwkGenService, err := cryptoutilJose.NewJwkGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		serverApplicationBasic.Shutdown()
		return nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}
	serverApplicationBasic.JwkGenService = jwkGenService

	return serverApplicationBasic, nil
}

func (c *ServerApplicationBasic) Shutdown() func() {
	return func() {
		if c.TelemetryService != nil {
			c.TelemetryService.Slogger.Debug("stopping server core")
		}
		if c.JwkGenService != nil {
			c.JwkGenService.Shutdown()
		}
		if c.TelemetryService != nil {
			c.TelemetryService.Shutdown()
		}
	}
}
