// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilKmsServerBusinesslogic "cryptoutil/internal/kms/server/businesslogic"
	cryptoutilKmsServerDemo "cryptoutil/internal/kms/server/demo"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilSQLRepository "cryptoutil/internal/kms/server/repository/sqlrepository"
	cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
)

// ServerApplicationCore provides core server application components including database, ORM, barrier, and business logic services.
type ServerApplicationCore struct {
	ServerApplicationBasic *ServerApplicationBasic
	SQLRepository          *cryptoutilSQLRepository.SQLRepository
	OrmRepository          *cryptoutilOrmRepository.OrmRepository
	BarrierService         *cryptoutilBarrierService.BarrierService
	BusinessLogicService   *cryptoutilKmsServerBusinesslogic.BusinessLogicService
	Settings               *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings
}

// StartServerApplicationCore initializes and starts a core server application with all essential services.
func StartServerApplicationCore(ctx context.Context, settings *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings) (*ServerApplicationCore, error) {
	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start basic server application: %w", err)
	}

	jwkGenService := serverApplicationBasic.JWKGenService

	serverApplicationCore := &ServerApplicationCore{}
	serverApplicationCore.ServerApplicationBasic = serverApplicationBasic
	serverApplicationCore.Settings = settings

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

	barrierService, err := cryptoutilBarrierService.NewService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, serverApplicationBasic.UnsealKeysService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	serverApplicationCore.BarrierService = barrierService

	businessLogicService, err := cryptoutilKmsServerBusinesslogic.NewBusinessLogicService(ctx, serverApplicationBasic.TelemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		serverApplicationBasic.TelemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		serverApplicationCore.Shutdown()

		return nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	serverApplicationCore.BusinessLogicService = businessLogicService

	// Seed or reset demo data if demo mode is enabled.
	if settings.DemoMode {
		serverApplicationBasic.TelemetryService.Slogger.Info("Demo mode enabled, seeding demo data")

		err = cryptoutilKmsServerDemo.SeedDemoData(ctx, serverApplicationBasic.TelemetryService, businessLogicService)
		if err != nil {
			serverApplicationBasic.TelemetryService.Slogger.Error("failed to seed demo data", "error", err)
			serverApplicationCore.Shutdown()

			return nil, fmt.Errorf("failed to seed demo data: %w", err)
		}
	} else if settings.ResetDemoMode {
		serverApplicationBasic.TelemetryService.Slogger.Info("Reset demo mode enabled, resetting demo data")

		err = cryptoutilKmsServerDemo.ResetDemoData(ctx, serverApplicationBasic.TelemetryService, businessLogicService)
		if err != nil {
			serverApplicationBasic.TelemetryService.Slogger.Error("failed to reset demo data", "error", err)
			serverApplicationCore.Shutdown()

			return nil, fmt.Errorf("failed to reset demo data: %w", err)
		}
	}

	return serverApplicationCore, nil
}

// Shutdown returns a shutdown function that gracefully stops all core application services.
func (c *ServerApplicationCore) Shutdown() func() {
	return func() {
		if c.ServerApplicationBasic != nil && c.ServerApplicationBasic.TelemetryService != nil {
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
