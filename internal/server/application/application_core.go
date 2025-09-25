package application

import (
	"context"
	"fmt"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/server/repository/sqlrepository"
)

type ServerApplicationCore struct {
	ServerApplicationBasic *ServerApplicationBasic
	SQLRepository          *cryptoutilSQLRepository.SQLRepository
	OrmRepository          *cryptoutilOrmRepository.OrmRepository
	BarrierService         *cryptoutilBarrierService.BarrierService
	BusinessLogicService   *cryptoutilBusinessLogic.BusinessLogicService
}

func StartServerApplicationCore(ctx context.Context, settings *cryptoutilConfig.Settings) (*ServerApplicationCore, error) {
	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start basic server application: %w", err)
	}
	jwkGenService := serverApplicationBasic.JwkGenService

	serverApplicationCore := &ServerApplicationCore{}
	serverApplicationCore.ServerApplicationBasic = serverApplicationBasic

	sqlRepository, err := cryptoutilSQLRepository.NewSQLRepository(ctx, serverApplicationBasic.TelemetryService, settings)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}
	serverApplicationCore.SQLRepository = sqlRepository

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, serverApplicationBasic.TelemetryService, sqlRepository, jwkGenService, settings)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}
	serverApplicationCore.OrmRepository = ormRepository

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, serverApplicationBasic.UnsealKeysService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}
	serverApplicationCore.BarrierService = barrierService

	businessLogicService, err := cryptoutilBusinessLogic.NewBusinessLogicService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		serverApplicationCore.Shutdown()
		return nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}
	serverApplicationCore.BusinessLogicService = businessLogicService

	return serverApplicationCore, nil
}

func (c *ServerApplicationCore) Shutdown() func() {
	return func() {
		if c.ServerApplicationBasic.TelemetryService != nil {
			c.ServerApplicationBasic.TelemetryService.Slogger.Debug("stopping server core")
		}
		if c.BarrierService != nil {
			c.BarrierService.Shutdown()
		}
		if c.OrmRepository != nil {
			c.OrmRepository.Shutdown()
		}
		if c.SQLRepository != nil {
			c.SQLRepository.Shutdown()
		}
		if c.ServerApplicationBasic != nil {
			c.ServerApplicationBasic.Shutdown()
		}
	}
}
