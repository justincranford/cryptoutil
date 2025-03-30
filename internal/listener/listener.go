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

	// tracer := telemetryService.TracesProvider.Tracer("fiber-tracer")
	// _, span := tracer.Start(context.Background(), "test-span")
	// fmt.Println(span.SpanContext().TraceID())
	// fmt.Println(span.SpanContext().SpanID())

	// const dbType = cryptoutilSqlProvider.SupportedSqlDBPostgres
	// const dbUrl = "?"
	const dbType = cryptoutilSqlProvider.SupportedSqlDBSQLite
	const dbUrl = ":memory:"
	sqlDB, shutdownDBContainer, err := cryptoutilSqlProvider.CreateSqlDB(ctx, dbType, dbUrl, cryptoutilSqlProvider.ContainerModeDisabled)
	if err != nil {
		telemetryService.Slogger.Error("failed to connect to SQL DB", "error", err)
		telemetryService.Shutdown()
		return nil, nil, fmt.Errorf("failed to connect to SQL DB: %w", err)
	}

	repositoryOrm, err := cryptoutilOrmRepository.NewRepositoryOrm(ctx, dbType, sqlDB, applyMigrations)
	if err != nil {
		telemetryService.Slogger.Error("failed to create ORM repository", "error", err)
		shutdownDBContainer()
		telemetryService.Shutdown()
		return nil, nil, fmt.Errorf("failed to create ORM repository: %w", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		telemetryService.Slogger.Error("failed to get swagger", "error", err)
		repositoryOrm.Shutdown()
		shutdownDBContainer()
		telemetryService.Shutdown()
		return nil, nil, fmt.Errorf("failed to get swagger: %w", err)
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
	app.Get("/swagger/doc.json", cryptoutilOpenapiServer.FiberHandlerOpenAPISpec())
	app.Get("/swagger/*", swagger.HandlerDefault)

	openapiHandler := cryptoutilOpenapiHandler.NewOpenapiHandler(cryptoutilServiceLogic.NewService(repositoryOrm))
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, cryptoutilOpenapiServer.NewStrictHandler(openapiHandler, nil), cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	})

	listenAddress := fmt.Sprintf("%s:%d", listenHost, listenPort)

	startServer := func() {
		err = app.Listen(listenAddress)
		if err != nil {
			fmt.Printf("Error starting fiber server: %s", err)
		}
	}

	stopServer := func() {
		err := app.Shutdown()
		if err != nil {
			fmt.Printf("Error stopping fiber server: %s", err)
		}
		repositoryOrm.Shutdown()
		shutdownDBContainer()
		telemetryService.Shutdown()
	}

	// listen for OS signals to gracefully shutdown the server
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Printf("Fiber is gracefully shutting down...")
		stopServer()
	}()

	return startServer, stopServer, nil
}
