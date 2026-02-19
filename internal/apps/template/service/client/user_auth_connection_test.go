// Copyright (c) 2025 Justin Cranford
//
//

package client_test

import (
	http "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceClient "cryptoutil/internal/apps/template/service/client"
)

func TestRegisterServiceUser_ConnectionError(t *testing.T) {
	t.Parallel()

	// Use a port that nothing is listening on.
	client := &http.Client{}

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(client, "http://127.0.0.1:59999", "user", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to send request")
}

// TestRegisterBrowserUser_ConnectionError verifies error when server is unreachable.
func TestRegisterBrowserUser_ConnectionError(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(client, "http://127.0.0.1:59999", "user", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to send request")
}

// TestLoginUser_ConnectionError verifies error when server is unreachable.
func TestLoginUser_ConnectionError(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(client, "http://127.0.0.1:59999", "/login", "user", "pass")

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "failed to send request")
}

// TestSendAuthenticatedRequest_ConnectionError verifies error when server is unreachable.
func TestSendAuthenticatedRequest_ConnectionError(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodGet, "http://127.0.0.1:59999/api", "token", nil) //nolint:bodyclose // resp is nil on connection error.

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to send request")
}

// TestVerifyHealthEndpoint_ConnectionError verifies error when server is unreachable.
func TestVerifyHealthEndpoint_ConnectionError(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(client, "http://127.0.0.1:59999", "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to send health request")
}

// TestRegisterServiceUser_InvalidURL verifies error when URL is malformed.
func TestRegisterServiceUser_InvalidURL(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	// Use invalid URL with control characters to trigger NewRequestWithContext error.
	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(client, "http://invalid\x00url", "user", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to create request")
}

// TestRegisterBrowserUser_InvalidURL verifies error when URL is malformed.
func TestRegisterBrowserUser_InvalidURL(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(client, "http://invalid\x00url", "user", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to create request")
}

// TestLoginUser_InvalidURL verifies error when URL is malformed.
func TestLoginUser_InvalidURL(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(client, "http://invalid\x00url", "/login", "user", "pass")

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "failed to create request")
}

// TestSendAuthenticatedRequest_InvalidURL verifies error when URL is malformed.
func TestSendAuthenticatedRequest_InvalidURL(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(client, http.MethodGet, "http://invalid\x00url/api", "token", nil) //nolint:bodyclose // resp is nil on invalid URL.

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "failed to create request")
}

// TestVerifyHealthEndpoint_InvalidURL verifies error when URL is malformed.
func TestVerifyHealthEndpoint_InvalidURL(t *testing.T) {
	t.Parallel()

	client := &http.Client{}

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(client, "http://invalid\x00url", "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create health request")
}
