package server

import (
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRegisterKMSRoutes_PKIEndpoints(t *testing.T) {
	t.Parallel()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	settings := &cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings{
		PublicBrowserAPIContextPath: cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
		PublicServiceAPIContextPath: cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
	}
	err := registerKMSRoutes(app, nil, settings, nil)
	require.NoError(t, err)

	tests := []struct {
		method string
		path   string
	}{
		{method: fiber.MethodGet, path: "/pki/cas"},
		{method: fiber.MethodGet, path: "/pki/cas/:caID"},
		{method: fiber.MethodGet, path: "/pki/cas/:caID/crl"},
		{method: fiber.MethodPost, path: "/pki/enrollments"},
		{method: fiber.MethodGet, path: "/pki/enrollments/:requestID"},
		{method: fiber.MethodGet, path: "/pki/certificates"},
		{method: fiber.MethodGet, path: "/pki/certificates/:serialNumber"},
		{method: fiber.MethodGet, path: "/pki/certificates/:serialNumber/chain"},
		{method: fiber.MethodPost, path: "/pki/certificates/:serialNumber/revoke"},
		{method: fiber.MethodGet, path: "/pki/profiles"},
		{method: fiber.MethodGet, path: "/pki/profiles/:profileID"},
		{method: fiber.MethodPost, path: "/pki/ocsp"},
		{method: fiber.MethodGet, path: "/pki/est/cacerts"},
		{method: fiber.MethodPost, path: "/pki/est/simpleenroll"},
		{method: fiber.MethodPost, path: "/pki/est/simplereenroll"},
		{method: fiber.MethodPost, path: "/pki/est/serverkeygen"},
		{method: fiber.MethodGet, path: "/pki/est/csrattrs"},
		{method: fiber.MethodPost, path: "/pki/tsa/timestamp"},
	}

	for _, prefix := range []string{
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
	} {
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s %s%s", tc.method, prefix, tc.path), func(t *testing.T) {
				t.Parallel()

				found := false

				for _, route := range app.GetRoutes() {
					if route.Method == tc.method && route.Path == prefix+tc.path {
						found = true

						break
					}
				}

				require.True(t, found, "route %s %s was not registered", tc.method, prefix+tc.path)
			})
		}
	}

	for _, prefix := range []string{
		cryptoutilSharedMagic.DefaultPublicServiceAPIContextPath,
		cryptoutilSharedMagic.DefaultPublicBrowserAPIContextPath,
	} {
		t.Run(prefix, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, prefix+"/pki/profiles", nil)
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
			require.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}
