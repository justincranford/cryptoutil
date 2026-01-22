// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestHandleRegisterUser_InvalidJSON tests registration with invalid JSON body.
func TestHandleRegisterUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	// Create service with mock repository.
	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	// Create Fiber app with handler.
	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	// Test invalid JSON.
	req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "Invalid request body")
}

// TestHandleRegisterUser_MissingFields tests registration with missing username/password.
func TestHandleRegisterUser_MissingFields(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	tests := []struct {
		name    string
		body    map[string]string
		wantErr string
	}{
		{
			name:    "missing username",
			body:    map[string]string{"password": "SecurePass123!"},
			wantErr: "Username and password are required",
		},
		{
			name:    "missing password",
			body:    map[string]string{"username": "testuser"},
			wantErr: "Username and password are required",
		},
		{
			name:    "empty username",
			body:    map[string]string{"username": "", "password": "SecurePass123!"},
			wantErr: "Username and password are required",
		},
		{
			name:    "empty password",
			body:    map[string]string{"username": "testuser", "password": ""},
			wantErr: "Username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response["error"], tt.wantErr)
		})
	}
}

// TestHandleRegisterUser_UsernameLengthValidation tests username length validation.
func TestHandleRegisterUser_UsernameLengthValidation(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	tests := []struct {
		name     string
		username string
		wantErr  string
	}{
		{
			name:     "username too short",
			username: "ab",
			wantErr:  "username must be 3-50 characters",
		},
		{
			name:     "username too long",
			username: "a" + string(make([]byte, 50)), // 51 characters
			wantErr:  "username must be 3-50 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bodyBytes, err := json.Marshal(map[string]string{
				"username": tt.username,
				"password": "SecurePass123!",
			})
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response["error"], tt.wantErr)
		})
	}
}

// TestHandleRegisterUser_PasswordTooShort tests password length validation.
func TestHandleRegisterUser_PasswordTooShort(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "short", // less than 8 characters
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "password must be at least 8 characters")
}

// TestHandleRegisterUser_Success tests successful user registration.
func TestHandleRegisterUser_Success(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/register", svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "testuser", response["username"])
	require.NotEmpty(t, response["user_id"])
}

// TestHandleLoginUser_InvalidJSON tests login with invalid JSON body.
func TestHandleLoginUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/login", svc.HandleLoginUser("test-secret"))

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "Invalid request body")
}

// TestHandleLoginUser_MissingFields tests login with missing username/password.
func TestHandleLoginUser_MissingFields(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/login", svc.HandleLoginUser("test-secret"))

	tests := []struct {
		name    string
		body    map[string]string
		wantErr string
	}{
		{
			name:    "missing username",
			body:    map[string]string{"password": "SecurePass123!"},
			wantErr: "Username and password are required",
		},
		{
			name:    "missing password",
			body:    map[string]string{"username": "testuser"},
			wantErr: "Username and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response["error"], tt.wantErr)
		})
	}
}

// TestHandleLoginUser_InvalidCredentials tests login with invalid credentials.
func TestHandleLoginUser_InvalidCredentials(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/login", svc.HandleLoginUser("test-secret"))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "nonexistent",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "Invalid credentials")
}

// TestHandleLoginUser_Success tests successful login and JWT generation.
func TestHandleLoginUser_Success(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	// First register a user.
	_, err := svc.RegisterUser(t.Context(), "testuser", "SecurePass123!")
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/login", svc.HandleLoginUser("test-secret-key-12345"))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response["token"])
	require.NotEmpty(t, response["expires_at"])
}

// TestGenerateJWT tests the JWT generation function directly.
func TestGenerateJWT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		userID   string
		username string
		secret   string
		wantErr  bool
	}{
		{
			name:     "valid parameters",
			userID:   "test-user-id",
			username: "testuser",
			secret:   "test-secret-key-12345",
			wantErr:  false,
		},
		{
			name:     "empty secret still works",
			userID:   "test-user-id",
			username: "testuser",
			secret:   "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Use a valid UUIDv7 for the test.
			userID := googleUuid.New()
			token, expiresAt, err := GenerateJWT(userID, tt.username, tt.secret)

			if tt.wantErr {
				require.Error(t, err)
				require.Empty(t, token)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
				require.False(t, expiresAt.IsZero())
			}
		})
	}
}

// TestHandleLoginUserWithSession_InvalidJSON tests session login with invalid JSON body.
func TestHandleLoginUserWithSession_InvalidJSON(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	// Pass nil session manager since we're testing the JSON parse path.
	app.Post("/login/session", svc.HandleLoginUserWithSession(nil, true))

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "Invalid request body")
}

// TestHandleLoginUserWithSession_MissingFields tests session login with missing fields.
func TestHandleLoginUserWithSession_MissingFields(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/login/session", svc.HandleLoginUserWithSession(nil, true))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		// missing password
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandleLoginUserWithSession_InvalidCredentials tests session login with invalid credentials.
func TestHandleLoginUserWithSession_InvalidCredentials(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post("/login/session", svc.HandleLoginUserWithSession(nil, true))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "nonexistent",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

// TestHandleLoginUserWithSession_InvalidSessionManager tests when session manager type assertion fails.
func TestHandleLoginUserWithSession_InvalidSessionManager(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	// Register a user first.
	_, err := svc.RegisterUser(t.Context(), "testuser", "SecurePass123!")
	require.NoError(t, err)

	app := fiber.New()
	// Pass a string (invalid type) as session manager.
	app.Post("/login/session", svc.HandleLoginUserWithSession("not-a-session-manager", true))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response["error"], "Invalid session manager implementation")
}

// TestClaims tests the Claims struct.
func TestClaims(t *testing.T) {
	t.Parallel()

	claims := &Claims{
		UserID:   "test-user-id",
		Username: "testuser",
	}

	require.Equal(t, "test-user-id", claims.UserID)
	require.Equal(t, "testuser", claims.Username)
}
