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

func TestGetUserIDFromResponse_Success(t *testing.T) {
	t.Parallel()

	expectedID := googleUuid.New()
	respBody := map[string]string{
		"user_id": expectedID.String(),
	}

	userID, err := cryptoutilAppsTemplateServiceClient.GetUserIDFromResponse(respBody)

	require.NoError(t, err)
	require.Equal(t, expectedID, userID)
}

// TestGetUserIDFromResponse_MissingField verifies error when user_id field absent.
func TestGetUserIDFromResponse_MissingField(t *testing.T) {
	t.Parallel()

	respBody := map[string]string{
		"message": "user created",
	}

	userID, err := cryptoutilAppsTemplateServiceClient.GetUserIDFromResponse(respBody)

	require.Error(t, err)
	require.Equal(t, googleUuid.Nil, userID)
	require.Contains(t, err.Error(), "response missing user_id field")
}

// TestGetUserIDFromResponse_InvalidUUID verifies error for malformed UUID.
func TestGetUserIDFromResponse_InvalidUUID(t *testing.T) {
	t.Parallel()

	respBody := map[string]string{
		"user_id": "not-a-valid-uuid",
	}

	userID, err := cryptoutilAppsTemplateServiceClient.GetUserIDFromResponse(respBody)

	require.Error(t, err)
	require.Equal(t, googleUuid.Nil, userID)
	require.Contains(t, err.Error(), "invalid user_id format")
}

// TestVerifyHealthEndpoint_Success verifies healthy response.
func TestVerifyHealthEndpoint_Success(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/service/api/v1/health", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
		})
	}))
	defer server.Close()

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(server.Client(), server.URL, "/service/api/v1/health")

	require.NoError(t, err)
}

// TestVerifyHealthEndpoint_UnhealthyStatus verifies error for non-healthy status.
func TestVerifyHealthEndpoint_UnhealthyStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "degraded",
		})
	}))
	defer server.Close()

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(server.Client(), server.URL, "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "health status is 'degraded'")
}

// TestVerifyHealthEndpoint_Non200 verifies error for non-200 status code.
func TestVerifyHealthEndpoint_Non200(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("service down"))
	}))
	defer server.Close()

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(server.Client(), server.URL, "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "returned status 503")
}

// TestVerifyHealthEndpoint_MissingStatusField verifies error when status field absent.
func TestVerifyHealthEndpoint_MissingStatusField(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "ok",
		})
	}))
	defer server.Close()

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(server.Client(), server.URL, "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "health response missing status field")
}

// TestRegisterServiceUser_InvalidUserIDInResponse verifies error when user_id is malformed.
func TestRegisterServiceUser_InvalidUserIDInResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == serviceRegisterPath {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": "not-a-valid-uuid",
			})
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid user_id in response")
}

// TestRegisterServiceUser_DecodeResponseError verifies error when response body is invalid JSON.
func TestRegisterServiceUser_DecodeResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == serviceRegisterPath {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not valid json"))
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to decode response")
}

// TestRegisterServiceUser_LoginFailsAfterRegistration verifies error when login fails after successful registration.
func TestRegisterServiceUser_LoginFailsAfterRegistration(t *testing.T) {
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
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid credentials"))
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterServiceUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to login after registration")
}

// TestRegisterBrowserUser_RegistrationFails verifies error handling when registration returns non-201.
func TestRegisterBrowserUser_RegistrationFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("username already exists"))
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(server.Client(), server.URL, "existing", "pass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "registration failed with status 400")
}

// TestRegisterBrowserUser_InvalidUserIDInResponse verifies error when user_id is malformed.
func TestRegisterBrowserUser_InvalidUserIDInResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == browserRegisterPath {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"user_id": "not-a-valid-uuid",
			})
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid user_id in response")
}

// TestRegisterBrowserUser_DecodeResponseError verifies error when response body is invalid JSON.
func TestRegisterBrowserUser_DecodeResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == browserRegisterPath {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("not valid json"))
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to decode response")
}

// TestRegisterBrowserUser_LoginFailsAfterRegistration verifies error when login fails after successful registration.
func TestRegisterBrowserUser_LoginFailsAfterRegistration(t *testing.T) {
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
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("invalid credentials"))
		}
	}))
	defer server.Close()

	user, err := cryptoutilAppsTemplateServiceClient.RegisterBrowserUser(server.Client(), server.URL, "testuser", "testpass")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to login after registration")
}

// TestLoginUser_DecodeResponseError verifies error when response body is invalid JSON.
func TestLoginUser_DecodeResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	token, err := cryptoutilAppsTemplateServiceClient.LoginUser(server.Client(), server.URL, "/login", "user", "pass")

	require.Error(t, err)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "failed to decode response")
}

// TestVerifyHealthEndpoint_DecodeResponseError verifies error when response body is invalid JSON.
func TestVerifyHealthEndpoint_DecodeResponseError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	err := cryptoutilAppsTemplateServiceClient.VerifyHealthEndpoint(server.Client(), server.URL, "/health")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode health response")
}

// TestRegisterServiceUser_ConnectionError verifies error when server is unreachable.
