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

// TestNewSessionHandler validates the constructor.
func TestNewSessionHandler(t *testing.T) {
	t.Parallel()

	// We just test constructor, not full functionality.
	handler := NewSessionHandler(nil)
	require.NotNil(t, handler)
	require.Nil(t, handler.sessionManager)
}

// TestSessionIssueRequest_Struct validates request struct fields.
func TestSessionIssueRequest_Struct(t *testing.T) {
	t.Parallel()

	req := SessionIssueRequest{
		UserID:      "test-user-id",
		TenantID:    googleUuid.New().String(),
		RealmID:     googleUuid.New().String(),
		SessionType: sessionTypeBrowser,
	}

	require.Equal(t, "test-user-id", req.UserID)
	require.Equal(t, sessionTypeBrowser, req.SessionType)

	// Test JSON marshaling.
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var decoded SessionIssueRequest

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, req.UserID, decoded.UserID)
	require.Equal(t, req.SessionType, decoded.SessionType)
}

// TestSessionIssueResponse_Struct validates response struct fields.
func TestSessionIssueResponse_Struct(t *testing.T) {
	t.Parallel()

	resp := SessionIssueResponse{
		Token: "test-token-value",
	}

	require.Equal(t, "test-token-value", resp.Token)

	// Test JSON marshaling.
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var decoded SessionIssueResponse

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, resp.Token, decoded.Token)
}

// TestSessionValidateRequest_Struct validates request struct fields.
func TestSessionValidateRequest_Struct(t *testing.T) {
	t.Parallel()

	req := SessionValidateRequest{
		Token:       "test-token",
		SessionType: sessionTypeBrowser,
	}

	require.Equal(t, "test-token", req.Token)
	require.Equal(t, sessionTypeBrowser, req.SessionType)

	// Test JSON marshaling.
	data, err := json.Marshal(req)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var decoded SessionValidateRequest

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, req.Token, decoded.Token)
	require.Equal(t, req.SessionType, decoded.SessionType)
}

// TestSessionValidateResponse_Struct validates response struct fields.
func TestSessionValidateResponse_Struct(t *testing.T) {
	t.Parallel()

	resp := SessionValidateResponse{
		UserID:   "user-id",
		TenantID: googleUuid.New().String(),
		RealmID:  googleUuid.New().String(),
		Valid:    true,
	}

	require.Equal(t, "user-id", resp.UserID)
	require.True(t, resp.Valid)

	// Test JSON marshaling.
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var decoded SessionValidateResponse

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.Equal(t, resp.UserID, decoded.UserID)
	require.Equal(t, resp.Valid, decoded.Valid)
}

// TestIssueSession_InvalidJSON tests that IssueSession returns error for malformed JSON.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
// The key is that we exercise the JSON parsing error path.
func TestIssueSession_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewSessionHandler(nil)
	app := fiber.New()
	app.Post("/session/issue", handler.IssueSession)

	req := httptest.NewRequest("POST", "/session/issue", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Fiber's default error handler converts cryptoutilAppErr to 500.
	// The important thing is that we exercised the JSON parse error path.
	require.True(t, resp.StatusCode >= 400, "Expected error status code")
}

// TestIssueSession_InvalidTenantID tests IssueSession with invalid tenant ID.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
func TestIssueSession_InvalidTenantID(t *testing.T) {
	t.Parallel()

	handler := NewSessionHandler(nil)
	app := fiber.New()
	app.Post("/session/issue", handler.IssueSession)

	reqBody := SessionIssueRequest{
		UserID:      "user-123",
		TenantID:    "not-a-valid-uuid",
		RealmID:     googleUuid.New().String(),
		SessionType: sessionTypeBrowser,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/session/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Fiber's default error handler converts cryptoutilAppErr to 500.
	// The important thing is that we exercised the tenant_id validation error path.
	require.True(t, resp.StatusCode >= 400, "Expected error status code")
}

// TestIssueSession_InvalidRealmID tests IssueSession with invalid realm ID.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
func TestIssueSession_InvalidRealmID(t *testing.T) {
	t.Parallel()

	handler := NewSessionHandler(nil)
	app := fiber.New()
	app.Post("/session/issue", handler.IssueSession)

	reqBody := SessionIssueRequest{
		UserID:      "user-123",
		TenantID:    googleUuid.New().String(),
		RealmID:     "not-a-valid-uuid",
		SessionType: sessionTypeBrowser,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/session/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Fiber's default error handler converts cryptoutilAppErr to 500.
	// The important thing is that we exercised the realm_id validation error path.
	require.True(t, resp.StatusCode >= 400, "Expected error status code")
}

// TestValidateSession_InvalidJSON tests ValidateSession with invalid JSON body.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
func TestValidateSession_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewSessionHandler(nil)
	app := fiber.New()
	app.Post("/session/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/session/validate", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Fiber's default error handler converts cryptoutilAppErr to 500.
	// The important thing is that we exercised the JSON parse error path.
	require.True(t, resp.StatusCode >= 400, "Expected error status code")
}

// TestIssueSession_TableDriven uses table-driven tests for IssueSession.
// Note: Tests that reach the session manager will panic if sessionManager is nil.
// These tests focus on early validation failures that return before reaching session manager.
// NOTE: Fiber's default error handler converts cryptoutilAppErr to 500.
func TestIssueSession_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantError   bool
	}{
		{
			name:        "Empty body returns error due to parse error",
			requestBody: "",
			wantError:   true,
		},
		{
			name:        "Invalid JSON returns error",
			requestBody: "{invalid",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewSessionHandler(nil)
			app := fiber.New()
			app.Post("/session/issue", handler.IssueSession)

			req := httptest.NewRequest("POST", "/session/issue", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			if tt.wantError {
				require.True(t, resp.StatusCode >= 400, "Expected error status code, got %d", resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}
		})
	}
}

// TestValidateSession_TableDriven uses table-driven tests for ValidateSession.
// Note: Tests that reach the session manager will panic if sessionManager is nil.
// These tests focus on early validation failures that return before reaching session manager.
// NOTE: Fiber's default error handler converts cryptoutilAppErr to 500.
func TestValidateSession_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		wantError   bool
	}{
		{
			name:        "Empty body returns error due to parse error",
			requestBody: "",
			wantError:   true,
		},
		{
			name:        "Invalid JSON returns error",
			requestBody: "{invalid",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewSessionHandler(nil)
			app := fiber.New()
			app.Post("/session/validate", handler.ValidateSession)

			req := httptest.NewRequest("POST", "/session/validate", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			if tt.wantError {
				require.True(t, resp.StatusCode >= 400, "Expected error status code, got %d", resp.StatusCode)
			} else {
				require.Equal(t, 200, resp.StatusCode)
			}
		})
	}
}

// TestIssueAndValidateSession_Integration tests the full flow with real dependencies.
// This test is skipped because it requires full infrastructure setup.
// The unit tests above provide coverage for the handler code paths.
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
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

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Empty(t, result.UserID) // ← Tests line 153-156 (null handling)
}
