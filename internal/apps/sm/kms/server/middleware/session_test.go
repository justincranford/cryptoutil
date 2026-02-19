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
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockSessionValidator is a test implementation of SessionValidator.
type mockSessionValidator struct {
	browserSession *SessionInfo
	browserErr     error
	serviceSession *SessionInfo
	serviceErr     error
}

func (m *mockSessionValidator) ValidateBrowserSession(_ context.Context, _ string) (*SessionInfo, error) {
	return m.browserSession, m.browserErr
}

func (m *mockSessionValidator) ValidateServiceSession(_ context.Context, _ string) (*SessionInfo, error) {
	return m.serviceSession, m.serviceErr
}

func TestGetSessionInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() context.Context
		wantNil bool
	}{
		{
			name: "nil context",
			setup: func() context.Context {
				return nil
			},
			wantNil: true,
		},
		{
			name: "context without session",
			setup: func() context.Context {
				return context.Background()
			},
			wantNil: true,
		},
		{
			name: "context with session",
			setup: func() context.Context {
				info := &SessionInfo{
					SessionID: googleUuid.New(),
					TenantID:  googleUuid.New(),
				}

				return context.WithValue(context.Background(), SessionContextKey{}, info)
			},
			wantNil: false,
		},
		{
			name: "context with wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), SessionContextKey{}, "not a session")
			},
			wantNil: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := tc.setup()

			result := GetSessionInfo(ctx)
			if tc.wantNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

func TestBrowserSessionMiddleware(t *testing.T) {
	t.Parallel()

	validSession := &SessionInfo{
		SessionID: googleUuid.New(),
		UserID:    googleUuid.New(),
		TenantID:  googleUuid.New(),
		RealmID:   googleUuid.New(),
		Scopes:    []string{"read", "write"},
		IssuedAt:  1000,
		ExpiresAt: 2000,
	}

	tests := []struct {
		name           string
		cookieValue    string
		validator      *mockSessionValidator
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:        "missing cookie",
			cookieValue: "",
			validator: &mockSessionValidator{
				browserSession: validSession,
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Missing session cookie",
		},
		{
			name:        "invalid session",
			cookieValue: "invalid-token",
			validator: &mockSessionValidator{
				browserErr: errors.New("invalid session"),
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Invalid or expired session",
		},
		{
			name:        "valid session",
			cookieValue: "valid-token",
			validator: &mockSessionValidator{
				browserSession: validSession,
			},
			expectedStatus: fiber.StatusOK,
			expectedMsg:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(BrowserSessionMiddleware(tc.validator, "session"))
			app.Get("/test", func(c *fiber.Ctx) error {
				info := GetSessionInfo(c.UserContext())
				if info != nil {
					return c.SendString("ok")
				}

				return c.SendString("no session")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.cookieValue != "" {
				req.AddCookie(&http.Cookie{Name: "session", Value: tc.cookieValue})
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedMsg != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), tc.expectedMsg)
			}
		})
	}
}

func TestServiceSessionMiddleware(t *testing.T) {
	t.Parallel()

	validSession := &SessionInfo{
		SessionID: googleUuid.New(),
		UserID:    googleUuid.New(),
		TenantID:  googleUuid.New(),
		RealmID:   googleUuid.New(),
		Scopes:    []string{"admin"},
		IssuedAt:  1000,
		ExpiresAt: 2000,
	}

	tests := []struct {
		name           string
		authHeader     string
		validator      *mockSessionValidator
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			validator: &mockSessionValidator{
				serviceSession: validSession,
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Missing Authorization header",
		},
		{
			name:       "invalid header format - no bearer",
			authHeader: "Basic token123",
			validator: &mockSessionValidator{
				serviceSession: validSession,
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Invalid Authorization header format",
		},
		{
			name:       "invalid header format - no space",
			authHeader: "Bearertoken123",
			validator: &mockSessionValidator{
				serviceSession: validSession,
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Invalid Authorization header format",
		},
		{
			name:       "invalid session",
			authHeader: "Bearer invalid-token",
			validator: &mockSessionValidator{
				serviceErr: errors.New("invalid session"),
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Invalid or expired session",
		},
		{
			name:       "valid session - lowercase bearer",
			authHeader: "bearer valid-token",
			validator: &mockSessionValidator{
				serviceSession: validSession,
			},
			expectedStatus: fiber.StatusOK,
			expectedMsg:    "",
		},
		{
			name:       "valid session - uppercase bearer",
			authHeader: "Bearer valid-token",
			validator: &mockSessionValidator{
				serviceSession: validSession,
			},
			expectedStatus: fiber.StatusOK,
			expectedMsg:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(ServiceSessionMiddleware(tc.validator))
			app.Get("/test", func(c *fiber.Ctx) error {
				info := GetSessionInfo(c.UserContext())
				if info != nil {
					return c.SendString("ok")
				}

				return c.SendString("no session")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedMsg != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), tc.expectedMsg)
			}
		})
	}
}

func TestRequireSessionMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupContext   func(c *fiber.Ctx)
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "no session in context",
			setupContext: func(_ *fiber.Ctx) {
				// No setup - context has no session
			},
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "Session required",
		},
		{
			name: "session present in context",
			setupContext: func(c *fiber.Ctx) {
				info := &SessionInfo{
					SessionID: googleUuid.New(),
					TenantID:  googleUuid.New(),
				}
				ctx := context.WithValue(c.UserContext(), SessionContextKey{}, info)
				c.SetUserContext(ctx)
			},
			expectedStatus: fiber.StatusOK,
			expectedMsg:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			// Setup middleware that sets context before RequireSessionMiddleware
			app.Use(func(c *fiber.Ctx) error {
				tc.setupContext(c)

				return c.Next()
			})
			app.Use(RequireSessionMiddleware())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedMsg != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), tc.expectedMsg)
			}
		})
	}
}

