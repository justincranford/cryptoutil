package listener

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cryptoutilBusinessLogic "cryptoutil/internal/businesslogic"
	cryptoutilBarrierService "cryptoutil/internal/crypto/barrier/barrierservice"
	cryptoutilUnsealRepository "cryptoutil/internal/crypto/barrier/unsealrepository"
	cryptoutilUnsealService "cryptoutil/internal/crypto/barrier/unsealservice"
	cryptoutilOpenapiHandler "cryptoutil/internal/handler"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlRepository "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

func NewHttpListener(listenHost string, listenPort int, applyMigrations bool) (func(), func(), error) {
	ctx := context.Background()

	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, "cryptoutil", false, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}

	const dbType = cryptoutilSqlRepository.DBTypeSQLite // DBTypeSQLite or DBTypePostgres
	const dbUrl = ":memory:"                            // ":memory:" for SQLite, full URL for Postgres
	sqlRepository, err := cryptoutilSqlRepository.NewSqlRepository(ctx, telemetryService, dbType, dbUrl, cryptoutilSqlRepository.ContainerModeDisabled)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunc(telemetryService, nil, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	ormRepository, err := cryptoutilOrmRepository.NewOrmRepository(ctx, telemetryService, sqlRepository, applyMigrations)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunc(telemetryService, sqlRepository, nil, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	unsealRepository, err := cryptoutilUnsealRepository.NewUnsealRepositoryFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		stopServerFunc(telemetryService, sqlRepository, ormRepository, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	unsealService, err := cryptoutilUnsealService.NewUnsealService(telemetryService, ormRepository, unsealRepository)
	if err != nil {
		stopServerFunc(telemetryService, sqlRepository, ormRepository, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create unseal service: %w", err)
	}

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, ormRepository, unsealService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize barrier service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, ormRepository, nil, nil)()
		return nil, nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	businessLogicService, err := cryptoutilBusinessLogic.NewBusinessLogicService(ctx, telemetryService, ormRepository, barrierService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		stopServerFunc(telemetryService, sqlRepository, ormRepository, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunc(telemetryService, sqlRepository, ormRepository, barrierService, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunc(telemetryService, sqlRepository, ormRepository, barrierService, nil)()
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
	app.Get("/swagger/doc.json", fiberHandlerOpenAPISpec)
	app.Get("/swagger/*", swagger.HandlerDefault)

	openapiHandler := cryptoutilOpenapiHandler.NewOpenapiHandler(businessLogicService)
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, cryptoutilOpenapiServer.NewStrictHandler(openapiHandler, nil), cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	})

	listenAddress := fmt.Sprintf("%s:%d", listenHost, listenPort)

	startServer := startServerFunc(err, listenAddress, app, telemetryService)
	stopServer := stopServerFunc(telemetryService, sqlRepository, ormRepository, barrierService, app)
	go stopServerSignalFunc(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunc(err error, listenAddress string, app *fiber.App, telemetryService *cryptoutilTelemetry.TelemetryService) func() {
	return func() {
		telemetryService.Slogger.Debug("starting server")
		err = app.Listen(listenAddress) // blocks until fiber app is stopped (e.g. stopServerFunc called by unit test or stopServerSignalFunc)
		if err != nil {
			telemetryService.Slogger.Error("failed to start fiber server", "error", err)
		}
	}
}

func stopServerFunc(telemetryService *cryptoutilTelemetry.TelemetryService, sqlRepository *cryptoutilSqlRepository.SqlRepository, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService, app *fiber.App) func() {
	return func() {
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopping server")
		}
		if app != nil {
			if telemetryService != nil {
				telemetryService.Slogger.Debug("stopping fiber server")
			}
			err := app.Shutdown()
			if err != nil {
				telemetryService.Slogger.Error("failed to stop fiber server", "error", err)
			}
		}
		if barrierService != nil {
			barrierService.Shutdown() // does its own logging
		}
		if ormRepository != nil {
			ormRepository.Shutdown() // does its own logging
		}
		if sqlRepository != nil {
			sqlRepository.Shutdown() // does its own logging
		}
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopped server")
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
