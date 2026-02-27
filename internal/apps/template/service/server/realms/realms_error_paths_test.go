// Copyright (c) 2025 Justin Cranford
//

package realms

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestRegisterUser_HashError tests the error path when PBKDF2 hashing fails.
// Cannot be parallel because it modifies a package-level variable.
func TestRegisterUser_HashError(t *testing.T) {
	orig := realmsServiceHashSecretPBKDF2Fn
	realmsServiceHashSecretPBKDF2Fn = func(_ string) (string, error) {
		return "", errors.New("injected hash failure")
	}

	defer func() { realmsServiceHashSecretPBKDF2Fn = orig }()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	_, err := svc.RegisterUser(t.Context(), "validuser", "SecurePass123!")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash password")
}

// TestHandleLoginUser_GenerateJWTError tests the error path when JWT generation fails.
// Cannot be parallel because it modifies a package-level variable.
func TestHandleLoginUser_GenerateJWTError(t *testing.T) {
	orig := realmsHandlersGenerateJWTFn
	realmsHandlersGenerateJWTFn = func(_ googleUuid.UUID, _ string, _ string) (string, time.Time, error) {
		return "", time.Time{}, errors.New("injected jwt failure")
	}

	defer func() { realmsHandlersGenerateJWTFn = orig }()

	repo := newMockUserRepository()
	factory := func() UserModel { return &BasicUser{} }
	svc := NewUserService(repo, factory)

	// Register user first.
	_, err := svc.RegisterUser(t.Context(), "jwtuser", "SecurePass123!")
	require.NoError(t, err)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/login", svc.HandleLoginUser("test-secret"))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "jwtuser",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

// mockSessionManagerWithRealm implements sessionIssuer and realmServiceProvider.
type mockSessionManagerWithRealm struct {
	issueError error
	token      string
	realmSvc   *mockRealmLookup
}

func (m *mockSessionManagerWithRealm) IssueBrowserSessionWithTenant(_ context.Context, _ string, _ googleUuid.UUID, _ googleUuid.UUID) (string, error) {
	if m.issueError != nil {
		return "", m.issueError
	}

	return m.token, nil
}

func (m *mockSessionManagerWithRealm) IssueServiceSessionWithTenant(_ context.Context, _ string, _ googleUuid.UUID, _ googleUuid.UUID) (string, error) {
	if m.issueError != nil {
		return "", m.issueError
	}

	return m.token, nil
}

func (m *mockSessionManagerWithRealm) GetRealmService() realmLookup {
	return m.realmSvc
}

// mockRealmLookup implements realmLookup.
type mockRealmLookup struct {
	realm    any
	realmErr error
}

func (m *mockRealmLookup) GetFirstActiveRealm(_ context.Context, _ googleUuid.UUID) (any, error) {
	return m.realm, m.realmErr
}

// mockRealmWithID implements realmIDGetter.
type mockRealmWithID struct {
	id googleUuid.UUID
}

func (r *mockRealmWithID) GetRealmID() googleUuid.UUID {
	return r.id
}

// TestHandleLoginUserWithSession_RealmLookupError tests graceful fallback when realm lookup fails.
func TestHandleLoginUserWithSession_RealmLookupError(t *testing.T) {
	t.Parallel()

	repo := newMockTenantAwareUserRepository()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	factory := func() UserModel {
		return &TenantAwareUser{
			BasicUser: BasicUser{},
			TenantID:  tenantID,
		}
	}
	svc := NewUserService(repo, factory)

	// Register user with TenantAwareUser factory.
	_, err := svc.RegisterUser(t.Context(), "tenantuser1", "SecurePass123!")
	require.NoError(t, err)

	// Session manager with realm service that returns an error.
	sessionMgr := &mockSessionManagerWithRealm{
		token: "test-session-token",
		realmSvc: &mockRealmLookup{
			realmErr: errors.New("realm lookup failure"),
		},
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, true))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "tenantuser1",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	// Graceful fallback: session should still be issued.
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestHandleLoginUserWithSession_RealmLookupSuccess tests that realm ID is used when realm lookup succeeds.
func TestHandleLoginUserWithSession_RealmLookupSuccess(t *testing.T) {
	t.Parallel()

	repo := newMockTenantAwareUserRepository()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())
	factory := func() UserModel {
		return &TenantAwareUser{
			BasicUser: BasicUser{},
			TenantID:  tenantID,
		}
	}
	svc := NewUserService(repo, factory)

	// Register user with TenantAwareUser factory.
	_, err := svc.RegisterUser(t.Context(), "tenantuser2", "SecurePass123!")
	require.NoError(t, err)

	// Session manager with realm service that returns a valid realm.
	sessionMgr := &mockSessionManagerWithRealm{
		token: "test-session-with-realm",
		realmSvc: &mockRealmLookup{
			realm: &mockRealmWithID{id: realmID},
		},
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Post("/login/session", svc.HandleLoginUserWithSession(sessionMgr, false))

	bodyBytes, err := json.Marshal(map[string]string{
		"username": "tenantuser2",
		"password": "SecurePass123!",
	})
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/login/session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { require.NoError(t, resp.Body.Close()) }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}
