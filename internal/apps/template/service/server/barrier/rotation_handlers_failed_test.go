// Copyright (c) 2025 Justin Cranford
//
//

package barrier

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilUnsealKeysService "cryptoutil/internal/apps/template/service/server/barrier/unsealkeysservice"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"

	fiber "github.com/gofiber/fiber/v2"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRotateKey_TooLongReason(t *testing.T) {
	app, _, _ := setupRotationTestEnvironment(t)

	// Create a reason that exceeds MaxRotationReasonLength (500 characters).
	longReason := strings.Repeat("a", MaxRotationReasonLength+1)

	// Test all three endpoints.
	endpoints := []string{
		"/admin/api/v1/barrier/rotate/root",
		"/admin/api/v1/barrier/rotate/intermediate",
		"/admin/api/v1/barrier/rotate/content",
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("endpoint=%s", endpoint), func(t *testing.T) {
			reqBody := map[string]string{
				"reason": longReason,
			}
			reqJSON, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", endpoint, bytes.NewReader(reqJSON))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			// Parse error response.
			respBody, _ := io.ReadAll(resp.Body)

			var errResp map[string]string

			err = json.Unmarshal(respBody, &errResp)
			require.NoError(t, err)
			require.Equal(t, "validation_error", errResp["error"])
			require.Contains(t, errResp["message"], "at most 500 characters")
		})
	}
}

// TestRotateKey_InvalidJSON tests that rotation requests fail with invalid JSON body.
func TestRotateKey_InvalidJSON(t *testing.T) {
	app, _, _ := setupRotationTestEnvironment(t)

	// Test all three endpoints.
	endpoints := []string{
		"/admin/api/v1/barrier/rotate/root",
		"/admin/api/v1/barrier/rotate/intermediate",
		"/admin/api/v1/barrier/rotate/content",
	}

	for _, endpoint := range endpoints {
		t.Run(fmt.Sprintf("endpoint=%s", endpoint), func(t *testing.T) {
			// Send invalid JSON.
			req := httptest.NewRequest("POST", endpoint, bytes.NewReader([]byte("{invalid json")))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			// Parse error response.
			respBody, _ := io.ReadAll(resp.Body)

			var errResp map[string]string

			err = json.Unmarshal(respBody, &errResp)
			require.NoError(t, err)
			require.Equal(t, "invalid_request_body", errResp["error"])
			require.Contains(t, errResp["message"], "Failed to parse request body")
		})
	}
}

// TestHandleRotateRootKey_RotationFailed tests that HandleRotateRootKey returns error on rotation failure.
func TestHandleRotateRootKey_RotationFailed(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error when getting latest root key.
	mockRepo := newMockRotationRepository()
	mockRepo.tx.getRootKeyLatestErr = errMockRotationDBFailure

	// Create rotation service with mock repository.
	ctx := context.Background()
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	// Create Fiber app and register routes.
	app := fiber.New()
	RegisterRotationRoutes(app, rotationService)

	// Make rotation request.
	reqBody := map[string]string{"reason": "test rotation"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/root", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Parse error response.
	respBody, _ := io.ReadAll(resp.Body)

	var errResp map[string]string

	err = json.Unmarshal(respBody, &errResp)
	require.NoError(t, err)
	require.Equal(t, "rotation_failed", errResp["error"])
	require.Contains(t, errResp["message"], "Failed to rotate root key")
}

// TestHandleRotateIntermediateKey_RotationFailed tests that HandleRotateIntermediateKey returns error on rotation failure.
func TestHandleRotateIntermediateKey_RotationFailed(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error when getting latest root key.
	mockRepo := newMockRotationRepository()
	mockRepo.tx.getRootKeyLatestErr = errMockRotationDBFailure

	// Create rotation service with mock repository.
	ctx := context.Background()
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	// Create Fiber app and register routes.
	app := fiber.New()
	RegisterRotationRoutes(app, rotationService)

	// Make rotation request.
	reqBody := map[string]string{"reason": "test rotation"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/intermediate", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Parse error response.
	respBody, _ := io.ReadAll(resp.Body)

	var errResp map[string]string

	err = json.Unmarshal(respBody, &errResp)
	require.NoError(t, err)
	require.Equal(t, "rotation_failed", errResp["error"])
	require.Contains(t, errResp["message"], "Failed to rotate intermediate key")
}

// TestHandleRotateContentKey_RotationFailed tests that HandleRotateContentKey returns error on rotation failure.
func TestHandleRotateContentKey_RotationFailed(t *testing.T) {
	t.Parallel()

	// Create mock repository that returns error when getting latest intermediate key.
	mockRepo := newMockRotationRepository()
	mockRepo.tx.getIntermediateKeyLatestErr = errMockRotationDBFailure

	// Create rotation service with mock repository.
	ctx := context.Background()
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true))
	require.NoError(t, err)
	t.Cleanup(func() { telemetryService.Shutdown() })

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)
	t.Cleanup(func() { jwkGenService.Shutdown() })

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilSharedCryptoJose.EncA256GCM, &cryptoutilSharedCryptoJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)
	t.Cleanup(func() { unsealService.Shutdown() })

	rotationService, err := NewRotationService(jwkGenService, mockRepo, unsealService)
	require.NoError(t, err)

	// Create Fiber app and register routes.
	app := fiber.New()
	RegisterRotationRoutes(app, rotationService)

	// Make rotation request.
	reqBody := map[string]string{"reason": "test rotation"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/admin/api/v1/barrier/rotate/content", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	// Parse error response.
	respBody, _ := io.ReadAll(resp.Body)

	var errResp map[string]string

	err = json.Unmarshal(respBody, &errResp)
	require.NoError(t, err)
	require.Equal(t, "rotation_failed", errResp["error"])
	require.Contains(t, errResp["message"], "Failed to rotate content key")
}
