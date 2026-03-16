// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestHandleDeviceAuthorization_HappyPath validates successful device authorization request.
func TestHandleDeviceAuthorization_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createDeviceAuthTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForDevice(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID: []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamScope:    []string{"openid profile"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/device_authorization", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	// Validate response fields (RFC 8628 Section 3.2).
	require.Contains(t, result, cryptoutilSharedMagic.ParamDeviceCode, "Response should include device_code")
	require.Contains(t, result, cryptoutilSharedMagic.ParamUserCode, "Response should include user_code")
	require.Contains(t, result, "verification_uri", "Response should include verification_uri")
	require.Contains(t, result, "verification_uri_complete", "Response should include verification_uri_complete")
	require.Contains(t, result, cryptoutilSharedMagic.ParamExpiresIn, "Response should include expires_in")
	require.Contains(t, result, "interval", "Response should include interval")

	// Validate field types and values.
	deviceCode, ok := result[cryptoutilSharedMagic.ParamDeviceCode].(string)
	require.True(t, ok, "device_code should be string")
	require.NotEmpty(t, deviceCode, "device_code should not be empty")
	require.GreaterOrEqual(t, len(deviceCode), 40, "device_code should be at least 40 characters")

	userCode, ok := result[cryptoutilSharedMagic.ParamUserCode].(string)
	require.True(t, ok, "user_code should be string")
	require.NotEmpty(t, userCode, "user_code should not be empty")
	require.Len(t, userCode, 9, "user_code should be 9 characters (XXXX-YYYY)")
	require.Equal(t, "-", string(userCode[4]), "user_code should have hyphen at position 4")

	verificationURI, ok := result["verification_uri"].(string)
	require.True(t, ok, "verification_uri should be string")
	require.Contains(t, verificationURI, "/device", "verification_uri should contain /device path")

	verificationURIComplete, ok := result["verification_uri_complete"].(string)
	require.True(t, ok, "verification_uri_complete should be string")
	require.Contains(t, verificationURIComplete, userCode, "verification_uri_complete should include user_code")
	require.Contains(t, verificationURIComplete, "user_code=", "verification_uri_complete should include user_code parameter")

	expiresIn, ok := result[cryptoutilSharedMagic.ParamExpiresIn].(float64)
	require.True(t, ok, "expires_in should be number")
	require.Equal(t, float64(cryptoutilSharedMagic.IMEnterpriseSessionTimeout), expiresIn, "expires_in should be 1800 seconds (30 minutes)")

	interval, ok := result["interval"].(float64)
	require.True(t, ok, "interval should be number")
	require.Equal(t, float64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), interval, "interval should be 5 seconds")

	// Verify device authorization stored in database.
	deviceAuthRepo := repoFactory.DeviceAuthorizationRepository()
	storedAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	require.NoError(t, err, "Should retrieve device authorization from database")
	require.Equal(t, testClient.ClientID, storedAuth.ClientID, "Client ID should match")
	require.Equal(t, "openid profile", storedAuth.Scope, "Scope should match")
	require.Equal(t, cryptoutilIdentityDomain.DeviceAuthStatusPending, storedAuth.Status, "Status should be pending")
	require.False(t, storedAuth.IsExpired(), "Device code should not be expired")
}

// TestHandleDeviceAuthorization_MissingClientID validates missing client_id parameter.
func TestHandleDeviceAuthorization_MissingClientID(t *testing.T) {
	t.Parallel()

	config, repoFactory := createDeviceAuthTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamScope: []string{cryptoutilSharedMagic.ScopeOpenID},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/device_authorization", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, result[cryptoutilSharedMagic.StringError], "Error code should be invalid_request")
	require.Contains(t, result["error_description"], cryptoutilSharedMagic.ClaimClientID, "Error description should mention client_id")
}

// TestHandleDeviceAuthorization_InvalidClientID validates invalid client_id.
func TestHandleDeviceAuthorization_InvalidClientID(t *testing.T) {
	t.Parallel()

	config, repoFactory := createDeviceAuthTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID: []string{"invalid-client-id-12345"},
		cryptoutilSharedMagic.ParamScope:    []string{cryptoutilSharedMagic.ScopeOpenID},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/device_authorization", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidClient, result[cryptoutilSharedMagic.StringError], "Error code should be invalid_client")
}

// TestHandleDeviceAuthorization_OptionalScope validates request without scope parameter.
func TestHandleDeviceAuthorization_OptionalScope(t *testing.T) {
	t.Parallel()

	config, repoFactory := createDeviceAuthTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForDevice(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID: []string{testClient.ClientID},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/device_authorization", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusOK, resp.StatusCode, "Should return 200 OK even without scope")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	require.Contains(t, result, cryptoutilSharedMagic.ParamDeviceCode, "Response should include device_code")
	require.Contains(t, result, cryptoutilSharedMagic.ParamUserCode, "Response should include user_code")
}

// createDeviceAuthTestDependencies creates test dependencies for device authorization tests.
func createDeviceAuthTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	config := &cryptoutilIdentityConfig.Config{
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: cryptoutilSharedMagic.TestDatabaseSQLite,
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

	// Auto-migrate all required tables for device authorization tests.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.DeviceAuthorization{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.ClientSecretHistory{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, err, "Failed to auto-migrate database tables")

	return config, repoFactory
}

// createTestClientForDevice creates a test client for device authorization tests.
func createTestClientForDevice(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	enabled := true
	requirePKCE := false

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-device-client-" + googleUuid.NewString()[:cryptoutilSharedMagic.IMMinPasswordLength],
		ClientSecret:            "$2a$10$examplehashedvalue",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test Device Client",
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode, cryptoutilSharedMagic.GrantTypeDeviceCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		TokenEndpointAuthMethod: cryptoutilSharedMagic.ClientAuthMethodSecretPost,
		RequirePKCE:             &requirePKCE,
		AccessTokenLifetime:     cryptoutilSharedMagic.IMDefaultSessionTimeout,
		RefreshTokenLifetime:    cryptoutilSharedMagic.IMDefaultSessionAbsoluteMax,
		IDTokenLifetime:         cryptoutilSharedMagic.IMDefaultSessionTimeout,
		Enabled:                 &enabled,
	}

	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err, "Failed to create test client")

	return client
}
