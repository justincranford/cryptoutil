// Copyright (c) 2025 Justin Cranford

package handler

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestErrorResponsePaths tests error response generation for various HTTP error codes.
func TestErrorResponsePaths(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name           string
		statusCode     int
		errorCode      string
		errorMessage   string
		expectedStatus int
	}{
		{
			name:           "NotFound_404",
			statusCode:     fiber.StatusNotFound,
			errorCode:      "not_found",
			errorMessage:   "resource not found",
			expectedStatus: 404,
		},
		{
			name:           "BadRequest_400",
			statusCode:     fiber.StatusBadRequest,
			errorCode:      "bad_request",
			errorMessage:   "invalid input",
			expectedStatus: 400,
		},
		{
			name:           "Unauthorized_401",
			statusCode:     fiber.StatusUnauthorized,
			errorCode:      "unauthorized",
			errorMessage:   "authentication required",
			expectedStatus: 401,
		},
		{
			name:           "InternalServerError_500",
			statusCode:     fiber.StatusInternalServerError,
			errorCode:      "internal_error",
			errorMessage:   "server error occurred",
			expectedStatus: 500,
		},
		{
			name:           "Conflict_409",
			statusCode:     fiber.StatusConflict,
			errorCode:      "conflict",
			errorMessage:   "resource already exists",
			expectedStatus: 409,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()

			app.Get("/test-error", func(c *fiber.Ctx) error {
				return handler.errorResponse(c, tc.statusCode, tc.errorCode, tc.errorMessage)
			})

			req := httptest.NewRequest(http.MethodGet, "/test-error", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestGetEnrollmentStatus_NotFound tests enrollment status retrieval when enrollment doesn't exist.
func TestGetEnrollmentStatus_NotFound(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer:            testSetup.Issuer,
		enrollmentTracker: newEnrollmentTracker(100),
	}

	app := fiber.New()

	// Test non-existent enrollment.
	randomID := googleUuid.New()

	app.Get("/enrollment/:id", func(c *fiber.Ctx) error {
		return handler.GetEnrollmentStatus(c, randomID)
	})

	req := httptest.NewRequest(http.MethodGet, "/enrollment/"+randomID.String(), nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
