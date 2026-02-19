// Copyright (c) 2025 Justin Cranford

// Package e2e_helpers provides reusable end-to-end testing helpers for all cryptoutil services.
// Extracted from cipher-im implementation to support 9-service migration.
package e2e_helpers

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// TestUser represents a user with authentication token for E2E testing.
// Reusable across all services implementing user authentication.
type TestUser struct {
	ID       googleUuid.UUID // UUIDv7
	Username string
	Password string
	Token    string // JWT authentication token
}

// RegisterServiceUser registers a user via /service/api/v1/users/register and returns user with token.
// Reusable for all services implementing user registration on /service paths.
func RegisterServiceUser(t *testing.T, client *http.Client, baseURL, username, password string) *TestUser {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	userID, err := googleUuid.Parse(respBody["user_id"])
	require.NoError(t, err)

	// Login to get JWT token.
	token := LoginUser(t, client, baseURL, "/service/api/v1/users/login", username, password)

	return &TestUser{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
	}
}

// RegisterBrowserUser registers a user via /browser/api/v1/users/register and returns user with token.
// Reusable for all services implementing user registration on /browser paths.
func RegisterBrowserUser(t *testing.T, client *http.Client, baseURL, username, password string) *TestUser {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/browser/api/v1/users/register", bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	userID, err := googleUuid.Parse(respBody["user_id"])
	require.NoError(t, err)

	// Login to get JWT token.
	token := LoginUser(t, client, baseURL, "/browser/api/v1/users/login", username, password)

	return &TestUser{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
	}
}

// LoginUser logs in a user at the specified path and returns JWT token.
// Reusable for all services implementing JWT-based authentication.
func LoginUser(t *testing.T, client *http.Client, baseURL, loginPath, username, password string) string {
	t.Helper()

	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+loginPath, bytes.NewReader(reqJSON))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]string

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	require.NoError(t, err)

	return respBody["token"]
}

// RegisterTestUserService registers a test user with randomly generated credentials via /service paths.
// Reusable for all services implementing user registration.
func RegisterTestUserService(t *testing.T, client *http.Client, baseURL string) *TestUser {
	t.Helper()

	username, err := cryptoutilSharedUtilRandom.GenerateUsernameSimple()
	require.NoError(t, err, "failed to generate username")

	password, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	require.NoError(t, err, "failed to generate password")

	return RegisterServiceUser(t, client, baseURL, username, password)
}

// RegisterTestUserBrowser registers a test user with randomly generated credentials via /browser paths.
// Reusable for all services implementing user registration.
func RegisterTestUserBrowser(t *testing.T, client *http.Client, baseURL string) *TestUser {
	t.Helper()

	username, err := cryptoutilSharedUtilRandom.GenerateUsernameSimple()
	require.NoError(t, err, "failed to generate username")

	password, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	require.NoError(t, err, "failed to generate password")

	return RegisterBrowserUser(t, client, baseURL, username, password)
}

// SendAuthenticatedRequest sends an HTTP request with Bearer token authorization.
// Reusable for all services implementing JWT-based authentication.
func SendAuthenticatedRequest(t *testing.T, client *http.Client, method, url, token string, body []byte) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+token)

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// DecodeJSONResponse decodes HTTP response body into provided target struct.
// Reusable for all services returning JSON responses.
func DecodeJSONResponse(t *testing.T, resp *http.Response, target any) {
	t.Helper()

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}

// GetUserIDFromResponse extracts and parses user_id from registration response.
// Reusable for all services returning user_id in registration responses.
func GetUserIDFromResponse(t *testing.T, respBody map[string]string) googleUuid.UUID {
	t.Helper()

	userIDStr, ok := respBody["user_id"]
	require.True(t, ok, "response should contain user_id")

	userID, err := googleUuid.Parse(userIDStr)
	require.NoError(t, err, "user_id should be valid UUID")

	return userID
}

// VerifyHealthEndpoint verifies /service/api/v1/health or /browser/api/v1/health responds with 200 OK.
// Reusable for all services implementing health endpoints.
func VerifyHealthEndpoint(t *testing.T, client *http.Client, baseURL, path string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+path, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err)
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("%s should return 200 OK", path))

	var healthResp map[string]string

	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)

	status, ok := healthResp["status"]
	require.True(t, ok, "health response should contain status field")
	require.Equal(t, "healthy", status, "status should be 'healthy'")
}
