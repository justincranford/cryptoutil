// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"context"
	"errors"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
)

// SessionContextKey is the context key for storing session information.
type SessionContextKey struct{}

// SessionInfo contains validated session information extracted from cookies or tokens.
type SessionInfo struct {
	SessionID googleUuid.UUID
	UserID    googleUuid.UUID
	TenantID  googleUuid.UUID
	RealmID   googleUuid.UUID
	Scopes    []string
	IssuedAt  int64
	ExpiresAt int64
}

// SessionValidator defines the interface for validating session tokens.
// Implementations must provide methods to validate browser and service sessions.
type SessionValidator interface {
	// ValidateBrowserSession validates a browser session token (cookie-based).
	ValidateBrowserSession(ctx context.Context, token string) (*SessionInfo, error)
	// ValidateServiceSession validates a service session token (header-based).
	ValidateServiceSession(ctx context.Context, token string) (*SessionInfo, error)
}

// GetSessionInfo retrieves session information from the context.
// Returns nil if no session info is present in the context.
func GetSessionInfo(ctx context.Context) *SessionInfo {
	if ctx == nil {
		return nil
	}

	if info, ok := ctx.Value(SessionContextKey{}).(*SessionInfo); ok {
		return info
	}

	return nil
}

// SessionMiddleware creates middleware that validates session tokens.
// For browser requests, it extracts the session token from cookies.
// For service requests, it extracts the token from the Authorization header.
// The validated session info is stored in the request context.
func SessionMiddleware(validator SessionValidator, cookieName string, isBrowser bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		if isBrowser {
			token = c.Cookies(cookieName)
			if token == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized",
					"message": "Missing session cookie",
				})
			}
		} else {
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized",
					"message": "Missing Authorization header",
				})
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized",
					"message": "Invalid Authorization header format",
				})
			}

			token = parts[1]
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Empty session token",
			})
		}

		ctx := c.UserContext()

		var (
			sessionInfo *SessionInfo
			err         error
		)

		if isBrowser {
			sessionInfo, err = validator.ValidateBrowserSession(ctx, token)
		} else {
			sessionInfo, err = validator.ValidateServiceSession(ctx, token)
		}

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid or expired session",
			})
		}

		ctx = context.WithValue(ctx, SessionContextKey{}, sessionInfo)

		realmCtx := &RealmContext{
			TenantID: sessionInfo.TenantID,
			RealmID:  sessionInfo.RealmID,
			UserID:   sessionInfo.UserID,
			Scopes:   sessionInfo.Scopes,
			Source:   "session",
		}
		ctx = context.WithValue(ctx, RealmContextKey{}, realmCtx)
		ctx = context.WithValue(ctx, TenantContextKey{}, sessionInfo.TenantID.String())

		c.SetUserContext(ctx)

		return c.Next()
	}
}

// BrowserSessionMiddleware creates middleware for validating browser session cookies.
func BrowserSessionMiddleware(validator SessionValidator, cookieName string) fiber.Handler {
	return SessionMiddleware(validator, cookieName, true)
}

// ServiceSessionMiddleware creates middleware for validating service session tokens.
func ServiceSessionMiddleware(validator SessionValidator) fiber.Handler {
	return SessionMiddleware(validator, "", false)
}

// RequireSessionMiddleware returns middleware that enforces session presence.
func RequireSessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionInfo := GetSessionInfo(c.UserContext())
		if sessionInfo == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Session required",
			})
		}

		return c.Next()
	}
}

// NoopSessionValidator is a session validator that always fails.
type NoopSessionValidator struct{}

// ValidateBrowserSession always returns an error.
func (n *NoopSessionValidator) ValidateBrowserSession(_ context.Context, _ string) (*SessionInfo, error) {
	return nil, errors.New("no session validator configured")
}

// ValidateServiceSession always returns an error.
func (n *NoopSessionValidator) ValidateServiceSession(_ context.Context, _ string) (*SessionInfo, error) {
	return nil, errors.New("no session validator configured")
}
