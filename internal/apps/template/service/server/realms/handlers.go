// Copyright (c) 2025 Justin Cranford

package realms

import (
	"context"
	"fmt"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// realmServiceProvider is an optional interface that session managers can implement
// to provide realm lookup functionality for multi-tenant deployments.
type realmServiceProvider interface {
	GetRealmService() realmLookup
}

// realmLookup is a minimal interface for looking up the first active realm for a tenant.
// This avoids tight coupling to the full RealmService interface.
type realmLookup interface {
	GetFirstActiveRealm(ctx context.Context, tenantID googleUuid.UUID) (any, error)
}

// realmIDGetter is implemented by realm objects that expose their realm ID.
type realmIDGetter interface {
	GetRealmID() googleUuid.UUID
}

// HandleRegisterUser returns a Fiber handler for user registration.
//
// Workflow:
// 1. Parse request body (username, password)
// 2. Validate input (non-empty, length requirements)
// 3. Call service.RegisterUser (business logic)
// 4. Return created user ID and username
//
// Request Body:
//
//	{
//	  "username": "alice",
//	  "password": "securePassword123"
//	}
//
// Success Response (201 Created):
//
//	{
//	  "user_id": "uuid-string",
//	  "username": "alice"
//	}
//
// Error Responses:
// - 400 Bad Request: Invalid request body, missing fields, validation failure
// - 409 Conflict: Username already exists
// - 500 Internal Server Error: Database errors, password hashing failure.
func (s *UserServiceImpl) HandleRegisterUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Validate username length (3-50 characters).
		if len(req.Username) < cryptoutilSharedMagic.CipherMinUsernameLength ||
			len(req.Username) > cryptoutilSharedMagic.CipherMaxUsernameLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "username must be 3-50 characters",
			})
		}

		// Validate password length (minimum 8 characters).
		if len(req.Password) < cryptoutilSharedMagic.CipherMinPasswordLength {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "password must be at least 8 characters",
			})
		}

		// Call service layer (business logic).
		user, err := s.RegisterUser(c.Context(), req.Username, req.Password)
		if err != nil {
			// Check for duplicate username (conflict).
			// Note: Service layer returns generic error for security.
			// Repository constraint violations bubble up here.
			if err.Error() == "failed to create user: UNIQUE constraint failed: users.username" ||
				err.Error() == "failed to create user: duplicate key value violates unique constraint \"users_username_key\"" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "Username already exists",
				})
			}

			// Generic error for other failures.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create user",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"user_id":  user.GetID().String(),
			"username": user.GetUsername(),
		})
	}
}

