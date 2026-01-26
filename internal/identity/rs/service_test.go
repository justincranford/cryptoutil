// Copyright (c) 2025 Justin Cranford
//
//

package rs_test

import (
	"context"
	json "encoding/json"
	"io"
	"log/slog"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRs "cryptoutil/internal/identity/rs"

	fiber "github.com/gofiber/fiber/v2"
	testify "github.com/stretchr/testify/require"
)

// mockTokenService provides a test double for TokenService.
type mockTokenService struct {
	validateFunc   func(ctx context.Context, token string) (map[string]any, error)
	isActiveFunc   func(claims map[string]any) bool
	introspectFunc func(ctx context.Context, token string) (*cryptoutilIdentityIssuer.TokenMetadata, error)
}

func (m *mockTokenService) ValidateAccessToken(ctx context.Context, token string) (map[string]any, error) {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, token)
	}

	return nil, nil
}

func (m *mockTokenService) IsTokenActive(claims map[string]any) bool {
	if m.isActiveFunc != nil {
		return m.isActiveFunc(claims)
	}

	return true
}

func (m *mockTokenService) IntrospectToken(ctx context.Context, token string) (*cryptoutilIdentityIssuer.TokenMetadata, error) {
	if m.introspectFunc != nil {
		return m.introspectFunc(ctx, token)
	}

	return nil, nil
}

// setupTestService creates a configured resource server service and Fiber app.
func setupTestService(t *testing.T) (*fiber.App, *mockTokenService) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		RS: &cryptoutilIdentityConfig.ServerConfig{
			BindAddress: "127.0.0.1",
			Port:        9100,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create mock token service.
	tokenSvc := &mockTokenService{}

	// Create resource server service.
	service := cryptoutilIdentityRs.NewService(config, logger, tokenSvc)

	// Create Fiber app and register routes.
	app := fiber.New()
	service.RegisterMiddleware(app)
	service.RegisterRoutes(app)

	return app, tokenSvc
}

// createBearerToken creates a test Bearer token string.
func createBearerToken(token string) string {
	return "Bearer " + token
}

// TestPublicEndpoint tests that public endpoints don't require authentication.
func TestPublicEndpoint(t *testing.T) {
	app, _ := setupTestService(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/health", nil)
	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse response.
	var result map[string]any

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err)
	err = json.Unmarshal(body, &result)
	testify.NoError(t, err)

	testify.Equal(t, "healthy", result["status"])
}

// TestProtectedEndpoint_NoToken tests that protected endpoints reject requests without tokens.
func TestProtectedEndpoint_NoToken(t *testing.T) {
	app, _ := setupTestService(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Parse error response.
	var result map[string]any

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err)
	err = json.Unmarshal(body, &result)
	testify.NoError(t, err)

	testify.Equal(t, cryptoutilIdentityMagic.ErrorInvalidToken, result["error"])
}

// TestProtectedEndpoint_InvalidTokenFormat tests Bearer token format validation.
func TestProtectedEndpoint_InvalidTokenFormat(t *testing.T) {
	app, _ := setupTestService(t)

	testCases := []struct {
		name   string
		header string
	}{
		{"Missing Bearer Prefix", "token123"},
		{"Empty Token", "Bearer "},
		{"Invalid Format", "Basic token123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
			req.Header.Set("Authorization", tc.header)

			resp, err := app.Test(req)
			testify.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			testify.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
}

// TestScopeEnforcement_MissingScope tests scope enforcement for protected resources.
func TestScopeEnforcement_MissingScope(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	// Configure mock to return claims without required scope.
	tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
		return map[string]any{
			cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			cryptoutilIdentityMagic.ClaimClientID: "test-client",
			cryptoutilIdentityMagic.ClaimScope:    "write:other",
		}, nil
	}
	tokenSvc.isActiveFunc = func(_ map[string]any) bool {
		return true
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
	req.Header.Set("Authorization", createBearerToken("valid-token"))

	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusForbidden, resp.StatusCode)

	// Parse error response.
	var result map[string]any

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err)
	err = json.Unmarshal(body, &result)
	testify.NoError(t, err)

	testify.Equal(t, cryptoutilIdentityMagic.ErrorInsufficientScope, result["error"])
}

// TestScopeEnforcement_ValidScope tests successful scope validation.
func TestScopeEnforcement_ValidScope(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	// Configure mock to return valid claims with required scope.
	tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
		return map[string]any{
			cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			cryptoutilIdentityMagic.ClaimClientID: "test-client",
			cryptoutilIdentityMagic.ClaimScope:    "read:resource write:resource",
		}, nil
	}
	tokenSvc.isActiveFunc = func(_ map[string]any) bool {
		return true
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
	req.Header.Set("Authorization", createBearerToken("valid-token"))

	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse success response.
	var result map[string]any

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err)
	err = json.Unmarshal(body, &result)
	testify.NoError(t, err)

	testify.Equal(t, "Protected resource accessed successfully", result["message"])
}

