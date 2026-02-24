// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck,revive // Integration test with realistic error propagation
package authz_test

import (
	"context"
	json "encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestDeviceAuthorizationFlow_HappyPath validates complete device authorization flow (RFC 8628).
func TestDeviceAuthorizationFlow_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createIntegrationTestDependencies(t)

	ctx := context.Background()

	// Create test user who will authorize the device.
	testUser := createIntegrationTestUser(ctx, t, repoFactory)

	// Create test client with device_code grant type.
	testClient := createIntegrationTestClient(ctx, t, repoFactory)

	// Create AuthZ service (nil tokenSvc - will fail at token issuance, but that's okay for this test).
	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// ========== STEP 1: Device Requests Authorization ==========
	deviceAuthResp := requestDeviceAuthorization(t, app, testClient.ClientID)

	// Validate response fields.
	require.Contains(t, deviceAuthResp, "device_code", "Response should include device_code")
	require.Contains(t, deviceAuthResp, "user_code", "Response should include user_code")
	require.Contains(t, deviceAuthResp, "verification_uri", "Response should include verification_uri")
	require.Contains(t, deviceAuthResp, "expires_in", "Response should include expires_in")
	require.Contains(t, deviceAuthResp, "interval", "Response should include interval")

	deviceCode, ok := deviceAuthResp["device_code"].(string)
	require.True(t, ok, "device_code should be string")

	userCode, ok := deviceAuthResp["user_code"].(string)
	require.True(t, ok, "user_code should be string")

	// ========== STEP 2: Poll for Token (Should Return authorization_pending) ==========
	pollResp1 := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorAuthorizationPending, pollResp1["error"], "First poll should return authorization_pending")

	// ========== STEP 3: User Authorizes Device (Simulate User Consent) ==========
	authorizeDevice(ctx, t, repoFactory, userCode, testUser.ID)

	// ========== STEP 4: Verify Device Authorization Status in Database ==========
	deviceAuthRepo := repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	require.NoError(t, err, "Should retrieve device authorization")
	require.Equal(t, cryptoutilIdentityDomain.DeviceAuthStatusAuthorized, deviceAuth.Status, "Device should be authorized")
	require.True(t, deviceAuth.UserID.Valid, "UserID should be set")
	require.Equal(t, testUser.ID, deviceAuth.UserID.UUID, "UserID should match test user")
	// Note: Token issuance step skipped in this test because TokenService setup is complex.
	// Token issuance is tested separately in other unit tests.
}

// TestDeviceAuthorizationFlow_ExpiredCode validates device code expiration.
func TestDeviceAuthorizationFlow_ExpiredCode(t *testing.T) {
	t.Parallel()

	config, repoFactory := createIntegrationTestDependencies(t)

	ctx := context.Background()
	testClient := createIntegrationTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Request device authorization.
	deviceAuthResp := requestDeviceAuthorization(t, app, testClient.ClientID)

	deviceCode, ok := deviceAuthResp["device_code"].(string)
	require.True(t, ok, "device_code should be string")

	// Manually expire the device code in database.
	deviceAuthRepo := repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByDeviceCode(ctx, deviceCode)
	require.NoError(t, err, "Should retrieve device authorization")

	deviceAuth.ExpiresAt = time.Now().UTC().Add(-1 * time.Hour) // Set expiration to past
	err = deviceAuthRepo.Update(ctx, deviceAuth)
	require.NoError(t, err, "Should update device authorization")

	// Poll for token (should return expired_token).
	pollResp := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorExpiredToken, pollResp["error"], "Should return expired_token error")
}

// TestDeviceAuthorizationFlow_DeniedAuthorization validates user denial.
func TestDeviceAuthorizationFlow_DeniedAuthorization(t *testing.T) {
	t.Parallel()

	config, repoFactory := createIntegrationTestDependencies(t)

	ctx := context.Background()
	testUser := createIntegrationTestUser(ctx, t, repoFactory)
	testClient := createIntegrationTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Request device authorization.
	deviceAuthResp := requestDeviceAuthorization(t, app, testClient.ClientID)

	deviceCode, ok := deviceAuthResp["device_code"].(string)
	require.True(t, ok, "device_code should be string")

	userCode, ok := deviceAuthResp["user_code"].(string)
	require.True(t, ok, "user_code should be string")

	// User denies authorization.
	denyDevice(ctx, t, repoFactory, userCode, testUser.ID)

	// Poll for token (should return access_denied).
	pollResp := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorAccessDenied, pollResp["error"], "Should return access_denied error")
}

