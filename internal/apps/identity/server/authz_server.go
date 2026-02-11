// Copyright (c) 2025 Justin Cranford
//
//

// Package server provides identity server implementations for authorization and authentication.
package server

import (
	"context"
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// AuthZServer represents the OAuth 2.1 authorization server.
type AuthZServer struct {
	config      *cryptoutilIdentityConfig.Config
	app         *fiber.App
	authzSvc    *cryptoutilIdentityAuthz.Service
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory
	tokenSvc    *cryptoutilIdentityIssuer.TokenService
}

// NewAuthZServer creates a new OAuth 2.1 authorization server.
func NewAuthZServer(
	config *cryptoutilIdentityConfig.Config,
	repoFactory *cryptoutilIdentityRepository.RepositoryFactory,
	tokenSvc *cryptoutilIdentityIssuer.TokenService,
) *AuthZServer {
	// Create Fiber app.
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cryptoutilIdentityMagic.FiberReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cryptoutilIdentityMagic.FiberWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cryptoutilIdentityMagic.FiberIdleTimeoutSeconds) * time.Second,
	})

	// Create authorization service.
	authzSvc := cryptoutilIdentityAuthz.NewService(config, repoFactory, tokenSvc)

	// Register middleware and routes.
	authzSvc.RegisterMiddleware(app)
	authzSvc.RegisterRoutes(app)

	return &AuthZServer{
		config:      config,
		app:         app,
		authzSvc:    authzSvc,
		repoFactory: repoFactory,
		tokenSvc:    tokenSvc,
	}
}

// Start starts the OAuth 2.1 authorization server.
func (s *AuthZServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.AuthZ.BindAddress, s.config.AuthZ.Port)

	// Start authorization service.
	if err := s.authzSvc.Start(ctx); err != nil {
		return fmt.Errorf("failed to start authorization service: %w", err)
	}

	// Start HTTP server.
	if s.config.AuthZ.TLSEnabled {
		if err := s.app.ListenTLS(addr, s.config.AuthZ.TLSCertFile, s.config.AuthZ.TLSKeyFile); err != nil {
			return fmt.Errorf("failed to start HTTPS server: %w", err)
		}

		return nil
	}

	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop stops the OAuth 2.1 authorization server.
func (s *AuthZServer) Stop(ctx context.Context) error {
	// Stop authorization service.
	if err := s.authzSvc.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop authorization service: %w", err)
	}

	// Shutdown HTTP server.
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}
