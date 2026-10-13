// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
)

func TestGetCRL_ConfigurationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		handler     *Handler
		wantStatus  int
		wantErrCode string
	}{
		{
			name:        "CRL service is not configured",
			handler:     &Handler{},
			wantStatus:  fiber.StatusServiceUnavailable,
			wantErrCode: "service_unavailable",
		},
		{
			name:        "issuer is not configured",
			handler:     &Handler{crlService: &cryptoutilCAServiceRevocation.CRLService{}},
			wantStatus:  fiber.StatusInternalServerError,
			wantErrCode: "issuer_not_configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
				return tc.handler.GetCRL(c, c.Params("caId"), cryptoutilApiCaServer.GetCRLParams{})
			})

			request := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl", nil)
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Contains(t, string(body), tc.wantErrCode)
		})
	}
}
