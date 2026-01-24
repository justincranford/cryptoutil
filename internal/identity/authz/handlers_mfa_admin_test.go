// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityAuthz "cryptoutil/internal/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// createMFAAdminTestDependencies creates test dependencies for MFA admin tests.
func createMFAAdminTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  "file::memory:?cache=private",
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://localhost:8080",
		},
	}

	ctx := context.Background()
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, config.Database)
	require.NoError(t, err, "Failed to create repository factory")
	require.NotNil(t, repoFactory, "Repository factory should not be nil")

	// Get GORM DB instance for AutoMigrate.
	db := repoFactory.DB()

	// Auto-migrate all required tables for MFA admin tests.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.AuthProfile{},
		&cryptoutilIdentityDomain.MFAFactor{},
	)
	require.NoError(t, err, "Failed to auto-migrate database tables")

	return config, repoFactory
}

// TestHandleEnrollMFA_HappyPath tests successful MFA factor enrollment.
func TestHandleEnrollMFA_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	// Create test user.
	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Prepare enrollment request.
	reqBody := map[string]any{
		"user_id":     user.ID.String(),
		"factor_type": "totp",
		"name":        "My TOTP",
		"required":    true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/enroll", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var enrollResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&enrollResp))
	require.NotEmpty(t, enrollResp["id"])
	require.Equal(t, "totp", enrollResp["factor_type"])
	require.Equal(t, "My TOTP", enrollResp["name"])
	require.Equal(t, true, enrollResp["required"])
	require.Equal(t, true, enrollResp["enabled"])
	require.NotEmpty(t, enrollResp["created_at"])
}

// TestHandleEnrollMFA_InvalidUserID tests enrollment with invalid user_id.
func TestHandleEnrollMFA_InvalidUserID(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := map[string]any{
		"user_id":     "not-a-uuid",
		"factor_type": "totp",
		"name":        "My TOTP",
		"required":    true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/enroll", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid_request", errResp["error"])
	require.Contains(t, errResp["error_description"], "invalid user_id format")
}

// TestHandleEnrollMFA_UserNotFound tests enrollment with non-existent user_id.
func TestHandleEnrollMFA_UserNotFound(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := map[string]any{
		"user_id":     googleUuid.Must(googleUuid.NewV7()).String(),
		"factor_type": "totp",
		"name":        "My TOTP",
		"required":    false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/enroll", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "user_not_found", errResp["error"])
}

// TestHandleEnrollMFA_InvalidFactorType tests enrollment with invalid factor_type.
func TestHandleEnrollMFA_InvalidFactorType(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	reqBody := map[string]any{
		"user_id":     user.ID.String(),
		"factor_type": "invalid_type",
		"name":        "My Factor",
		"required":    false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/oidc/v1/mfa/enroll", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid_request", errResp["error"])
	require.Contains(t, errResp["error_description"], "invalid factor_type")
}

// TestHandleListMFAFactors_HappyPath tests listing MFA factors for a user.
func TestHandleListMFAFactors_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	// Create test user.
	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create auth profile with naming convention.
	authProfileRepo := repoFactory.AuthProfileRepository()
	authProfile := &cryptoutilIdentityDomain.AuthProfile{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        fmt.Sprintf("user_%s_default", user.ID.String()),
		Description: "Test profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		Enabled:     true,
	}
	require.NoError(t, authProfileRepo.Create(ctx, authProfile))

	// Create MFA factors.
	mfaFactorRepo := repoFactory.MFAFactorRepository()
	factor1 := &cryptoutilIdentityDomain.MFAFactor{
		ID:            googleUuid.Must(googleUuid.NewV7()),
		Name:          "TOTP Factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		AuthProfileID: authProfile.ID,
		Required:      cryptoutilIdentityDomain.IntBool(true),
		Enabled:       true,
		Order:         1,
	}
	require.NoError(t, mfaFactorRepo.Create(ctx, factor1))

	factor2 := &cryptoutilIdentityDomain.MFAFactor{
		ID:            googleUuid.Must(googleUuid.NewV7()),
		Name:          "Email OTP Factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeEmailOTP,
		AuthProfileID: authProfile.ID,
		Required:      cryptoutilIdentityDomain.IntBool(false),
		Enabled:       true,
		Order:         2,
	}
	require.NoError(t, mfaFactorRepo.Create(ctx, factor2))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", fmt.Sprintf("/oidc/v1/mfa/factors?user_id=%s", user.ID.String()), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var listResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))

	factors, ok := listResp["factors"].([]any)
	require.True(t, ok, "factors should be an array")
	require.Len(t, factors, 2, "should return 2 factors")

	// Verify first factor.
	f1, ok := factors[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "TOTP Factor", f1["name"])
	require.Equal(t, "totp", f1["factor_type"])
	require.Equal(t, true, f1["required"])
	require.Equal(t, true, f1["enabled"])

	// Verify second factor.
	f2, ok := factors[1].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Email OTP Factor", f2["name"])
	require.Equal(t, "email_otp", f2["factor_type"])
	require.Equal(t, false, f2["required"])
	require.Equal(t, true, f2["enabled"])
}

