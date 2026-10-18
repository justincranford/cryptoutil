// Copyright (c) 2025-2026 Justin Cranford.
package apis_test

import (
	"bytes"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity-authz/server/apis"
)

func TestMFAAdminHandlers_ValidationPaths(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	t.Cleanup(func() {
		require.NoError(t, repoFactory.Close())
	})

	service := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)

	tests := []struct {
		name       string
		method     string
		path       func(string, string) string
		body       string
		wantStatus int
	}{
		{
			name:       "enroll rejects malformed body",
			method:     testPost,
			path:       func(_, _ string) string { return "/oidc/v1/mfa/enroll" },
			body:       "{",
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "list requires user id",
			method:     http.MethodGet,
			path:       func(_, _ string) string { return "/oidc/v1/mfa/factors" },
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:   "list reports unknown user",
			method: http.MethodGet,
			path: func(userID, _ string) string {
				return fmt.Sprintf("/oidc/v1/mfa/factors?user_id=%s", userID)
			},
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:   "delete rejects malformed factor id",
			method: http.MethodDelete,
			path: func(userID, _ string) string {
				return "/oidc/v1/mfa/factors/not-a-uuid?user_id=" + userID
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:   "delete requires user id",
			method: http.MethodDelete,
			path: func(_, factorID string) string {
				return "/oidc/v1/mfa/factors/" + factorID
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:   "delete rejects malformed user id",
			method: http.MethodDelete,
			path: func(_, factorID string) string {
				return "/oidc/v1/mfa/factors/" + factorID + "?user_id=not-a-uuid"
			},
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:   "delete reports unknown user",
			method: http.MethodDelete,
			path: func(userID, factorID string) string {
				return fmt.Sprintf("/oidc/v1/mfa/factors/%s?user_id=%s", factorID, userID)
			},
			wantStatus: fiber.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userID := googleUuid.Must(googleUuid.NewV7()).String()
			factorID := googleUuid.Must(googleUuid.NewV7()).String()
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			service.RegisterRoutes(app)

			request := httptest.NewRequest(tc.method, tc.path(userID, factorID), bytes.NewReader([]byte(tc.body)))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}

			response, err := app.Test(request, -1)

			require.NoError(t, err)
			defer func() {
				require.NoError(t, response.Body.Close())
			}()

			require.Equal(t, tc.wantStatus, response.StatusCode)
		})
	}
}
