// Copyright (c) 2025 Justin Cranford

package authz_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

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

	resp, err := app.Test(req, -1)
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, errResp[cryptoutilSharedMagic.StringError])
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

	resp, err := app.Test(req, -1)
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "factor_not_found", errResp[cryptoutilSharedMagic.StringError])
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

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	var errResp map[string]any

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "unauthorized", errResp[cryptoutilSharedMagic.StringError])
	require.Contains(t, errResp["error_description"], "does not belong to specified user")
}
