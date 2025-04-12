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
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

func NewListener(listenHost string, listenPort int, applyMigrations bool) (func(), func(), error) {
	ctx := context.Background()

	telemetryService, err := cryptoutilTelemetry.NewService(ctx, "cryptoutil", false, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initailize telemetry: %w", err)
	}

	// const dbType = cryptoutilSqlProvider.SupportedSqlDBPostgres
	// const dbUrl = "?"
	const dbType = cryptoutilSqlProvider.DBTypeSQLite
	const dbUrl = ":memory:"
	sqlProvider, err := cryptoutilSqlProvider.NewSqlProvider(ctx, telemetryService, dbType, dbUrl, cryptoutilSqlProvider.ContainerModeDisabled)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunc(telemetryService, sqlProvider, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	repositoryOrm, err := cryptoutilOrmRepository.NewRepositoryOrm(ctx, telemetryService, sqlProvider, applyMigrations)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunc(telemetryService, sqlProvider, repositoryOrm, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	unsealRepository, err := cryptoutilUnsealRepository.NewUnsealRepositoryFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create unseal repository: %w", err)
	}

	unsealService, err := cryptoutilUnsealService.NewUnsealService(telemetryService, repositoryOrm, unsealRepository)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create unseal service: %w", err)
	}

	barrierService, err := cryptoutilBarrierService.NewBarrierService(ctx, telemetryService, repositoryOrm, unsealService)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create barrier service: %w", err)
	}

	businessLogicService, err := cryptoutilBusinessLogic.NewService(ctx, telemetryService, repositoryOrm, barrierService)
	if err != nil {
		telemetryService.Slogger.Error("failed to initialize business logic service", "error", err)
		stopServerFunc(telemetryService, sqlProvider, repositoryOrm, nil)()
		return nil, nil, fmt.Errorf("failed to initialize business logic service: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunc(telemetryService, sqlProvider, repositoryOrm, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunc(telemetryService, sqlProvider, repositoryOrm, nil)()
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
	stopServer := stopServerFunc(telemetryService, sqlProvider, repositoryOrm, app)
	go stopServerSignalFunc(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunc(err error, listenAddress string, app *fiber.App, telemetryService *cryptoutilTelemetry.Service) func() {
	return func() {
		telemetryService.Slogger.Debug("starting server")
		err = app.Listen(listenAddress) // blocks until fiber app is stopped (e.g. stopServerFunc called by unit test or stopServerSignalFunc)
		if err != nil {
			telemetryService.Slogger.Error("failed to start fiber server", "error", err)
		}
	}
}

func stopServerFunc(telemetryService *cryptoutilTelemetry.Service, sqlProvider *cryptoutilSqlProvider.SqlProvider, repositoryOrm *cryptoutilOrmRepository.RepositoryProvider, app *fiber.App) func() {
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
		if repositoryOrm != nil {
			repositoryOrm.Shutdown() // does its own logging
		}
		if sqlProvider != nil {
			sqlProvider.Shutdown() // does its own logging
		}
		if telemetryService != nil {
			telemetryService.Slogger.Debug("stopped server")
			telemetryService.Shutdown() // does its own logging
		}
	}
}

func stopServerSignalFunc(telemetryService *cryptoutilTelemetry.Service, stopServerFunc func()) func() {
	return func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		telemetryService.Slogger.Info("received stop server signal")
		stopServerFunc()
	}
}
