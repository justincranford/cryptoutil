// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilLearnDomain "cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/server/config"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

// initTestConfig creates a properly configured AppConfig for testing.
func initTestConfig() *config.AppConfig {
	cfg := config.DefaultAppConfig()
	cfg.BindPublicPort = 0                                                          // Dynamic port allocation for tests
	cfg.BindPrivatePort = 0                                                         // Dynamic port allocation for tests
	cfg.OTLPService = "learn-im-test"                                               // Required for telemetry initialization
	cfg.LogLevel = "info"                                                           // Required for logger initialization
	cfg.OTLPEndpoint = "grpc://" + cryptoutilMagic.HostnameLocalhost + ":" + "4317" // Required for OTLP endpoint validation
	cfg.OTLPEnabled = false                                                         // Disable actual OTLP export in tests

	return cfg
}

// createHTTPClient creates an HTTP client that trusts self-signed certificates.
func createHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilMagic.LearnDefaultTimeout, // Increased for concurrent test execution.
	}
}

// testUserWithToken represents a test user with authentication token.
type testUserWithToken struct {
	User  *cryptoutilLearnDomain.User
	Token string
}

// registerAndLoginTestUser registers a user and logs in to get JWT token.
func registerAndLoginTestUser(t *testing.T, client *http.Client, baseURL string) *testUserWithToken {
	t.Helper()

	// Generate random username and password using shared random utilities.
	username, err := cryptoutilRandom.GenerateUsernameSimple()
	require.NoError(t, err)

	password, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	// Register user.
	user := registerTestUser(t, client, baseURL, username, password)

	// Login to get token.
	loginReqBody := map[string]string{
		"username": username,
		"password": password,
	}
	loginReqJSON, err := json.Marshal(loginReqBody)
	require.NoError(t, err)

	loginReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, baseURL+"/service/api/v1/users/login", bytes.NewReader(loginReqJSON))
	require.NoError(t, err)
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := client.Do(loginReq)
	require.NoError(t, err)

	defer func() { _ = loginResp.Body.Close() }()

	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginRespBody map[string]string

	err = json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
	require.NoError(t, err)

	return &testUserWithToken{
		User:  user,
		Token: loginRespBody["token"],
	}
}

// registerTestUser is a helper that registers a user and returns the user domain object.
func registerTestUser(t *testing.T, client *http.Client, baseURL, username, password string) *cryptoutilLearnDomain.User {
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

	return &cryptoutilLearnDomain.User{
		ID:       userID,
		Username: username,
	}
}
