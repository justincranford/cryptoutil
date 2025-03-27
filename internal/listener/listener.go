package listener

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	cryptoutilBusinessLogic "cryptoutil/internal/businesslogic"
	cryptoutilOpenapiHandler "cryptoutil/internal/handler"
	cryptoutilOpenapiServer "cryptoutil/internal/openapi/server"
	cryptoutilRepositoryOrm "cryptoutil/internal/repository/orm"
	cryptoutilRepositorySqlProvider "cryptoutil/internal/repository/sqlprovider"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
	"go.opentelemetry.io/otel/trace"
)

func NewListener(listenHost string, listenPort int, applyMigrations bool) (func(), func()) {
	ctx := context.Background()

	telemetryService := cryptoutilTelemetry.NewService(ctx, "asn1_test", false, false)

	// tracer := telemetryService.TracesProvider.Tracer("fiber-tracer")
	// _, span := tracer.Start(context.Background(), "test-span")
	// fmt.Println(span.SpanContext().TraceID())
	// fmt.Println(span.SpanContext().SpanID())

	// repositoryOrm, err := orm.NewService(ctx, orm.SupportedSqlDBPostgres, "", orm.ContainerModeRequired, applyMigrations)
	repositoryOrm, err := cryptoutilRepositoryOrm.NewRepositoryOrm(ctx, cryptoutilRepositorySqlProvider.SupportedSqlDBSQLite, ":memory:", cryptoutilRepositorySqlProvider.ContainerModeDisabled, applyMigrations)
	if err != nil {
		log.Fatalf("open ORM service error: %v", err)
	}

	swaggerApi, err := cryptoutilOpenapiServer.GetSwagger()
	if err != nil {
		repositoryOrm.Shutdown()
		log.Fatalf("get swagger error: %v", err)
	}

	app := fiber.New(fiber.Config{Immutable: true})
	app.Use(recover.New())
	app.Use(logger.New()) // TODO Remove this since it prints unstructured logs, and doesn't push to OpenTelemetry
	app.Use(otelLoggerMiddleware(telemetryService.Slogger))
	app.Use(otelfiber.Middleware(
		otelfiber.WithTracerProvider(telemetryService.TracesProvider),
		otelfiber.WithMeterProvider(telemetryService.MetricsProvider),
		otelfiber.WithPropagators(*telemetryService.TextMapPropagator),
		otelfiber.WithServerName(listenHost),
		otelfiber.WithPort(listenPort),
	))
	app.Get("/swagger/doc.json", cryptoutilOpenapiServer.FiberHandlerOpenAPISpec())
	app.Get("/swagger/*", swagger.HandlerDefault)

	openapiHandler := cryptoutilOpenapiHandler.NewOpenapiHandler(cryptoutilBusinessLogic.NewService(repositoryOrm))
	cryptoutilOpenapiServer.RegisterHandlersWithOptions(app, cryptoutilOpenapiServer.NewStrictHandler(openapiHandler, nil), cryptoutilOpenapiServer.FiberServerOptions{
		Middlewares: []cryptoutilOpenapiServer.MiddlewareFunc{
			fibermiddleware.OapiRequestValidatorWithOptions(swaggerApi, &fibermiddleware.Options{}),
		},
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Printf("Fiber is gracefully shutting down...")
		if err := app.Shutdown(); err != nil {
			fmt.Printf("Fiber graceful shutdown error: %v", err)
		}
		repositoryOrm.Shutdown()
	}()

	listenAddress := fmt.Sprintf("%s:%d", listenHost, listenPort)
	startServer := func() {
		err = app.Listen(listenAddress)
		if err != nil {
			fmt.Printf("Error starting fiber server: %s", err)
		}
	}
	stopServer := func() {
		repositoryOrm.Shutdown()
		err := app.Shutdown()
		if err != nil {
			fmt.Printf("Error stopping fiber server: %s", err)
		}
		telemetryService.Shutdown()
	}
	return startServer, stopServer
}

func otelLoggerMiddleware(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		// Extract tracing information
		span := trace.SpanFromContext(c.Context())
		spanContext := span.SpanContext()

		// Log request details with OpenTelemetry correlation
		logger.Info("Responded",
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration_ms", time.Since(start)),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
		return err
	}
}
