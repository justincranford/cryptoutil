// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	json "encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
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
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	// Test invalid JSON.
	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "Invalid request body")
}

// TestHandleRegisterUser_MissingFields tests registration with missing username/password.
func TestHandleRegisterUser_MissingFields(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

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

			req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response[cryptoutilSharedMagic.StringError], tt.wantErr)
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
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

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
			username: "a" + string(make([]byte, cryptoutilSharedMagic.IMMaxUsernameLength)), // 51 characters
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

			req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response[cryptoutilSharedMagic.StringError], tt.wantErr)
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
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "short", // less than 8 characters
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "password must be at least 8 characters")
}

// TestHandleRegisterUser_Success tests successful user registration.
func TestHandleRegisterUser_Success(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]any

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "testuser", response["username"])
	require.NotEmpty(t, response["user_id"])
}

// TestHandleRegisterUser_DuplicateUsernameSQLite tests 409 Conflict for SQLite duplicate username.
func TestHandleRegisterUser_DuplicateUsernameSQLite(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	// Set the createErr to simulate SQLite unique constraint violation.
	repo.createErr = fmt.Errorf("UNIQUE constraint failed: users.username")
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "existinguser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 409 Conflict for duplicate username.
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "Username already exists", response[cryptoutilSharedMagic.StringError])
}

// TestHandleRegisterUser_DuplicateUsernamePostgreSQL tests 409 Conflict for PostgreSQL duplicate username.
func TestHandleRegisterUser_DuplicateUsernamePostgreSQL(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	// Set the createErr to simulate PostgreSQL unique constraint violation.
	repo.createErr = fmt.Errorf("duplicate key value violates unique constraint \"users_username_key\"")
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "existinguser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 409 Conflict for duplicate username.
	require.Equal(t, fiber.StatusConflict, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "Username already exists", response[cryptoutilSharedMagic.StringError])
}

// TestHandleRegisterUser_GenericError tests 500 Internal Server Error for non-duplicate errors.
func TestHandleRegisterUser_GenericError(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	// Set the createErr to simulate a generic database error (not duplicate).
	repo.createErr = fmt.Errorf("database connection lost")
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	app := fiber.New()
	app.Post(cryptoutilSharedMagic.PathRegistration, svc.HandleRegisterUser())

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "newuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", cryptoutilSharedMagic.PathRegistration, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 500 Internal Server Error for generic errors.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "Failed to create user", response[cryptoutilSharedMagic.StringError])
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "Invalid request body")
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

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var response map[string]string

			err = json.Unmarshal(body, &response)
			require.NoError(t, err)
			require.Contains(t, response[cryptoutilSharedMagic.StringError], tt.wantErr)
		})
	}
}

// TestHandleLoginUser_InvalidCredentials tests login with invalid credentials.
