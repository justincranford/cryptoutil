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

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityRs "cryptoutil/internal/apps/identity/rs"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	testify "github.com/stretchr/testify/require"
)

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
					cryptoutilSharedMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilSharedMagic.ClaimClientID: "test-client",
					cryptoutilSharedMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req, -1)
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
			cryptoutilSharedMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
			cryptoutilSharedMagic.ClaimClientID: "test-client",
			cryptoutilSharedMagic.ClaimScope:    "write:resource",
		}, nil
	}
	tokenSvc.isActiveFunc = func(_ map[string]any) bool {
		return true
	}

	reqBody := strings.NewReader(`{"name":"test","value":"data"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/protected/resource", reqBody)
	req.Header.Set("Authorization", createBearerToken("valid-token"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
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
					cryptoutilSharedMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilSharedMagic.ClaimClientID: "test-client",
					cryptoutilSharedMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/protected/resource/123", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req, -1)
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
			cryptoutilSharedMagic.ClaimExp:      float64(time.Now().UTC().Add(-1 * time.Hour).Unix()),
			cryptoutilSharedMagic.ClaimClientID: "test-client",
			cryptoutilSharedMagic.ClaimScope:    "read:resource",
		}, nil
	}
	tokenSvc.isActiveFunc = func(claims map[string]any) bool {
		// Check expiration.
		now := time.Now().UTC().Unix()
		if exp, ok := claims[cryptoutilSharedMagic.ClaimExp].(float64); ok {
			return int64(exp) >= now
		}

		return false
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/protected/resource", nil)
	req.Header.Set("Authorization", createBearerToken("expired-token"))

	resp, err := app.Test(req, -1)
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
					cryptoutilSharedMagic.ClaimExp:      float64(time.Now().UTC().Add(1 * time.Hour).Unix()),
					cryptoutilSharedMagic.ClaimClientID: "test-client",
					cryptoutilSharedMagic.ClaimScope:    tc.scope,
				}, nil
			}
			tokenSvc.isActiveFunc = func(_ map[string]any) bool {
				return true
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/metrics", nil)
			req.Header.Set("Authorization", createBearerToken("valid-token"))

			resp, err := app.Test(req, -1)
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
				testify.Equal(t, float64(cryptoutilSharedMagic.ExampleMetricRequestsTotal), metrics["requests_total"])
				testify.Equal(t, float64(cryptoutilSharedMagic.ExampleMetricRequestsSuccess), metrics["requests_success"])
				testify.Equal(t, float64(cryptoutilSharedMagic.ExampleMetricRequestsFailed), metrics["requests_failed"])
			}
		})
	}
}
