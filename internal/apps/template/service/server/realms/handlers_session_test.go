// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"context"
	json "encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestHandleLoginUserWithSession_UserWithoutTenantID(t *testing.T) {
	t.Parallel()

	repo := newMockUserRepository()
	// BasicUser doesn't implement GetTenantID, so this will hit the tenantAware type assertion failure.
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	// Register a user first.
	_, err := svc.RegisterUser(t.Context(), "testuser", "SecurePass123!")
	require.NoError(t, err)

	// Create a valid session manager that passes type assertion.
	sessionMgr := &mockSessionManager{token: "test-token-123"}

	app := fiber.New()
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, true))

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

	// Should get 500 because BasicUser doesn't implement GetTenantID.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "User model does not expose tenant ID")
}

// mockTenantAwareUserRepository is a mock repository that returns TenantAwareUser.
type mockTenantAwareUserRepository struct {
	users map[string]UserModel
}

func newMockTenantAwareUserRepository() *mockTenantAwareUserRepository {
	return &mockTenantAwareUserRepository{
		users: make(map[string]UserModel),
	}
}

func (r *mockTenantAwareUserRepository) Create(_ context.Context, user UserModel) error {
	r.users[user.GetUsername()] = user

	return nil
}

func (r *mockTenantAwareUserRepository) FindByUsername(_ context.Context, username string) (UserModel, error) {
	if user, ok := r.users[username]; ok {
		return user, nil
	}

	return nil, nil
}

func (r *mockTenantAwareUserRepository) FindByID(_ context.Context, id googleUuid.UUID) (UserModel, error) {
	for _, user := range r.users {
		if user.GetID() == id {
			return user, nil
		}
	}

	return nil, nil
}

// TestHandleLoginUserWithSession_BrowserSession_Success tests successful browser session login.
func TestHandleLoginUserWithSession_BrowserSession_Success(t *testing.T) {
	t.Parallel()

	repo := newMockTenantAwareUserRepository()
	tenantID := googleUuid.New()
	// Factory returns TenantAwareUser which implements GetTenantID.
	factory := func() UserModel {
		return &TenantAwareUser{
			TenantID: tenantID,
		}
	}
	svc := NewUserService(repo, factory)

	// Register a user first.
	_, err := svc.RegisterUser(t.Context(), "testuser", "SecurePass123!")
	require.NoError(t, err)

	// Create a session manager that returns a token.
	sessionMgr := &mockSessionManager{token: "browser-session-token-123"}

	app := fiber.New()
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, true))

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

	// Should get 200 with token.
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "browser-session-token-123", response[cryptoutilSharedMagic.ParamToken])
	require.NotEmpty(t, response["expires_at"])
}

// TestHandleLoginUserWithSession_ServiceSession_Success tests successful service session login.
func TestHandleLoginUserWithSession_ServiceSession_Success(t *testing.T) {
	t.Parallel()

	repo := newMockTenantAwareUserRepository()
	tenantID := googleUuid.New()
	factory := func() UserModel {
		return &TenantAwareUser{
			TenantID: tenantID,
		}
	}
	svc := NewUserService(repo, factory)

	// Register a user first.
	_, err := svc.RegisterUser(t.Context(), "serviceuser", "SecurePass456!")
	require.NoError(t, err)

	// Create a session manager that returns a token.
	sessionMgr := &mockSessionManager{token: "service-session-token-456"}

	app := fiber.New()
	// isBrowser=false for service session.
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, false))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "serviceuser",
		"password": "SecurePass456!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 200 with token.
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Equal(t, "service-session-token-456", response[cryptoutilSharedMagic.ParamToken])
}

// TestHandleLoginUserWithSession_SessionIssueError tests error during session issuance.
func TestHandleLoginUserWithSession_SessionIssueError(t *testing.T) {
	t.Parallel()

	repo := newMockTenantAwareUserRepository()
	tenantID := googleUuid.New()
	factory := func() UserModel {
		return &TenantAwareUser{
			TenantID: tenantID,
		}
	}
	svc := NewUserService(repo, factory)

	// Register a user first.
	_, err := svc.RegisterUser(t.Context(), "erroruser", "SecurePass789!")
	require.NoError(t, err)

	// Create a session manager that returns an error.
	sessionMgr := &mockSessionManager{issueError: io.EOF}

	app := fiber.New()
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, true))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "erroruser",
		"password": "SecurePass789!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Should get 500 because session issuance failed.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string

	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	require.Contains(t, response[cryptoutilSharedMagic.StringError], "Failed to generate session token")
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