func TestNoopSessionValidator(t *testing.T) {
	t.Parallel()

	validator := &NoopSessionValidator{}

	t.Run("browser session always fails", func(t *testing.T) {
		t.Parallel()

		info, err := validator.ValidateBrowserSession(context.Background(), "any-token")
		require.Error(t, err)
		require.Nil(t, info)
		require.Contains(t, err.Error(), "no session validator configured")
	})

	t.Run("service session always fails", func(t *testing.T) {
		t.Parallel()

		info, err := validator.ValidateServiceSession(context.Background(), "any-token")
		require.Error(t, err)
		require.Nil(t, info)
		require.Contains(t, err.Error(), "no session validator configured")
	})
}

func TestSessionMiddleware_SetsRealmContext(t *testing.T) {
	t.Parallel()

	validSession := &SessionInfo{
		SessionID: googleUuid.New(),
		UserID:    googleUuid.New(),
		TenantID:  googleUuid.New(),
		RealmID:   googleUuid.New(),
		Scopes:    []string{"read", "write"},
		IssuedAt:  1000,
		ExpiresAt: 2000,
	}

	validator := &mockSessionValidator{
		browserSession: validSession,
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(BrowserSessionMiddleware(validator, "session"))
	app.Get("/test", func(c *fiber.Ctx) error {
		// Verify SessionInfo is set
		sessionInfo := GetSessionInfo(c.UserContext())
		if sessionInfo == nil {
			return c.Status(500).SendString("no session info")
		}

		// Verify RealmContext is set
		realmCtx := GetRealmContext(c.UserContext())
		if realmCtx == nil {
			return c.Status(500).SendString("no realm context")
		}

		// Verify values match
		if realmCtx.TenantID != sessionInfo.TenantID {
			return c.Status(500).SendString("tenant mismatch")
		}

		if realmCtx.RealmID != sessionInfo.RealmID {
			return c.Status(500).SendString("realm mismatch")
		}

		if realmCtx.UserID != sessionInfo.UserID {
			return c.Status(500).SendString("user mismatch")
		}

		if realmCtx.Source != "session" {
			return c.Status(500).SendString("source mismatch")
		}

		// Verify TenantContextKey is set
		tenantStr := c.UserContext().Value(TenantContextKey{})
		if tenantStr == nil {
			return c.Status(500).SendString("no tenant key")
		}

		tenantStrVal, ok := tenantStr.(string)
		if !ok || tenantStrVal != sessionInfo.TenantID.String() {
			return c.Status(500).SendString("tenant key mismatch")
		}

		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-token"})

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(body))
}

// TestSessionMiddleware_EmptyTokenAfterParse tests the edge case where
// token is empty after parsing Bearer prefix.
// Note: HTTP header values are trimmed, so "Bearer " becomes "Bearer"
// which fails the format check (needs exactly 2 parts after split).
