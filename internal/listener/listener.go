package listener

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	cryptoutilOpenapiHandler "cryptoutil/internal/handler"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilSqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilServiceLogic "cryptoutil/internal/servicelogic"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

func NewListener(listenHost string, listenPort int, applyMigrations bool) (func(), func(), error) {
	ctx := context.Background()

	telemetryService := cryptoutilTelemetry.NewService(ctx, "cryptoutil", false, false)

	// const dbType = cryptoutilSqlProvider.SupportedSqlDBPostgres
	// const dbUrl = "?"
	const dbType = cryptoutilSqlProvider.SupportedSqlDBSQLite
	const dbUrl = ":memory:"
	sqlDB, shutdownDBContainer, err := cryptoutilSqlProvider.CreateSqlDB(ctx, telemetryService, dbType, dbUrl, cryptoutilSqlProvider.ContainerModeDisabled)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		stopServerFunction(telemetryService, shutdownDBContainer, nil, nil)()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	repositoryOrm, err := cryptoutilOrmRepository.NewRepositoryOrm(ctx, telemetryService, dbType, sqlDB, applyMigrations)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		stopServerFunction(telemetryService, shutdownDBContainer, repositoryOrm, nil)()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		stopServerFunction(telemetryService, shutdownDBContainer, repositoryOrm, nil)()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
	}

	fiberHandlerOpenAPISpec, err := cryptoutilOpenapiServer.FiberHandlerOpenAPISpec()
	if err != nil {
		telemetryService.Slogger.Error("failed to get fiber handler for OpenAPI spec", "error", err)
		stopServerFunction(telemetryService, shutdownDBContainer, repositoryOrm, nil)()
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

	openapiHandler := cryptoutilOpenapiHandler.NewOpenapiHandler(cryptoutilServiceLogic.NewService(repositoryOrm))
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, cryptoutilOpenapiServer.NewStrictHandler(openapiHandler, nil), cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	})

	listenAddress := fmt.Sprintf("%s:%d", listenHost, listenPort)

	startServer := startServerFunction(err, listenAddress, app, telemetryService)
	stopServer := stopServerFunction(telemetryService, shutdownDBContainer, repositoryOrm, app)
	go stopServerSignalFunction(telemetryService, stopServer)() // listen for OS signals to gracefully shutdown the server

	return startServer, stopServer, nil
}

func startServerFunction(err error, listenAddress string, app *fiber.App, telemetryService *cryptoutilTelemetry.Service) func() {
	startServer := func() {
		telemetryService.Slogger.Info("starting fiber server")
		err = app.Listen(listenAddress) // blocks until fiber app is stopped
		if err != nil {
			telemetryService.Slogger.Error("failed to start fiber server", "error", err)
		}
	}
	return startServer
}

func stopServerFunction(telemetryService *cryptoutilTelemetry.Service, shutdownDBContainer func(), repositoryOrm *cryptoutilOrmRepository.Repository, app *fiber.App) func() {
	stopServer := func() {
		if telemetryService != nil {
			telemetryService.Slogger.Error("stopping fiber server")
		}
		if app != nil {
			err := app.Shutdown()
			if err != nil {
				telemetryService.Slogger.Error("failed to stop fiber server", "error", err)
			}
		}
		if repositoryOrm != nil {
			repositoryOrm.Shutdown()
		}
		if shutdownDBContainer != nil {
			shutdownDBContainer()
		}
		if telemetryService != nil {
			telemetryService.Shutdown()
		}
	}
	return stopServer
}

func stopServerSignalFunction(telemetryService *cryptoutilTelemetry.Service, stopServer func()) func() {
	newVar := func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		telemetryService.Slogger.Info("gracefully stopped fiber server")
		stopServer()
	}
	return newVar
}
