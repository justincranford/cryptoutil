// Copyright (c) 2025-2026 Justin Cranford.
// Licensed under the MIT License. See LICENSE file in the project root for license information.

//go:build !integration

package apis

import (
	"bytes"
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServerRepository "cryptoutil/internal/apps-framework/service/server/repository"
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

// TestIssueSession_ValidationErrors tests error handling using table-driven pattern.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
// Tests focus on early validation failures before reaching session manager.
func TestIssueSession_ValidationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		requestBody any
		setupApp    func() *fiber.App
		wantErr     bool
	}{
		{
			name:        testInvalidJSONCaseName,
			requestBody: "not-json",
			setupApp: func() *fiber.App {
				handler := NewSessionHandler(nil)
				app := fiber.New()
				app.Post("/session/issue", handler.IssueSession)

				return app
			},
			wantErr: true,
		},
		{
			name: "invalid tenant ID",
			requestBody: SessionIssueRequest{
				UserID:      testUser123,
				TenantID:    testNotAValidUUID,
				RealmID:     googleUuid.New().String(),
				SessionType: sessionTypeBrowser,
			},
			setupApp: func() *fiber.App {
				handler := NewSessionHandler(nil)
				app := fiber.New()
				app.Post("/session/issue", handler.IssueSession)

				return app
			},
			wantErr: true,
		},
		{
			name: "invalid realm ID",
			requestBody: SessionIssueRequest{
				UserID:      testUser123,
				TenantID:    googleUuid.New().String(),
				RealmID:     testNotAValidUUID,
				SessionType: sessionTypeBrowser,
			},
			setupApp: func() *fiber.App {
				handler := NewSessionHandler(nil)
				app := fiber.New()
				app.Post("/session/issue", handler.IssueSession)

				return app
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			app := tt.setupApp()

			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/session/issue", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			if tt.wantErr {
				require.True(t, resp.StatusCode >= 400, "Expected error status code")
			}
		})
	}
}

