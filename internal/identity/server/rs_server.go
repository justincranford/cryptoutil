package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
)

// RSServer represents the resource server.
type RSServer struct {
	config *cryptoutilIdentityConfig.Config
	app    *fiber.App
}

// NewRSServer creates a new resource server.
func NewRSServer(config *cryptoutilIdentityConfig.Config) *RSServer {
	// Create Fiber app.
	app := fiber.New(fiber.Config{
		ReadTimeout:  time.Duration(cryptoutilIdentityMagic.FiberReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cryptoutilIdentityMagic.FiberWriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cryptoutilIdentityMagic.FiberIdleTimeoutSeconds) * time.Second,
	})

	// Register routes.
	app.Get("/api/v1/protected", func(c *fiber.Ctx) error {
		// TODO: Validate access token.
		// TODO: Check scopes.
		return c.JSON(fiber.Map{
			"message": "Protected resource accessed successfully",
			"data":    "Sensitive information",
		})
	})

	app.Get("/api/v1/public", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Public resource accessed successfully",
			"data":    "Public information",
		})
	})

	return &RSServer{
		config: config,
		app:    app,
	}
}

// Start starts the resource server.
func (s *RSServer) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.config.RS.BindAddress, s.config.RS.Port)

	// Start HTTP server.
	if s.config.RS.TLSEnabled {
		if err := s.app.ListenTLS(addr, s.config.RS.TLSCertFile, s.config.RS.TLSKeyFile); err != nil {
			return fmt.Errorf("failed to start HTTPS server: %w", err)
		}

		return nil
	}

	if err := s.app.Listen(addr); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop stops the resource server.
func (s *RSServer) Stop(ctx context.Context) error {
	// Shutdown HTTP server.
	if err := s.app.ShutdownWithContext(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}
