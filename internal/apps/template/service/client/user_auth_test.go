// Copyright (c) 2025 Justin Cranford
//
//

package client_test

import (
	json "encoding/json"
	http "net/http"
	httptest "net/http/httptest"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceClient "cryptoutil/internal/apps/template/service/client"
)

// Test path constants.
const (
	serviceRegisterPath = "/service/api/v1/users/register"
	serviceLoginPath    = "/service/api/v1/users/login"
	browserRegisterPath = "/browser/api/v1/users/register"
	browserLoginPath    = "/browser/api/v1/users/login"
)

// TestRegisterServiceUser_Success verifies successful user registration via /service path.
func TestRegisterServiceUser_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()
	token := "test-jwt-token"

	// Mock server for registration and login.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case serviceRegisterPath:
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "application/json", r.Header.Get("Content-Type"))

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": userID.String(),
			})

		case serviceLoginPath:
			require.Equal(t, http.MethodPost, r.Method)

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": token,
			})

		default:
			require.Failf(t, "unexpected path", "unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(server.Client(), server.URL, "testuser", "testpass")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.Equal(t, "testuser", user.Username)
	require.Equal(t, "testpass", user.Password)
	require.Equal(t, token, user.Token)
}

// TestRegisterServiceUser_RegistrationFails verifies error handling when registration returns non-201.
func TestRegisterServiceUser_RegistrationFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("username already exists"))
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(server.Client(), server.URL, "existing", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "registration failed with status 400")
}

// TestRegisterBrowserUser_Success verifies successful user registration via /browser path.
func TestRegisterBrowserUser_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()
	token := "browser-jwt-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case browserRegisterPath:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": userID.String(),
			})

		case browserLoginPath:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": token,
			})

		default:
			require.Failf(t, "unexpected path", "unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(server.Client(), server.URL, "browseruser", "browserpass")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, userID, user.ID)
	require.Equal(t, "browseruser", user.Username)
	require.Equal(t, "browserpass", user.Password)
	require.Equal(t, token, user.Token)
}

// TestLoginUser_Success verifies successful login with JWT token.
func TestLoginUser_Success(t *testing.T) {
	t.Parallel()

	expectedToken := "valid-jwt-token"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/custom/login", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"token": expectedToken,
		})
	}))
	defer server.Close()

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(server.Client(), server.URL, "/custom/login", "user", "pass")

	require.NoError(t, err)
	require.Equal(t, expectedToken, token)
}

// TestLoginUser_InvalidCredentials verifies error handling for login failures.
func TestLoginUser_InvalidCredentials(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid credentials"))
	}))
	defer server.Close()

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(server.Client(), server.URL, "/login", "wrong", "creds")

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "login failed with status 401")
}

// TestLoginUser_MissingToken verifies error when response lacks token field.
func TestLoginUser_MissingToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "login successful",
		})
	}))
	defer server.Close()

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(server.Client(), server.URL, "/login", "user", "pass")

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "response missing token field")
}

// TestRegisterTestUserService_Success verifies registration with random credentials via /service paths.
func TestRegisterTestUserService_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case serviceRegisterPath:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": userID.String(),
			})
		case serviceLoginPath:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": "random-user-token",
			})
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterTestUserService(server.Client(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotEqual(t, googleUuid.Nil, user.ID)
	require.NotEmpty(t, user.Username)
	require.NotEmpty(t, user.Password)
	require.NotEmpty(t, user.Token)
}

// TestRegisterTestUserBrowser_Success verifies registration with random credentials via /browser paths.
func TestRegisterTestUserBrowser_Success(t *testing.T) {
	t.Parallel()

	userID := googleUuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case browserRegisterPath:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": userID.String(),
			})
		case browserLoginPath:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"token": "browser-random-token",
			})
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterTestUserBrowser(server.Client(), server.URL)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotEmpty(t, user.Username)
	require.NotEmpty(t, user.Token)
}

// TestSendAuthenticatedRequest_Success verifies authenticated request with Bearer token.
func TestSendAuthenticatedRequest_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	body := []byte(`{"key":"value"}`)
	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(server.Client(), http.MethodPost, server.URL+"/api", "test-token", body)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

// TestSendAuthenticatedRequest_NoBody verifies request without body.
func TestSendAuthenticatedRequest_NoBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		require.Equal(t, "", r.Header.Get("Content-Type")) // No Content-Type for GET.

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := cryptoutilAppsTemplateServiceClient.SendAuthenticatedRequest(server.Client(), http.MethodGet, server.URL+"/data", "token123", nil)

	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = resp.Body.Close()
}

// TestDecodeJSONResponse_Success verifies JSON decoding.
func TestDecodeJSONResponse_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"field1": "value1",
			"field2": "value2",
		})
	}))
	defer server.Close()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, _ := http.DefaultClient.Do(req) //nolint:bodyclose // DecodeJSONResponse closes body internally.

	var target map[string]string

	err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &target)

	require.NoError(t, err)
	require.Equal(t, "value1", target["field1"])
	require.Equal(t, "value2", target["field2"])
}

// TestDecodeJSONResponse_InvalidJSON verifies error handling for malformed JSON.
func TestDecodeJSONResponse_InvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	resp, _ := http.DefaultClient.Do(req) //nolint:bodyclose // DecodeJSONResponse closes body internally.

	var target map[string]string

	err := cryptoutilAppsTemplateServiceClient.DecodeJSONResponse(resp, &target)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

// TestGetUserIDFromResponse_Success verifies user_id extraction and parsing.
