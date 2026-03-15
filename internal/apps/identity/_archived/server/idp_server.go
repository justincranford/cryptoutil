// Copyright (c) 2025 Justin Cranford
//
//

package server

import (
	"context"
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// IDPServer represents the OIDC identity provider server.
type IDPServer struct {
	config      *cryptoutilIdentityConfig.Config
	app         *fiber.App
	idpSvc      *cryptoutilIdentityIdp.Service
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc    *cryptoutilIdentityIssuer.TokenService
}

// NewIDPServer creates a new OIDC identity provider server.
func NewIDPServer(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *IDPServer {
	// Create Fiber app.
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cryptoutilSharedMagic.FiberReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cryptoutilSharedMagic.FiberWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cryptoutilSharedMagic.FiberIdleTimeoutSeconds) * time.Second,
	})

	// Create identity provider service.
	idpSvc := cryptoutilIdentityIdp.NewService(config, repoFactory, tokenSvc)

	// Register middleware and routes.
	idpSvc.RegisterMiddleware(app)
	idpSvc.RegisterRoutes(app)

	return &IDPServer{
		config:      config,
		app:         app,
		idpSvc:      idpSvc,
		repoFactory: repoFactory,
		tokenSvc:    tokenSvc,
	}
}

// Start starts the OIDC identity provider server.
func (s *IDPServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.IDP.BindAddress, s.config.IDP.Port)

	// Start identity provider service.
	if err := s.idpSvc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start identity provider service: %w", err)
	}

	// Start HTTP server.
	if s.config.IDP.TLSEnabled {
		if err := s.app.ListenTLS(addr, s.config.IDP.TLSCertFile, s.config.IDP.TLSKeyFile); err != nil {
			return fmt.Errorf("failed to start HTTPS server: %w", err)
		}

		return nil
	}

	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop stops the OIDC identity provider server.
func (s *IDPServer) Stop(ctx context.Context) error {
	// Stop identity provider service.
	if err := s.idpSvc.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop identity provider service: %w", err)
	}

	// Shutdown HTTP server.
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}
