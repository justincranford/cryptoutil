// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "Invalid credentials")
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.NotEmpty(t, response[cryptoutilSharedMagic.ParamToken])
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

	resp, err := app.Test(req, -1)
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

	resp, err := app.Test(req, -1)
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "Invalid session manager implementation")
}

// mockSessionManager implements the sessionIssuer interface for testing.
type mockSessionManager struct {
	issueError error
	token      string
}

func (m *mockSessionManager) IssueBrowserSessionWithTenant(_ context.Context, _ string, _ googleUuid.UUID, _ googleUuid.UUID) (string, error) {
	if m.issueError != nil {
		return "", m.issueError
	}

	return m.token, nil
}

func (m *mockSessionManager) IssueServiceSessionWithTenant(_ context.Context, _ string, _ googleUuid.UUID, _ googleUuid.UUID) (string, error) {
	if m.issueError != nil {
		return "", m.issueError
	}

	return m.token, nil
}

// TenantAwareUser is a user model that implements GetTenantID for testing.
type TenantAwareUser struct {
	BasicUser
	TenantID googleUuid.UUID
}

// GetTenantID returns the user's tenant ID.
func (u *TenantAwareUser) GetTenantID() googleUuid.UUID {
	return u.TenantID
}

// TestHandleLoginUserWithSession_UserWithoutTenantID tests when user model doesn't implement GetTenantID.
