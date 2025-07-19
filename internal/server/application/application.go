package application

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilUnsealKeysService "cryptoutil/internal/server/barrier/unsealkeysservice"
	cryptoutilBusinessLogic "cryptoutil/internal/server/businesslogic"
	cryptoutilOpenapiHandler "cryptoutil/internal/server/handler"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/server/repository/sqlrepository"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

var ready atomic.Bool

func StartServerApplication(settings *cryptoutilConfig.Settings, listenHost string, listenPort int, applyMigrations bool) (func(), func(), error) {
	ctx := context.Background()

	settings, err := cryptoutilConfig.Parse()
	if err != nil {
		log.Fatal("Error parsing config:", err)
	}

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings.OTLPScope, settings.OTLP, settings.OTLPConsole)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}

	jwkGenService, err := cryptoutilJose.NewJwkGenService(ctx, telemetryService)
	if err != nil {
		telemetryService.Slogger.Error("failed to create JWK Gen Service", "error", err)
		stopServerFunc(telemetryService, nil, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create JWK Gen Service: %w", err)
	}

	sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, cryptoutilSqlRepository.DBTypeSQLite, ":memory:", cryptoutilSqlRepository.ContainerModeDisabled)
	// sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, cryptoutilSqlRepository.DBTypePostgres, nil, cryptoutilSqlRepository.ContainerModeRequired)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunc(telemetryService, jwkGenService, nil, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, telemetryService, jwkGenService, sqlRepository, applyMigrations)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	unsealKeysService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		telemetryService.Slogger.Error("failed to create unseal repository", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, jwkGenService, ormRepository, unsealKeysService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, unsealKeysService, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	businessLogicService, err := cryptoutilBusinessLogic.NewBusinessLogicService(ctx, telemetryService, jwkGenService, ormRepository, barrierService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, unsealKeysService, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to get fiber handler for OpenAPI spec: %w", err)
	}

	app := fiber.New(fiber.Config{Immutable: true})
	app.Use(recover.New())
	app.Use(logger.New()) // TODO Remove this since it prints unstructured logs, and doesn't push to OpenTelemetry
	app.Use(fiberOtelLoggerMiddleware(telemetryService.Slogger))
	app.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(listenHost),
		otelfiber.WithPort(listenPort),
	))
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		if ready.Load() {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.SendStatus(fiber.StatusServiceUnavailable)
	})
	app.Get("/swagger/doc.json", fiberHandlerOpenAPISpec)
	app.Get("/swagger/*", swagger.HandlerDefault)

	openapiStrictServer := cryptoutilOpenapiHandler.NewOpenapiStrictServer(businessLogicService)
	openapiStrictHandler := cryptoutilOpenapiServer.NewStrictHandler(openapiStrictServer, nil)
	fiberServerOptions := cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{ // Defined as MiddlewareFunc => Fiber.Handler in generated code
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	}
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, openapiStrictHandler, fiberServerOptions)

	listenAddress := fmt.Sprintf("%s:%d", listenHost, listenPort)

	startServer := startServerFunc(err, listenAddress, app, telemetryService)
	stopServer := stopServerFunc(telemetryService, jwkGenService, sqlRepository, ormRepository, unsealKeysService, barrierService, app)
	go stopServerSignalFunc(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunc(err error, listenAddress string, app *fiber.App, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting fiber listener")
		ready.Store(true)
		err = app.Listen(listenAddress) // blocks until fiber app is stopped (e.g. stopServerFunc called by unit test or stopServerSignalFunc)
		if err != nil {
			telemetryService.Slogger.Error("failed to start fiber listener", "error", err)
		}
		telemetryService.Slogger.Debug("listener fiber stopped")
	}
}

func stopServerFunc(telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JwkGenService, sqlRepository *cryptoutilSqlRepository.SqlRepository, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealKeysService cryptoutilUnsealKeysService.UnsealKeysService, barrierService *cryptoutilBarrierService.BarrierService, app *fiber.App) func() {
	return func() {
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopping server")
		}
		if app != nil {
			err := app.Shutdown()
			if err != nil {
				telemetryService.Slogger.Error("failed to stop fiber server", "error", err)
			}
		}
		if barrierService != nil {
			barrierService.Shutdown() // does its own logging
		}
		if unsealKeysService != nil {
			unsealKeysService.Shutdown() // does its own logging
		}
		if ormRepository != nil {
			ormRepository.Shutdown() // does its own logging
		}
		if sqlRepository != nil {
			sqlRepository.Shutdown() // does its own logging
		}
		if jwkGenService != nil {
			jwkGenService.Shutdown() // does its own logging
		}
		if telemetryService != nil {
			telemetryService.Shutdown() // does its own logging
		}
	}
}

func stopServerSignalFunc(telemetryService *cryptoutilTelemetry.TelemetryService, stopServerFunc func()) func() {
	return func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		telemetryService.Slogger.Info("received stop server signal")
		stopServerFunc()
	}
}
