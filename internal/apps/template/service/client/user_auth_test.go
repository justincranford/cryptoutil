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
