// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testPKIComponentsRequiredError = "PKI components and handler are required"

func TestRegisterPKIRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceGroup func(*fiber.App) fiber.Router
		browserGroup func(*fiber.App) fiber.Router
		components   *PKIComponents
		wantErr      string
	}{
		{
			name: "missing service route group",
			browserGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath)
			},
			components: &PKIComponents{},
			wantErr:    "PKI service route group is required",
		},
		{
			name: "missing browser route group",
			serviceGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)
			},
			components: &PKIComponents{},
			wantErr:    "PKI browser route group is required",
		},
		{
			name: "missing components",
			serviceGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)
			},
			browserGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath)
			},
			wantErr: testPKIComponentsRequiredError,
		},
		{
			name: "missing handler",
			serviceGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)
			},
			browserGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath)
			},
			components: &PKIComponents{},
			wantErr:    testPKIComponentsRequiredError,
		},
		{
			name: "registers service and browser routes",
			serviceGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath)
			},
			browserGroup: func(app *fiber.App) fiber.Router {
				return app.Group(cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath)
			},
			components: newPKIComponentsForTest(t),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			var serviceGroup, browserGroup fiber.Router
			if tc.serviceGroup != nil {
				serviceGroup = tc.serviceGroup(app)
			}

			if tc.browserGroup != nil {
				browserGroup = tc.browserGroup(app)
			}

			err := RegisterPKIRoutes(serviceGroup, browserGroup, tc.components)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			for _, path := range []string{"/service/api/v1/pki/profiles", "/browser/api/v1/pki/profiles"} {
				req := httptest.NewRequest("GET", path, nil)
				resp, requestErr := app.Test(req, -1)
				require.NoError(t, requestErr)
				require.Equal(t, fiber.StatusOK, resp.StatusCode)
				require.NoError(t, resp.Body.Close())
			}
		})
	}
}

func newPKIComponentsForTest(t *testing.T) *PKIComponents {
	t.Helper()

	components, err := NewInMemoryPKIComponents()
	require.NoError(t, err)

	return components
}
