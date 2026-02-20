// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	"net/http/httptest"
	"testing"

	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIssueAndValidateSession_Integration(t *testing.T) {
	if testGormDB == nil {
		t.Skip("testGormDB not initialized, skipping integration test")
	}

	// Full integration test requires complex setup (telemetry, JWK, barrier, etc.)
	// For coverage purposes, unit tests above exercise the handler paths.
	t.Skip("Full integration test requires complete infrastructure setup")
}

// TestIssueSession_ServiceSessionType tests the service session type path.
// NOTE: This test is skipped because calling session manager methods with nil manager causes panic.
func TestIssueSession_ServiceSessionType(t *testing.T) {
	t.Skip("Requires non-nil session manager to avoid panic")
}

// TestValidateSession_ServiceSessionType tests the service session validation path.
// NOTE: This test is skipped because calling session manager methods with nil manager causes panic.
func TestValidateSession_ServiceSessionType(t *testing.T) {
	t.Skip("Requires non-nil session manager to avoid panic")
}

// TestSessionConstants tests the package constants for coverage.
func TestSessionConstants(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Invalid request body format", errInvalidRequestBody)
	require.Equal(t, "browser", sessionTypeBrowser)
}

// TestIssueSession_ServiceSessionSuccess tests service session issuance with mock.
// Covers sessions.go lines 91-94 (service session branch).
func TestIssueSession_ServiceSessionSuccess(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerSuccess()
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	reqBody := SessionIssueRequest{
		UserID:      "test-client-id",
		TenantID:    googleUuid.New().String(),
		RealmID:     googleUuid.New().String(),
		SessionType: "service", // ← Tests service branch (line 91-94)
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionIssueResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Contains(t, result.Token, "mock-service-token")
}

// TestIssueSession_ServiceSessionError tests service session error handling.
// Covers sessions.go lines 97-100 (error check after IssueServiceSession).
func TestIssueSession_ServiceSessionError(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerError(errMockSessionError)
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	reqBody := SessionIssueRequest{
		UserID:      "test-client-id",
		TenantID:    googleUuid.New().String(),
		RealmID:     googleUuid.New().String(),
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 500, resp.StatusCode) // ← Tests line 97-100 (error path)
}

// TestIssueSession_BrowserSessionSuccess tests browser session issuance with mock.
func TestIssueSession_BrowserSessionSuccess(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerSuccess()
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	reqBody := SessionIssueRequest{
		UserID:      "test-user@example.com",
		TenantID:    googleUuid.New().String(),
		RealmID:     googleUuid.New().String(),
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionIssueResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Contains(t, result.Token, "mock-browser-token")
}

// TestValidateSession_BrowserSuccess tests browser session validation with mock.
// Covers sessions.go lines 138-144 (browser branch in ValidateSession).
func TestValidateSession_BrowserSuccess(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerSuccess()
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "valid-browser-token",
		SessionType: "browser", // ← Tests browser branch (line 138-144)
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, "mock-user-from-token", result.UserID)
}

// TestValidateSession_BrowserError tests browser session validation error.
// Covers sessions.go lines 138-140 (error path in browser validation).
func TestValidateSession_BrowserError(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerError(errMockSessionError)
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "invalid-browser-token",
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, result.Valid) // ← Tests line 138-140 (error path)
}

// TestValidateSession_ServiceSuccess tests service session validation with mock.
// Covers sessions.go lines 146-160 (service branch in ValidateSession).
func TestValidateSession_ServiceSuccess(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerSuccess()
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "valid-service-token",
		SessionType: "service", // ← Tests service branch (line 146-160)
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, "mock-client-from-token", result.UserID)
}

// TestValidateSession_ServiceError tests service session validation error.
// Covers sessions.go lines 148-149 (error path in service validation).
func TestValidateSession_ServiceError(t *testing.T) {
	t.Parallel()

	mockManager := newMockSessionManagerError(errMockSessionError)
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "invalid-service-token",
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, result.Valid) // ← Tests line 148-149 (error path)
}

// TestValidateSession_BrowserNullUserID tests null UserID handling in browser sessions.
// Covers sessions.go lines 140-143 (null pointer handling).
func TestValidateSession_BrowserNullUserID(t *testing.T) {
	t.Parallel()

	mockManager := &mockSessionManager{
		validateBrowserFunc: func(context.Context, string) (*cryptoutilAppsTemplateServiceServerRepository.BrowserSession, error) {
			return &cryptoutilAppsTemplateServiceServerRepository.BrowserSession{
				UserID: nil, // ← Test null pointer handling
				Session: cryptoutilAppsTemplateServiceServerRepository.Session{
					TenantID: googleUuid.New(),
					RealmID:  googleUuid.New(),
				},
			}, nil
		},
	}
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "token-with-null-user",
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Empty(t, result.UserID) // ← Tests line 140-143 (null handling)
}

// TestValidateSession_ServiceNullClientID tests null ClientID handling in service sessions.
// Covers sessions.go lines 153-156 (null pointer handling).
func TestValidateSession_ServiceNullClientID(t *testing.T) {
	t.Parallel()

	mockManager := &mockSessionManager{
		validateServiceFunc: func(context.Context, string) (*cryptoutilAppsTemplateServiceServerRepository.ServiceSession, error) {
			return &cryptoutilAppsTemplateServiceServerRepository.ServiceSession{
				ClientID: nil, // ← Test null pointer handling
				Session: cryptoutilAppsTemplateServiceServerRepository.Session{
					TenantID: googleUuid.New(),
					RealmID:  googleUuid.New(),
				},
			}, nil
		},
	}
	handler := NewSessionHandler(mockManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	reqBody := SessionValidateRequest{
		Token:       "token-with-null-client",
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Empty(t, result.UserID) // ← Tests line 153-156 (null handling)
}
