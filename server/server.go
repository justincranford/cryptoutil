package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cryptoutil/api/handlers"
	openapi2 "cryptoutil/api/openapi"
	"cryptoutil/orm"
	"cryptoutil/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	fibermiddleware "github.com/oapi-codegen/fiber-middleware"
)

func NewServer(listenAddress string, applyMigrations bool) (func(), func()) {
	ctx := context.Background()
	// ormService, err := orm.NewService(ctx, orm.DBTypePostgres, "", orm.ContainerModeRequired, applyMigrations)
	ormService, err := orm.NewService(ctx, orm.DBTypeSQLite, ":memory:", orm.ContainerModeDisabled, applyMigrations)
	if err != nil {
		log.Fatalf("open ORM service error: %v", err)
	}

	swaggerApi, err := openapi2.GetSwagger()
	if err != nil {
		ormService.Shutdown()
		log.Fatalf("get swagger error: %v", err)
	}

	app := fiber.New(fiber.Config{Immutable: true})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Get("/swagger/doc.json", openapi2.FiberHandlerOpenAPISpec())
	app.Get("/swagger/*", swagger.HandlerDefault)

	newVar := service.NewService(ormService)

	strictServer := handlers.NewStrictServer(newVar)
	openapi2.RegisterHandlersWithOptions(app, openapi2.NewStrictHandler(strictServer, nil), openapi2.FiberServerOptions{
		Middlewares: []openapi2.MiddlewareFunc{
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
		ormService.Shutdown()
	}()

	startServer := func() {
		err = app.Listen(listenAddress)
		if err != nil {
			fmt.Printf("Error starting server: %s", err)
		}
	}
	stopServer := func() {
		ormService.Shutdown()
		err := app.Shutdown()
		if err != nil {
			fmt.Printf("Error stopping server: %s", err)
		}
	}
	return startServer, stopServer
}
