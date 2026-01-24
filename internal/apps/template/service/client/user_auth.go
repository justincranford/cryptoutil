// Copyright (c) 2025 Justin Cranford

// Package client provides reusable client utilities for all cryptoutil services.
// Extracted from E2E testing helpers to support client implementations.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	googleUuid "github.com/google/uuid"

	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// User represents a user with authentication token for client operations.
// Reusable across all services implementing user authentication.
type User struct {
	ID       googleUuid.UUID // UUIDv7
	Username string
	Password string
	Token    string // JWT authentication token
}

// RegisterServiceUser registers a user via /service/api/v1/users/register and returns user with token.
// Reusable for all services implementing user registration on /service paths.
func RegisterServiceUser(client *http.Client, baseURL, username, password string) (*User, error) {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/register", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userID, err := googleUuid.Parse(respBody["user_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid user_id in response: %w", err)
	}

	// Login to get JWT token.
	token, err := LoginUser(client, baseURL, "/service/api/v1/users/login", username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login after registration: %w", err)
	}

	return &User{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
	}, nil
}

// RegisterBrowserUser registers a user via /browser/api/v1/users/register and returns user with token.
// Reusable for all services implementing user registration on /browser paths.
func RegisterBrowserUser(client *http.Client, baseURL, username, password string) (*User, error) {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/browser/api/v1/users/register", bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	userID, err := googleUuid.Parse(respBody["user_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid user_id in response: %w", err)
	}

	// Login to get JWT token.
	token, err := LoginUser(client, baseURL, "/browser/api/v1/users/login", username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to login after registration: %w", err)
	}

	return &User{
		ID:       userID,
		Username: username,
		Password: password,
		Token:    token,
	}, nil
}

// LoginUser logs in a user at the specified path and returns JWT token.
// Reusable for all services implementing JWT-based authentication.
func LoginUser(client *http.Client, baseURL, loginPath, username, password string) (string, error) {
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+loginPath, bytes.NewReader(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	token, ok := respBody["token"]
	if !ok {
		return "", fmt.Errorf("response missing token field")
	}

	return token, nil
}

// RegisterTestUserService registers a test user with randomly generated credentials via /service paths.
// Reusable for all services implementing user registration.
func RegisterTestUserService(client *http.Client, baseURL string) (*User, error) {
	username, err := cryptoutilSharedUtilRandom.GenerateUsernameSimple()
	if err != nil {
		return nil, fmt.Errorf("failed to generate username: %w", err)
	}

	password, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	return RegisterServiceUser(client, baseURL, username, password)
}

// RegisterTestUserBrowser registers a test user with randomly generated credentials via /browser paths.
// Reusable for all services implementing user registration.
func RegisterTestUserBrowser(client *http.Client, baseURL string) (*User, error) {
	username, err := cryptoutilSharedUtilRandom.GenerateUsernameSimple()
	if err != nil {
		return nil, fmt.Errorf("failed to generate username: %w", err)
	}

	password, err := cryptoutilSharedUtilRandom.GeneratePasswordSimple()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	return RegisterBrowserUser(client, baseURL, username, password)
}

// SendAuthenticatedRequest sends an HTTP request with Bearer token authorization.
// Reusable for all services implementing JWT-based authentication.
func SendAuthenticatedRequest(client *http.Client, method, url, token string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

// DecodeJSONResponse decodes HTTP response body into provided target struct.
// Reusable for all services returning JSON responses.
func DecodeJSONResponse(resp *http.Response, target any) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// GetUserIDFromResponse extracts and parses user_id from registration response.
// Reusable for all services returning user_id in registration responses.
func GetUserIDFromResponse(respBody map[string]string) (googleUuid.UUID, error) {
	userIDStr, ok := respBody["user_id"]
	if !ok {
		return googleUuid.Nil, fmt.Errorf("response missing user_id field")
	}

	userID, err := googleUuid.Parse(userIDStr)
	if err != nil {
		return googleUuid.Nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	return userID, nil
}

// VerifyHealthEndpoint verifies /service/api/v1/health or /browser/api/v1/health responds with 200 OK.
// Reusable for all services implementing health endpoints.
func VerifyHealthEndpoint(client *http.Client, baseURL, path string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create health request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send health request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("%s returned status %d: %s", path, resp.StatusCode, string(body))
	}

	var healthResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return fmt.Errorf("failed to decode health response: %w", err)
	}

	status, ok := healthResp["status"]
	if !ok {
		return fmt.Errorf("health response missing status field")
	}

	if status != "healthy" {
		return fmt.Errorf("health status is '%s', expected 'healthy'", status)
	}

	return nil
}
