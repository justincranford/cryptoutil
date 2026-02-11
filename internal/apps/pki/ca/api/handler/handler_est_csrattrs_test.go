// Copyright (c) 2025 Justin Cranford

package handler

import (
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestEstCSRAttrs tests EST CSR attributes endpoint.
func TestEstCSRAttrs(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()

	app.Get("/est/csrattrs", func(c *fiber.Ctx) error {
		return handler.EstCSRAttrs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/est/csrattrs", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	// Should return 204 No Content (no specific CSR attributes required).
	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