// TestHandleListMFAFactors_NoFactors tests listing when user has no MFA factors.
func TestHandleListMFAFactors_NoFactors(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", fmt.Sprintf("/oidc/v1/mfa/factors?user_id=%s", user.ID.String()), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	var listResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listResp))

	factors, ok := listResp["factors"].([]any)
	require.True(t, ok)
	require.Len(t, factors, 0, "should return empty array")
}

// TestHandleListMFAFactors_InvalidUserID tests listing with invalid user_id.
func TestHandleListMFAFactors_InvalidUserID(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("GET", "/oidc/v1/mfa/factors?user_id=not-a-uuid", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid_request", errResp["error"])
	require.Contains(t, errResp["error_description"], "invalid user_id format")
}

// TestHandleDeleteMFAFactor_HappyPath tests successful MFA factor deletion.
func TestHandleDeleteMFAFactor_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	// Create test user.
	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	// Create auth profile with naming convention.
	authProfileRepo := repoFactory.AuthProfileRepository()
	authProfile := &cryptoutilIdentityDomain.AuthProfile{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        fmt.Sprintf("user_%s_default", user.ID.String()),
		Description: "Test profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		Enabled:     true,
	}
	require.NoError(t, authProfileRepo.Create(ctx, authProfile))

	// Create MFA factor.
	mfaFactorRepo := repoFactory.MFAFactorRepository()
	factor := &cryptoutilIdentityDomain.MFAFactor{
		ID:            googleUuid.Must(googleUuid.NewV7()),
		Name:          "TOTP Factor",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		AuthProfileID: authProfile.ID,
		Required:      cryptoutilIdentityDomain.IntBool(true),
		Enabled:       true,
		Order:         1,
	}
	require.NoError(t, mfaFactorRepo.Create(ctx, factor))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/oidc/v1/mfa/factors/%s?user_id=%s", factor.ID.String(), user.ID.String()), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	// Verify factor is soft-deleted (should return error).
	deletedFactor, err := mfaFactorRepo.GetByID(ctx, factor.ID)
	require.Error(t, err, "Should error when getting deleted factor")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrMFAFactorNotFound)
	require.Nil(t, deletedFactor, "Deleted factor should be nil")
}

// TestHandleDeleteMFAFactor_FactorNotFound tests deletion with non-existent factor_id.
func TestHandleDeleteMFAFactor_FactorNotFound(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	userRepo := repoFactory.UserRepository()
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          googleUuid.Must(googleUuid.NewV7()).String(),
		Email:        fmt.Sprintf("test-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash: "hash",
	}
	require.NoError(t, userRepo.Create(ctx, user))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	nonExistentID := googleUuid.Must(googleUuid.NewV7())
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/oidc/v1/mfa/factors/%s?user_id=%s", nonExistentID.String(), user.ID.String()), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "factor_not_found", errResp["error"])
}

// TestHandleDeleteMFAFactor_Unauthorized tests deletion when factor doesn't belong to user.
func TestHandleDeleteMFAFactor_Unauthorized(t *testing.T) {
	t.Parallel()

	config, repoFactory := createMFAAdminTestDependencies(t)
	ctx := context.Background()

	// Create two test users.
	userRepo := repoFactory.UserRepository()

	ownerUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		Email:             fmt.Sprintf("owner-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PreferredUsername: fmt.Sprintf("owner-%s", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash:      "hash",
	}
	require.NoError(t, userRepo.Create(ctx, ownerUser))

	otherUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		Email:             fmt.Sprintf("other-%s@example.com", googleUuid.Must(googleUuid.NewV7()).String()),
		PreferredUsername: fmt.Sprintf("other-%s", googleUuid.Must(googleUuid.NewV7()).String()),
		PasswordHash:      "hash",
	}
	require.NoError(t, userRepo.Create(ctx, otherUser))

	// Create auth profile for owner.
	authProfileRepo := repoFactory.AuthProfileRepository()
	authProfile := &cryptoutilIdentityDomain.AuthProfile{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        fmt.Sprintf("user_%s_default", ownerUser.ID.String()),
		Description: "Owner profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		Enabled:     true,
	}
	require.NoError(t, authProfileRepo.Create(ctx, authProfile))

	// Create MFA factor for owner.
	mfaFactorRepo := repoFactory.MFAFactorRepository()
	factor := &cryptoutilIdentityDomain.MFAFactor{
		ID:            googleUuid.Must(googleUuid.NewV7()),
		Name:          "Owner TOTP",
		FactorType:    cryptoutilIdentityDomain.MFAFactorTypeTOTP,
		AuthProfileID: authProfile.ID,
		Required:      cryptoutilIdentityDomain.IntBool(false),
		Enabled:       true,
		Order:         1,
	}
	require.NoError(t, mfaFactorRepo.Create(ctx, factor))

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	app := fiber.New()
	svc.RegisterRoutes(app)

	// Attempt to delete factor using other user's ID.
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/oidc/v1/mfa/factors/%s?user_id=%s", factor.ID.String(), otherUser.ID.String()), nil)

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "unauthorized", errResp["error"])
	require.Contains(t, errResp["error_description"], "does not belong to specified user")
}
