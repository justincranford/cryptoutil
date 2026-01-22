// Copyright (c) 2025 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build integration

package apis

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Note: These tests depend on testGormDB, testBrowserSessionRepo, testServiceSessionRepo,
// testBarrierService which should be set up in a TestMain for integration tests.
// If not available, tests will fail.

func TestIssueSession_Browser_HappyPath_Integration(t *testing.T) {
	// Skip if test infrastructure not available
	if testSessionManager == nil {
		t.Skip("Session manager not initialized - integration test infrastructure not available")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Create test tenant and realm for session
	testTenantID := googleUuid.New()
	testRealmID := googleUuid.New()
	testUserID := "test-user@example.com"

	// Prepare request
	reqBody := SessionIssueRequest{
		UserID:      testUserID,
		TenantID:    testTenantID.String(),
		RealmID:     testRealmID.String(),
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionIssueResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)

	// Verify session was created in database
	ctx := context.Background()
	session, err := testSessionManager.ValidateBrowserSession(ctx, result.Token)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, testUserID, *session.UserID)
	require.Equal(t, testTenantID, session.TenantID)
	require.Equal(t, testRealmID, session.RealmID)
}

func TestIssueSession_Service_HappyPath_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Create test tenant and realm for session
	testTenantID := googleUuid.New()
	testRealmID := googleUuid.New()
	testClientID := "client-" + googleUuid.NewString()

	// Prepare request
	reqBody := SessionIssueRequest{
		UserID:      testClientID,
		TenantID:    testTenantID.String(),
		RealmID:     testRealmID.String(),
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionIssueResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)

	// Verify session was created in database
	ctx := context.Background()
	session, err := testSessionManager.ValidateServiceSession(ctx, result.Token)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, testClientID, *session.ClientID)
	require.Equal(t, testTenantID, session.TenantID)
	require.Equal(t, testRealmID, session.RealmID)
}

func TestIssueSession_InvalidRequestBody_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestIssueSession_InvalidTenantID_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Prepare request with invalid tenant_id
	reqBody := SessionIssueRequest{
		UserID:      "test-user",
		TenantID:    "not-a-uuid",
		RealmID:     googleUuid.NewString(),
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestIssueSession_InvalidRealmID_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Prepare request with invalid realm_id
	reqBody := SessionIssueRequest{
		UserID:      "test-user",
		TenantID:    googleUuid.NewString(),
		RealmID:     "not-a-uuid",
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/issue", handler.IssueSession)

	req := httptest.NewRequest("POST", "/sessions/issue", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

func TestValidateSession_Browser_ValidToken_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Create a browser session first
	ctx := context.Background()
	testTenantID := googleUuid.New()
	testRealmID := googleUuid.New()
	testUserID := "validate-user@example.com"

	token, err := testSessionManager.IssueBrowserSessionWithTenant(ctx, testUserID, testTenantID, testRealmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token
	reqBody := SessionValidateRequest{
		Token:       token,
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, testUserID, result.UserID)
	require.Equal(t, testTenantID.String(), result.TenantID)
	require.Equal(t, testRealmID.String(), result.RealmID)
}

func TestValidateSession_Service_ValidToken_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Create a service session first
	ctx := context.Background()
	testTenantID := googleUuid.New()
	testRealmID := googleUuid.New()
	testClientID := "validate-client-" + googleUuid.NewString()

	token, err := testSessionManager.IssueServiceSessionWithTenant(ctx, testClientID, testTenantID, testRealmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Validate the token
	reqBody := SessionValidateRequest{
		Token:       token,
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.True(t, result.Valid)
	require.Equal(t, testClientID, result.UserID)
	require.Equal(t, testTenantID.String(), result.TenantID)
	require.Equal(t, testRealmID.String(), result.RealmID)
}

func TestValidateSession_Browser_InvalidToken_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Validate non-existent token
	reqBody := SessionValidateRequest{
		Token:       "invalid-token-" + googleUuid.NewString(),
		SessionType: "browser",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Empty(t, result.UserID)
	require.Empty(t, result.TenantID)
	require.Empty(t, result.RealmID)
}

func TestValidateSession_Service_InvalidToken_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	// Validate non-existent token
	reqBody := SessionValidateRequest{
		Token:       "invalid-token-" + googleUuid.NewString(),
		SessionType: "service",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Verify response
	require.Equal(t, 200, resp.StatusCode)

	var result SessionValidateResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, result.Valid)
	require.Empty(t, result.UserID)
	require.Empty(t, result.TenantID)
	require.Empty(t, result.RealmID)
}

func TestValidateSession_InvalidRequestBody_Integration(t *testing.T) {
	if testSessionManager == nil {
		t.Skip("Session manager not initialized")
	}

	t.Parallel()

	handler := NewSessionHandler(testSessionManager)

	app := fiber.New()
	app.Post("/sessions/validate", handler.ValidateSession)

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/sessions/validate", bytes.NewReader([]byte("invalid-json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, 400, resp.StatusCode)
}

