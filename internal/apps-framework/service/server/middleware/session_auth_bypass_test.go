// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

package middleware

import (
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func newSessionAuthBypassTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, _ error) error {
			return c.SendStatus(fiber.StatusUnauthorized)
		},
	})
}

func TestSessionMiddleware_ExactPublicAuthBypass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		isBrowser      bool
		expectedStatus int
	}{
		{
			name:           "service registration bypasses session validation",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthRegisterPath,
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:           "service client registration bypasses session validation",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath,
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:           "service client credential exchange bypasses session validation",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthTokenPath,
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:           "browser registration retains unauthenticated behavior",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicBrowserAuthRegisterPath,
			isBrowser:      true,
			expectedStatus: fiber.StatusNoContent,
		},
		{
			name:           "service username password login remains protected",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthLoginPath,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "service registration with a different method remains protected",
			method:         fiber.MethodGet,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthRegisterPath,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "service nested client registration path remains protected",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAuthClientRegisterPath + "/extra",
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "service unrelated auth path remains protected",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath + "/auth/other",
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "browser login remains protected",
			method:         fiber.MethodPost,
			path:           cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath + "/auth/login",
			isBrowser:      true,
			expectedStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := newSessionAuthBypassTestApp()
			app.Use(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath+"/*", ServiceSessionMiddleware(nil))
			app.Use(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath+"/*", BrowserSessionMiddleware(nil))
			app.All(tc.path, func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusNoContent)
			})

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			resp, err := app.Test(req, -1)

			require.NoError(t, err)
			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}
