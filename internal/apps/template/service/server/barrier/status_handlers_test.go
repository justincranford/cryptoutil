// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"bytes"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestHandleGetBarrierKeysStatus_Success tests successful retrieval of barrier keys status.
func TestHandleGetBarrierKeysStatus_Success(t *testing.T) {
	t.Parallel()

	// Setup test environment (creates root + intermediate keys automatically).
	app, rotationService, _ := setupRotationTestEnvironment(t)

	// Create status service from rotation service's repository.
	statusService, err := NewStatusService(rotationService.repository)
	require.NoError(t, err)

	// Register status routes.
	RegisterStatusRoutes(app, statusService)

	// Make HTTP request.
	req := httptest.NewRequest("GET", "/admin/api/v1/barrier/keys/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Verify response body.
	var statusResp KeysStatusResponse

	err = json.NewDecoder(resp.Body).Decode(&statusResp)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	// Verify root key exists.
	require.NotNil(t, statusResp.RootKey)
	require.NotEmpty(t, statusResp.RootKey.UUID)
	require.Greater(t, statusResp.RootKey.CreatedAt, int64(0))

	// Verify intermediate key exists.
	require.NotNil(t, statusResp.IntermediateKey)
	require.NotEmpty(t, statusResp.IntermediateKey.UUID)
	require.Greater(t, statusResp.IntermediateKey.CreatedAt, int64(0))
}

// TestNewStatusService_NilRepository tests NewStatusService with nil repository.
func TestNewStatusService_NilRepository(t *testing.T) {
	t.Parallel()

	statusService, err := NewStatusService(nil)
	require.Nil(t, statusService)
	require.Error(t, err)
	require.Contains(t, err.Error(), "repository must be non-nil")
}

// TestRegisterStatusRoutes_Integration tests full HTTP integration.
func TestRegisterStatusRoutes_Integration(t *testing.T) {
	t.Parallel()

	// Setup test environment.
	_, rotationService, _ := setupRotationTestEnvironment(t)

	statusService, err := NewStatusService(rotationService.repository)
	require.NoError(t, err)

	// Create fiber app and register routes.
	app := fiber.New()
	RegisterStatusRoutes(app, statusService)

	// Verify route is registered (GET request succeeds).
	req := httptest.NewRequest("GET", "/admin/api/v1/barrier/keys/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())

	// Verify wrong HTTP method returns 405 Method Not Allowed.
	req = httptest.NewRequest("POST", "/admin/api/v1/barrier/keys/status", bytes.NewReader([]byte(`{}`)))
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusMethodNotAllowed, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