// HandleLoginUser returns a Fiber handler for user login.
//
// Workflow:
// 1. Parse request body (username, password)
// 2. Validate input (non-empty)
// 3. Call service.AuthenticateUser (business logic)
// 4. Generate JWT token (15-minute expiration)
// 5. Return token and expiration time
//
// Request Body:
//
//	{
//	  "username": "alice",
//	  "password": "securePassword123"
//	}
//
// Success Response (200 OK):
//
//	{
//	  "token": "jwt-token-string",
//	  "expires_at": "2025-01-02T15:04:05Z"
//	}
//
// Error Responses:
// - 400 Bad Request: Invalid request body, missing fields
// - 401 Unauthorized: Invalid credentials (user not found or wrong password)
// - 500 Internal Server Error: Database errors, JWT generation failure.
func (s *UserServiceImpl) HandleLoginUser(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Call service layer (business logic).
		user, err := s.AuthenticateUser(c.Context(), req.Username, req.Password)
		if err != nil {
			// Service layer returns generic error for security.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Generate JWT token.
		token, expiresAt, err := GenerateJWT(user.GetID(), user.GetUsername(), jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.JSON(fiber.Map{
			"token":      token,
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// HandleLoginUserWithSession returns a Fiber handler for user login that issues session tokens.
//
// Workflow:
// 1. Parse request body (username, password)
// 2. Validate input (non-empty)
// 3. Call service.AuthenticateUser (business logic)
// 4. Issue session token via SessionManager (configurable: JWE, JWS, or OPAQUE)
// 5. Return session token and expiration time
//
// Request Body:
//
//	{
//	  "username": "alice",
//	  "password": "securePassword123"
//	}
//
// Success Response (200 OK):
//
//	{
//	  "token": "session-token-string",
//	  "expires_at": "2025-01-02T15:04:05Z"
//	}
//
// Error Responses:
// - 400 Bad Request: Invalid request body, missing fields
// - 401 Unauthorized: Invalid credentials (user not found or wrong password)
// - 500 Internal Server Error: Database errors, session generation failure
//
// Parameters:
// - sessionManager: Service providing session token issuance (must not be nil)
// - isBrowser: true for browser sessions, false for service sessions.
func (s *UserServiceImpl) HandleLoginUserWithSession(sessionManager any, isBrowser bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Username and password are required",
			})
		}

		// Call service layer (business logic).
		user, err := s.AuthenticateUser(c.Context(), req.Username, req.Password)
		if err != nil {
			// Service layer returns generic error for security.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		// Issue session token via SessionManager using type assertion.
		// SessionManager interface is defined in the specific service implementation.
		// The multi-tenant version requires tenantID and realmID parameters.
		type sessionIssuer interface {
			IssueBrowserSessionWithTenant(ctx context.Context, userID string, tenantID googleUuid.UUID, realmID googleUuid.UUID) (string, error)
			IssueServiceSessionWithTenant(ctx context.Context, clientID string, tenantID googleUuid.UUID, realmID googleUuid.UUID) (string, error)
		}

		manager, ok := sessionManager.(sessionIssuer)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Invalid session manager implementation",
			})
		}

		var (
			token    string
			issueErr error
		)

		// Extract tenant ID from authenticated user model.
		// For services using dynamic tenant creation (cipher-im), this is populated from user.TenantID.
		// For multi-tenant deployments, the realm lookup would query tenant_realms table.
		type tenantAware interface {
			GetTenantID() googleUuid.UUID
		}

		var tenantID googleUuid.UUID
		if templateUser, ok := user.(tenantAware); ok {
			tenantID = templateUser.GetTenantID()
		} else {
			// Fallback for user models without TenantID exposure.
			// This should not happen in practice since template.User implements this.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "User model does not expose tenant ID",
			})
		}

		// Realm lookup for multi-tenant deployments.
		// Try to extract RealmService from session manager using optional interface.
		// If available, use GetFirstActiveRealm() to find the tenant's default realm.
		// If no RealmService or no active realm exists, use zero UUID (graceful fallback).
		var realmID googleUuid.UUID

		if provider, ok := sessionManager.(realmServiceProvider); ok {
			if realmSvc := provider.GetRealmService(); realmSvc != nil {
				realm, err := realmSvc.GetFirstActiveRealm(c.Context(), tenantID)
				if err != nil {
					// Log error but continue with zero UUID fallback
					// This maintains availability even if realm lookup fails
					_ = err // Explicit ignore for linting
				} else if realm != nil {
					// Extract realm ID from returned realm object
					if r, ok := realm.(realmIDGetter); ok {
						realmID = r.GetRealmID()
					}
				}
			}
		}
		// If no realm found or RealmService not available, realmID remains zero UUID (backward compatible)

		if isBrowser {
			token, issueErr = manager.IssueBrowserSessionWithTenant(c.Context(), user.GetID().String(), tenantID, realmID)
		} else {
			token, issueErr = manager.IssueServiceSessionWithTenant(c.Context(), user.GetID().String(), tenantID, realmID)
		}

		if issueErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate session token",
			})
		}

		// Session expiration is handled by SessionManager configuration.
		// For compatibility, return current time + configured session expiration.
		expiresAt := time.Now().UTC().Add(cryptoutilSharedMagic.DefaultCompatibilitySessionExpiration)

		return c.JSON(fiber.Map{
			"token":      token,
			"expires_at": expiresAt.Format(time.RFC3339),
		})
	}
}

// GenerateJWT creates a new JWT token for the given user.
//
// Parameters:
// - userID: User's unique identifier (UUIDv7)
// - username: User's username
// - secret: JWT signing secret (HMAC-SHA256)
//
// Returns:
// - token: JWT token string
// - expiresAt: Token expiration time (15 minutes from now)
// - error: JWT generation failure
//
// Security Notes:
// - Algorithm: HMAC-SHA256 (HS256) - symmetric key signing
// - Expiration: 15 minutes (configurable via cryptoutilMagic.CipherJWTExpiration)
// - Issuer: "cipher-im" (configurable via cryptoutilMagic.CipherJWTIssuer)
// - Claims: user_id (string), username (string), iat, exp, iss.
func GenerateJWT(userID googleUuid.UUID, username, secret string) (string, time.Time, error) {
	expirationTime := time.Now().UTC().Add(cryptoutilSharedMagic.CipherJWTExpiration)
	claims := &Claims{
		UserID:   userID.String(),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    cryptoutilSharedMagic.CipherJWTIssuer,
		},
	}

	// Create and sign token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, expirationTime, nil
}
