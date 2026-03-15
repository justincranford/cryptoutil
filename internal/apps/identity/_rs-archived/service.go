// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package rs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Service represents the resource server.
type Service struct {
	config    *cryptoutilIdentityConfig.Config
	logger    *slog.Logger
	tokenSvc  TokenService
	validator *TokenValidator
}

// NewService creates a new resource server service.
func NewService(
	config *cryptoutilIdentityConfig.Config,
	logger *slog.Logger,
	tokenSvc TokenService,
) *Service {
	validator := NewTokenValidator(tokenSvc, logger)

	return &Service{
		config:    config,
		logger:    logger,
		tokenSvc:  tokenSvc,
		validator: validator,
	}
}

// Start starts the resource server service.
func (s *Service) Start(_ context.Context) error {
	s.logger.Info("Resource server starting")

	return nil
}

// Stop stops the resource server service.
func (s *Service) Stop(_ context.Context) error {
	s.logger.Info("Resource server stopping")

	return nil
}

// RegisterMiddleware registers resource server middleware.
func (s *Service) RegisterMiddleware(app *fiber.App) {
	// CORS middleware for cross-origin requests.
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS" {
			return c.SendStatus(fiber.StatusOK)
		}

		return c.Next()
	})
}

// RegisterRoutes registers resource server routes.
func (s *Service) RegisterRoutes(app *fiber.App) {
	// Swagger UI OpenAPI spec endpoint.
	swaggerHandler, err := ServeOpenAPISpec()
	if err != nil {
		s.logger.Error("Failed to create Swagger UI handler", cryptoutilSharedMagic.StringError, err)
	} else {
		app.Get("/ui/swagger/doc.json", swaggerHandler)
	}

	// Health check endpoint (outside /api/v1 - per OpenAPI spec).
	app.Get("/health", s.handlePublicHealth)

	// Resource server API endpoints with /api/v1 prefix.
	api := app.Group("/api/v1")

	// Public endpoints (no authentication required).
	api.Get("/public/health", s.handlePublicHealth)

	// Protected endpoints (require valid access token).
	protected := api.Group("/protected")
	protected.Use(s.validator.ValidateToken())
	protected.Get("/resource", s.RequireScopes("read:resource"), s.handleProtectedResource)
	protected.Post("/resource", s.RequireScopes("write:resource"), s.handleCreateResource)
	protected.Delete("/resource/:id", s.RequireScopes("delete:resource"), s.handleDeleteResource)

	// Admin endpoints (require admin scope).
	admin := api.Group("/admin")
	admin.Use(s.validator.ValidateToken())
	admin.Use(s.RequireScopes("admin"))
	admin.Get("/users", s.handleAdminUsers)
	admin.Get("/metrics", s.handleAdminMetrics)
}

// RequireScopes creates middleware that enforces required scopes.
func (s *Service) RequireScopes(requiredScopes ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token claims from context (set by ValidateToken middleware).
		claims, ok := c.Locals("token_claims").(map[string]any)
		if !ok {
			s.logger.Warn("Token validation middleware not executed",
				"path", c.Path(),
				"method", c.Method())

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidToken,
				"error_description":               "Token validation required",
			})
		}

		// Extract scopes from claims.
		scopeStr, ok := claims[cryptoutilSharedMagic.ClaimScope].(string)
		if !ok {
			s.logger.Warn("Missing scope claim in token",
				cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID],
				"path", c.Path())

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInsufficientScope,
				"error_description":               "Token missing scope claim",
			})
		}

		// Parse scopes.
		grantedScopes := strings.Split(scopeStr, " ")
		scopeSet := make(map[string]bool)

		for _, scope := range grantedScopes {
			scopeSet[scope] = true
		}

		// Check if all required scopes are present.
		for _, required := range requiredScopes {
			if !scopeSet[required] {
				s.logger.Warn("Insufficient scope for resource access",
					cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID],
					"required_scopes", requiredScopes,
					"granted_scopes", grantedScopes,
					"path", c.Path())

				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInsufficientScope,
					"error_description":               fmt.Sprintf("Required scope missing: %s", required),
				})
			}
		}

		s.logger.Debug("Scope validation successful",
			cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID],
			"required_scopes", requiredScopes,
			"path", c.Path())

		return c.Next()
	}
}

// handlePublicHealth handles GET /api/v1/public/health - public health check.
func (s *Service) handlePublicHealth(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		cryptoutilSharedMagic.StringStatus: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		"service":                          "resource-server",
		"version":                          cryptoutilSharedMagic.ServiceVersion,
	})
}

// handleProtectedResource handles GET /api/v1/protected/resource - protected resource access.
func (s *Service) handleProtectedResource(c *fiber.Ctx) error {
	claims, ok := c.Locals("token_claims").(map[string]any)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid token claims format",
		})
	}

	s.logger.Info("Protected resource accessed",
		cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID],
		cryptoutilSharedMagic.ClaimScope, claims[cryptoutilSharedMagic.ClaimScope])

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":                           "Protected resource accessed successfully",
		cryptoutilSharedMagic.ClaimClientID: claims[cryptoutilSharedMagic.ClaimClientID],
		cryptoutilSharedMagic.ClaimScope:    claims[cryptoutilSharedMagic.ClaimScope],
		"data": fiber.Map{
			"id":                            "resource-123",
			cryptoutilSharedMagic.ClaimName: "Sample Protected Resource",
			"type":                          "example",
		},
	})
}

// handleCreateResource handles POST /api/v1/protected/resource - create resource.
func (s *Service) handleCreateResource(c *fiber.Ctx) error {
	claims, ok := c.Locals("token_claims").(map[string]any)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid token claims format",
		})
	}

	s.logger.Info("Resource created",
		cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID])

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Resource created successfully",
		"resource_id": "new-resource-456",
	})
}

// handleDeleteResource handles DELETE /api/v1/protected/resource - delete resource.
func (s *Service) handleDeleteResource(c *fiber.Ctx) error {
	claims, ok := c.Locals("token_claims").(map[string]any)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid token claims format",
		})
	}

	s.logger.Info("Resource deleted",
		cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID])

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Resource deleted successfully",
	})
}

// handleAdminUsers handles GET /api/v1/admin/users - admin user management.
func (s *Service) handleAdminUsers(c *fiber.Ctx) error {
	claims, ok := c.Locals("token_claims").(map[string]any)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid token claims format",
		})
	}

	s.logger.Info("Admin users accessed",
		cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID])

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Admin user list",
		"users": []fiber.Map{
			{"id": "user-1", "username": "alice", "role": "admin"},
			{"id": "user-2", "username": "bob", "role": "user"},
		},
	})
}

// handleAdminMetrics handles GET /api/v1/admin/metrics - admin metrics.
func (s *Service) handleAdminMetrics(c *fiber.Ctx) error {
	claims, ok := c.Locals("token_claims").(map[string]any)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Invalid token claims format",
		})
	}

	s.logger.Info("Admin metrics accessed",
		cryptoutilSharedMagic.ClaimClientID, claims[cryptoutilSharedMagic.ClaimClientID])

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "System metrics",
		"metrics": fiber.Map{
			"requests_total":   cryptoutilSharedMagic.ExampleMetricRequestsTotal,
			"requests_success": cryptoutilSharedMagic.ExampleMetricRequestsSuccess,
			"requests_failed":  cryptoutilSharedMagic.ExampleMetricRequestsFailed,
		},
	})
}