// TestAdminEndpoint_RequiresAdminScope tests admin endpoint scope enforcement.
func TestAdminEndpoint_RequiresAdminScope(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	testCases := []struct {
		name           string
		scope          string
		expectedStatus int
	}{
		{"No Admin Scope", "read:resource write:resource", http.StatusForbidden},
		{"With Admin Scope", "admin read:resource", http.StatusOK},
		{"Only Admin Scope", "admin", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
				return map[string]any{
					cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilIdentityMagic.ClaimClientID: "test-client",
					cryptoutilIdentityMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req)
			testify.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			testify.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestCreateResource_RequiresWriteScope tests POST endpoint scope enforcement.
func TestCreateResource_RequiresWriteScope(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	// Configure mock for write scope.
	tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
		return map[string]any{
			cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			cryptoutilIdentityMagic.ClaimClientID: "test-client",
			cryptoutilIdentityMagic.ClaimScope:    "write:resource",
		}, nil
	}
	tokenSvc.isActiveFunc = func(_ map[string]any) bool {
		return true
	}

	reqBody := strings.NewReader(`{"name":"test","value":"data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/protected/resource", reqBody)
	req.Header.Set("Authorization", createBearerToken("valid-token"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusCreated, resp.StatusCode)

	// Parse success response.
	var result map[string]any

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err)
	err = json.Unmarshal(body, &result)
	testify.NoError(t, err)

	testify.Equal(t, "Resource created successfully", result["message"])
}

// TestDeleteResource_RequiresDeleteScope tests DELETE endpoint scope enforcement.
func TestDeleteResource_RequiresDeleteScope(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	testCases := []struct {
		name           string
		scope          string
		expectedStatus int
	}{
		{"Missing Delete Scope", "read:resource write:resource", http.StatusForbidden},
		{"With Delete Scope", "delete:resource", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
				return map[string]any{
					cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilIdentityMagic.ClaimClientID: "test-client",
					cryptoutilIdentityMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/protected/resource/123", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req)
			testify.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			testify.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

// TestExpiredToken tests that expired tokens are rejected.
func TestExpiredToken(t *testing.T) {
	app, tokenSvc := setupTestService(t)

	// Configure mock to return expired token.
	tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
		return map[string]any{
			cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(-1 * time.Hour).Unix()),
			cryptoutilIdentityMagic.ClaimClientID: "test-client",
			cryptoutilIdentityMagic.ClaimScope:    "read:resource",
		}, nil
	}
	tokenSvc.isActiveFunc = func(claims map[string]any) bool {
		// Check expiration.
		now := time.Now().UTC().Unix()
		if exp, ok := claims[cryptoutilIdentityMagic.ClaimExp].(float64); ok {
			return int64(exp) >= now
		}

		return false
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
	req.Header.Set("Authorization", createBearerToken("expired-token"))

	resp, err := app.Test(req)
	testify.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	testify.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestService_Start tests service startup.
func TestService_Start(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		RS: &cryptoutilIdentityConfig.ServerConfig{
			BindAddress: "127.0.0.1",
			Port:        9100,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenSvc := &mockTokenService{}

	service := cryptoutilIdentityRs.NewService(config, logger, tokenSvc)

	ctx := context.Background()

	err := service.Start(ctx)
	testify.NoError(t, err, "Start should complete without error")
}

// TestService_Stop tests service shutdown.
func TestService_Stop(t *testing.T) {
	t.Parallel()

	config := &cryptoutilIdentityConfig.Config{
		RS: &cryptoutilIdentityConfig.ServerConfig{
			BindAddress: "127.0.0.1",
			Port:        9100,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokenSvc := &mockTokenService{}

	service := cryptoutilIdentityRs.NewService(config, logger, tokenSvc)

	ctx := context.Background()

	err := service.Stop(ctx)
	testify.NoError(t, err, "Stop should complete without error")
}

// TestAdminMetrics_RequiresAdminScope tests admin metrics endpoint.
func TestAdminMetrics_RequiresAdminScope(t *testing.T) {
	// NOTE: No t.Parallel() - subtests share app/tokenSvc, parallel causes race on validateFunc.
	app, tokenSvc := setupTestService(t)

	testCases := []struct {
		name           string
		scope          string
		expectedStatus int
		checkMetrics   bool
	}{
		{
			name:           "missing_admin_scope",
			scope:          "read:resource write:resource",
			expectedStatus: http.StatusForbidden,
			checkMetrics:   false,
		},
		{
			name:           "with_admin_scope",
			scope:          "admin",
			expectedStatus: http.StatusOK,
			checkMetrics:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTE: No t.Parallel() - subtests share mockTokenService, parallel execution causes race.
			tokenSvc.validateFunc = func(_ context.Context, _ string) (map[string]any, error) {
				return map[string]any{
					cryptoutilIdentityMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilIdentityMagic.ClaimClientID: "test-client",
					cryptoutilIdentityMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/metrics", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req)
			testify.NoError(t, err)

			defer func() {
				testify.NoError(t, resp.Body.Close())
			}()

			testify.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.checkMetrics {
				// Parse metrics response.
				var result map[string]any

				body, err := io.ReadAll(resp.Body)
				testify.NoError(t, err)

				err = json.Unmarshal(body, &result)
				testify.NoError(t, err)

				testify.Equal(t, "System metrics", result["message"])

				// Verify metrics data exists.
				metrics, ok := result["metrics"].(map[string]any)
				testify.True(t, ok, "metrics field should be a map")
				testify.Contains(t, metrics, "requests_total")
				testify.Contains(t, metrics, "requests_success")
				testify.Contains(t, metrics, "requests_failed")

				// Verify metric values are correct integers.
				testify.Equal(t, float64(cryptoutilIdentityMagic.ExampleMetricRequestsTotal), metrics["requests_total"])
				testify.Equal(t, float64(cryptoutilIdentityMagic.ExampleMetricRequestsSuccess), metrics["requests_success"])
				testify.Equal(t, float64(cryptoutilIdentityMagic.ExampleMetricRequestsFailed), metrics["requests_failed"])
			}
		})
	}
}
