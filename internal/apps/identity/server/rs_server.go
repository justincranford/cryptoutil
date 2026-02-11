// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"fmt"
	"log/slog"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRs "cryptoutil/internal/apps/identity/rs"
)

// RSServer wraps the Resource Server service with HTTP server lifecycle.
type RSServer struct {
	config     *cryptoutilIdentityConfig.Config
	logger     *slog.Logger
	fiberApp   *fiber.App
	service    *cryptoutilIdentityRs.Service
	shutdownCh chan struct{}
}

// NewRSServer creates a new Resource Server HTTP server.
func NewRSServer(_ context.Context, config *cryptoutilIdentityConfig.Config, logger *slog.Logger, tokenSvc cryptoutilIdentityRs.TokenService) (*RSServer, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	} else if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	} else if tokenSvc == nil {
		return nil, fmt.Errorf("tokenSvc cannot be nil")
	}

	// Create Fiber app for resource server.
	fiberApp := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           cryptoutilIdentityMagic.DefaultReadTimeout,
		WriteTimeout:          cryptoutilIdentityMagic.DefaultWriteTimeout,
		IdleTimeout:           cryptoutilIdentityMagic.DefaultIdleTimeout,
		ServerHeader:          "Resource-Server",
		AppName:               "Identity Resource Server",
	})

	// Create resource server service.
	service := cryptoutilIdentityRs.NewService(config, logger, tokenSvc)

	// Register middleware and routes.
	service.RegisterMiddleware(fiberApp)
	service.RegisterRoutes(fiberApp)

	return &RSServer{
		config:     config,
		logger:     logger,
		fiberApp:   fiberApp,
		service:    service,
		shutdownCh: make(chan struct{}),
	}, nil
}

// Start begins listening for HTTP requests.
func (s *RSServer) Start(_ context.Context) error {
	listenAddr := fmt.Sprintf("%s:%d", s.config.RS.BindAddress, s.config.RS.Port)

	s.logger.Info("Starting Resource Server",
		slog.String("address", listenAddr))

	if err := s.fiberApp.Listen(listenAddr); err != nil {
		return fmt.Errorf("resource server failed to start: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server.
func (s *RSServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Resource Server")

	if err := s.fiberApp.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("resource server shutdown failed: %w", err)
	}

	close(s.shutdownCh)
	s.logger.Info("Resource Server stopped")

	return nil
}

// Wait blocks until the server is shut down.
func (s *RSServer) Wait() {
	<-s.shutdownCh
}
