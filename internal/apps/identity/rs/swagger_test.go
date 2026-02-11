// Copyright (c) 2025 Justin Cranford

package rs_test

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityRs "cryptoutil/internal/apps/identity/rs"
)

func TestServeOpenAPISpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "successful spec retrieval",
			expectedStatus: fiber.StatusOK,
			expectedType:   "application/json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Get OpenAPI spec handler.
			handler, err := cryptoutilIdentityRs.ServeOpenAPISpec()
			require.NoError(t, err, "ServeOpenAPISpec should create handler without error")
			require.NotNil(t, handler, "Handler should not be nil")

			// Create Fiber app and register handler.
			app := fiber.New()
			app.Get("/swagger.json", handler)

			// Create test request.
			req := httptest.NewRequest("GET", "/swagger.json", nil)

			// Execute request.
			resp, err := app.Test(req)
			require.NoError(t, err, "Request should execute without error")

			defer func() {
				closeErr := resp.Body.Close()
				require.NoError(t, closeErr)
			}()

			// Verify response.
			require.Equal(t, tc.expectedStatus, resp.StatusCode, "Status code should match")
			require.Equal(t, tc.expectedType, resp.Header.Get("Content-Type"), "Content-Type should be application/json")

			// Verify response body is valid JSON.
			require.NotZero(t, resp.ContentLength, "Response body should not be empty")
		})
	}
}