// TestValidateSession_InvalidJSON tests ValidateSession with invalid JSON body.
// NOTE: The handler uses cryptoutilAppErr types which Fiber's default error handler converts to 500.
func TestValidateSession_InvalidJSON(t *testing.T) {
	t.Parallel()

	handler := NewSessionHandler(nil)
	app := fiber.New()
	app.Post("/session/validate", handler.ValidateSession)

	req := httptest.NewRequest("POST", "/session/validate", bytes.NewReader([]byte(testInvalidJsonValue9E1FF427)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
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
	t.Parallel()

	tests := []struct {
		name        string
		requestBody string
		wantError   bool
	}{
		{
			name:        testEmptyBodyReturnsErrorDueToParseError,
			requestBody: "",
			wantError:   true,
		},
		{
			name:        testInvalidJsonReturnsError,
			requestBody: testInvalid,
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

			resp, err := app.Test(req, -1)
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
	t.Parallel()

	tests := []struct {
		name        string
		requestBody string
		wantError   bool
	}{
		{
			name:        testEmptyBodyReturnsErrorDueToParseError,
			requestBody: "",
			wantError:   true,
		},
		{
			name:        testInvalidJsonReturnsError,
			requestBody: testInvalid,
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

			resp, err := app.Test(req, -1)
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

const (
	testSessionIssuePath       = "/session/issue"
	testSessionValidatePath    = "/session/validate"
	testSessionServiceClientID = "test-service-session-client"
)

func TestIssueSession_ServiceSessionType(t *testing.T) {
	t.Parallel()

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	realmID, err := googleUuid.NewV7()
	require.NoError(t, err)

	var (
		issuedClientID string
		issuedTenantID googleUuid.UUID
		issuedRealmID  googleUuid.UUID
	)

	sessionManager := newMockSessionManagerSuccess()
	sessionManager.issueServiceFunc = func(
		_ context.Context,
		clientID string,
		gotTenantID googleUuid.UUID,
		gotRealmID googleUuid.UUID,
	) (string, error) {
		issuedClientID = clientID
		issuedTenantID = gotTenantID
		issuedRealmID = gotRealmID

		return testServiceSessionToken, nil
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post(testSessionIssuePath, NewSessionHandler(sessionManager).IssueSession)

	request := newSessionJSONRequest(t, http.MethodPost, testSessionIssuePath, SessionIssueRequest{
		UserID:      testSessionServiceClientID,
		TenantID:    tenantID.String(),
		RealmID:     realmID.String(),
		SessionType: testSessionTypeService,
	})
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var actual SessionIssueResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&actual))
	require.Equal(t, testServiceSessionToken, actual.Token)
	require.Equal(t, testSessionServiceClientID, issuedClientID)
	require.Equal(t, tenantID, issuedTenantID)
	require.Equal(t, realmID, issuedRealmID)
}

func TestValidateSession_ServiceSessionType(t *testing.T) {
	t.Parallel()

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	realmID, err := googleUuid.NewV7()
	require.NoError(t, err)

	var validatedToken string

	sessionManager := newMockSessionManagerSuccess()
	sessionManager.validateServiceFunc = func(
		_ context.Context,
		token string,
	) (*cryptoutilAppsFrameworkServiceServerRepository.ServiceSession, error) {
		validatedToken = token
		clientID := testSessionServiceClientID

		return &cryptoutilAppsFrameworkServiceServerRepository.ServiceSession{
			ClientID: &clientID,
			Session: cryptoutilAppsFrameworkServiceServerRepository.Session{
				TenantID: tenantID,
				RealmID:  realmID,
			},
		}, nil
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post(testSessionValidatePath, NewSessionHandler(sessionManager).ValidateSession)

	request := newSessionJSONRequest(t, http.MethodPost, testSessionValidatePath, SessionValidateRequest{
		Token:       testServiceSessionToken,
		SessionType: testSessionTypeService,
	})
	response, err := app.Test(request, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, response.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, response.StatusCode)

	var actual SessionValidateResponse
	require.NoError(t, json.NewDecoder(response.Body).Decode(&actual))
	require.True(t, actual.Valid)
	require.Equal(t, testSessionServiceClientID, actual.UserID)
	require.Equal(t, tenantID.String(), actual.TenantID)
	require.Equal(t, realmID.String(), actual.RealmID)
	require.Equal(t, testServiceSessionToken, validatedToken)
}

func TestIssueAndValidateSession_Integration(t *testing.T) {
	t.Parallel()

	tenantID, err := googleUuid.NewV7()
	require.NoError(t, err)
	realmID, err := googleUuid.NewV7()
	require.NoError(t, err)

	var (
		issuedClientID string
		issuedTenantID googleUuid.UUID
		issuedRealmID  googleUuid.UUID
		validatedToken string
	)

	sessionManager := newMockSessionManagerSuccess()
	sessionManager.issueServiceFunc = func(
		_ context.Context,
		clientID string,
		gotTenantID googleUuid.UUID,
		gotRealmID googleUuid.UUID,
	) (string, error) {
		issuedClientID = clientID
		issuedTenantID = gotTenantID
		issuedRealmID = gotRealmID

		return testServiceSessionToken, nil
	}
	sessionManager.validateServiceFunc = func(
		_ context.Context,
		token string,
	) (*cryptoutilAppsFrameworkServiceServerRepository.ServiceSession, error) {
		validatedToken = token
		clientID := issuedClientID

		return &cryptoutilAppsFrameworkServiceServerRepository.ServiceSession{
			ClientID: &clientID,
			Session: cryptoutilAppsFrameworkServiceServerRepository.Session{
				TenantID: issuedTenantID,
				RealmID:  issuedRealmID,
			},
		}, nil
	}

	handler := NewSessionHandler(sessionManager)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post(testSessionIssuePath, handler.IssueSession)
	app.Post(testSessionValidatePath, handler.ValidateSession)

	issueRequest := newSessionJSONRequest(t, http.MethodPost, testSessionIssuePath, SessionIssueRequest{
		UserID:      testSessionServiceClientID,
		TenantID:    tenantID.String(),
		RealmID:     realmID.String(),
		SessionType: testSessionTypeService,
	})
	issueResponse, err := app.Test(issueRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, issueResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, issueResponse.StatusCode)

	var issued SessionIssueResponse
	require.NoError(t, json.NewDecoder(issueResponse.Body).Decode(&issued))
	require.Equal(t, testServiceSessionToken, issued.Token)

	validateRequest := newSessionJSONRequest(t, http.MethodPost, testSessionValidatePath, SessionValidateRequest{
		Token:       issued.Token,
		SessionType: testSessionTypeService,
	})
	validateResponse, err := app.Test(validateRequest, -1)

	require.NoError(t, err)
	defer func() { require.NoError(t, validateResponse.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, validateResponse.StatusCode)

	var validated SessionValidateResponse
	require.NoError(t, json.NewDecoder(validateResponse.Body).Decode(&validated))
	require.True(t, validated.Valid)
	require.Equal(t, testSessionServiceClientID, validated.UserID)
	require.Equal(t, tenantID.String(), validated.TenantID)
	require.Equal(t, realmID.String(), validated.RealmID)
	require.Equal(t, issued.Token, validatedToken)
}

func newSessionJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()

	encodedBody, err := json.Marshal(body)
	require.NoError(t, err)

	request := httptest.NewRequestWithContext(
		t.Context(),
		method,
		path,
		bytes.NewReader(encodedBody),
	)
	request.Header.Set("Content-Type", "application/json")

	return request
}
