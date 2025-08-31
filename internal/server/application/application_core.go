package application

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	telemetryService "cryptoutil/internal/common/telemetry"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"
)

type ServerApplicationCore struct {
	TelemetryService  *telemetryService.TelemetryService
	SqlRepository     *cryptoutilSqlRepository.SqlRepository
	JwkGenService     *cryptoutilJose.JwkGenService
	OrmRepository     *cryptoutilOrmRepository.OrmRepository
	UnsealKeysService cryptoutilUnsealKeysService.UnsealKeysService
	BarrierService    *cryptoutilBarrierService.BarrierService
}

func StartServerApplicationCore(ctx context.Context, settings *cryptoutilConfig.Settings) (*ServerApplicationCore, error) {
	serverApplicationCore := &ServerApplicationCore{}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}
	serverApplicationCore.TelemetryService = telemetryService

	sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}
	serverApplicationCore.SqlRepository = sqlRepository

	jwkGenService, err := cryptoutilJose.NewJwkGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}
	serverApplicationCore.JwkGenService = jwkGenService

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, telemetryService, sqlRepository, jwkGenService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}
	serverApplicationCore.OrmRepository = ormRepository

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSettings(ctx, telemetryService, settings)
	if err != nil {
		telemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}
	serverApplicationCore.UnsealKeysService = unsealKeysService

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}
	serverApplicationCore.BarrierService = barrierService

	return serverApplicationCore, nil
}

func (c *ServerApplicationCore) Shutdown() func() {
	return func() {
		if c.TelemetryService != nil {
			c.TelemetryService.Slogger.Debug("stopping server core")
		}
		if c.BarrierService != nil {
			c.BarrierService.Shutdown()
		}
		if c.UnsealKeysService != nil {
			c.UnsealKeysService.Shutdown()
		}
		if c.OrmRepository != nil {
			c.OrmRepository.Shutdown()
		}
		if c.JwkGenService != nil {
			c.JwkGenService.Shutdown()
		}
		if c.SqlRepository != nil {
			c.SqlRepository.Shutdown()
		}
		if c.TelemetryService != nil {
			c.TelemetryService.Shutdown()
		}
	}
}