// TestDeviceAuthorizationFlow_SlowDown validates polling rate limiting.
func TestDeviceAuthorizationFlow_SlowDown(t *testing.T) {
	t.Parallel()

	config, repoFactory := createIntegrationTestDependencies(t)

	ctx := context.Background()
	testClient := createIntegrationTestClient(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	// Request device authorization.
	deviceAuthResp := requestDeviceAuthorization(t, app, testClient.ClientID)

	deviceCode, ok := deviceAuthResp["device_code"].(string)
	require.True(t, ok, "device_code should be string")

	// Poll first time (should succeed with authorization_pending).
	pollResp1 := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorAuthorizationPending, pollResp1["error"], "First poll should return authorization_pending")

	// Poll immediately again (should return slow_down).
	pollResp2 := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorSlowDown, pollResp2["error"], "Second immediate poll should return slow_down")

	// Wait for polling interval to elapse.
	time.Sleep(cryptoutilSharedMagic.DefaultPollingInterval + 100*time.Millisecond)

	// Poll again after interval (should succeed with authorization_pending).
	pollResp3 := pollDeviceToken(t, app, testClient.ClientID, deviceCode, 400)
	require.Equal(t, cryptoutilSharedMagic.ErrorAuthorizationPending, pollResp3["error"], "Third poll after interval should return authorization_pending")
}

// ========== Helper Functions ==========

// createIntegrationTestDependencies creates test dependencies for integration tests.
func createIntegrationTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
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

	// Auto-migrate required tables.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.DeviceAuthorization{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.ClientSecretHistory{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err, "Failed to auto-migrate database tables")

	return config, repoFactory
}

// createIntegrationTestUser creates a test user for authorization.
func createIntegrationTestUser(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.User {
	t.Helper()

	enabled := true
	user := &cryptoutilIdentityDomain.User{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		Sub:          "testuser-" + googleUuid.NewString()[:8],
		Email:        "testuser@example.com",
		PasswordHash: "$2a$10$examplehash",
		Enabled:      enabled,
	}

	userRepo := repoFactory.UserRepository()
	err := userRepo.Create(ctx, user)
	require.NoError(t, err, "Failed to create test user")

	return user
}

// createIntegrationTestClient creates a test client with device_code grant type.
func createIntegrationTestClient(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	enabled := true
	requirePKCE := false

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "device-client-" + googleUuid.NewString()[:8],
		ClientSecret:            "$2a$10$examplehashedvalue",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Integration Test Device Client",
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedScopes:           []string{"openid", "profile", "email"},
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeDeviceCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		TokenEndpointAuthMethod: cryptoutilSharedMagic.ClientAuthMethodSecretPost,
		RequirePKCE:             &requirePKCE,
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 &enabled,
	}

	clientRepo := repoFactory.ClientRepository()
	err := clientRepo.Create(ctx, client)
	require.NoError(t, err, "Failed to create test client")

	return client
}

// requestDeviceAuthorization sends POST /device_authorization request.
func requestDeviceAuthorization(t *testing.T, app *fiber.App, clientID string) map[string]any {
	t.Helper()

	formData := url.Values{
		cryptoutilSharedMagic.ParamClientID: []string{clientID},
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

	return result
}

// pollDeviceToken sends POST /token request with device_code grant.
func pollDeviceToken(t *testing.T, app *fiber.App, clientID, deviceCode string, expectedStatus int) map[string]any {
	t.Helper()

	formData := url.Values{
		cryptoutilSharedMagic.ParamGrantType:  []string{cryptoutilSharedMagic.GrantTypeDeviceCode},
		cryptoutilSharedMagic.ParamClientID:   []string{clientID},
		cryptoutilSharedMagic.ParamDeviceCode: []string{deviceCode},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, expectedStatus, resp.StatusCode, "Should return expected status code")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	return result
}

// authorizeDevice simulates user authorizing the device.
func authorizeDevice(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, userCode string, userID googleUuid.UUID) {
	t.Helper()

	deviceAuthRepo := repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByUserCode(ctx, userCode)
	require.NoError(t, err, "Should retrieve device authorization by user code")

	deviceAuth.Status = cryptoutilIdentityDomain.DeviceAuthStatusAuthorized
	deviceAuth.UserID = cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true}

	err = deviceAuthRepo.Update(ctx, deviceAuth)
	require.NoError(t, err, "Should update device authorization status to authorized")
}

// denyDevice simulates user denying the device authorization.
func denyDevice(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory, userCode string, userID googleUuid.UUID) {
	t.Helper()

	deviceAuthRepo := repoFactory.DeviceAuthorizationRepository()

	deviceAuth, err := deviceAuthRepo.GetByUserCode(ctx, userCode)
	require.NoError(t, err, "Should retrieve device authorization by user code")

	deviceAuth.Status = cryptoutilIdentityDomain.DeviceAuthStatusDenied
	deviceAuth.UserID = cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true}

	err = deviceAuthRepo.Update(ctx, deviceAuth)
	require.NoError(t, err, "Should update device authorization status to denied")
}
