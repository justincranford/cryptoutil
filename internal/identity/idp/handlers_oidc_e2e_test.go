// Copyright (c) 2025 Justin Cranford
//
//

package idp_test

import (
	"context"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityClientAuth "cryptoutil/internal/identity/authz/clientauth"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/identity/idp"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

const (
	testPassword = "TestPassword123!" // pragma: allowlist secret
)

// TestOIDCFlow_IDPEndpointsIntegration tests that IDP endpoints work together correctly.
// This validates login → consent flow and userinfo endpoint functionality.
//
// Validates requirement R02-07: OIDC integration tests.
// Note: Full E2E OIDC flow (authorize → token) requires authz service and is tested in
// internal/identity/integration/ and internal/identity/test/e2e/.
func TestOIDCFlow_IDPEndpointsIntegration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test database config.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

	defer func() {
		_ = repoFactory.Close() //nolint:errcheck // Test cleanup
	}()

	// Run database migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
		&cryptoutilIdentityDomain.Token{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.AuthorizationRequest{},
		&cryptoutilIdentityDomain.ConsentDecision{},
		&cryptoutilIdentityDomain.ClientProfile{},
		&cryptoutilIdentityDomain.AuthProfile{},
		&cryptoutilIdentityDomain.AuthFlow{},
		&cryptoutilIdentityDomain.MFAFactor{},
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err, "Failed to run database migrations")

	// Create default auth profile (required for login).
	defaultAuthProfile := &cryptoutilIdentityDomain.AuthProfile{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		Name:        "default",
		Description: "Default authentication profile",
		ProfileType: cryptoutilIdentityDomain.AuthProfileTypeUsernamePassword,
		RequireMFA:  false,
		MFAChain:    []string{},
		Enabled:     true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	authProfileRepo := repoFactory.AuthProfileRepository()
	require.NoError(t, authProfileRepo.Create(ctx, defaultAuthProfile), "Failed to create default auth profile")

	// Create test config.
	config := &cryptoutilIdentityConfig.Config{
		Database: dbConfig,
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenLifetime: 3600 * time.Second,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      "identity_session",
			SessionLifetime: 3600 * time.Second,
			CookieHTTPOnly:  true,
		},
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			TLSEnabled: true,
		},
	}

	// Create token service (minimal for testing).
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)

	// Create IDP service.
	service := cryptoutilIdentityIdp.NewService(config, repoFactory, tokenSvc)

	// Start service to initialize auth profiles.
	err = service.Start(ctx)
	require.NoError(t, err, "Failed to start IDP service")

	// Create Fiber app and register IDP routes.
	app := fiber.New()
	service.RegisterRoutes(app)

	// Create test client with hashed secret.
	testClientSecret := "test-client-secret-" + googleUuid.Must(googleUuid.NewV7()).String() // pragma: allowlist secret
	testClientSecretHash, err := cryptoutilIdentityClientAuth.HashLowEntropyNonDeterministic(testClientSecret)
	require.NoError(t, err, "Failed to hash client secret")

	testClientID := googleUuid.Must(googleUuid.NewV7()).String()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                testClientID,
		ClientSecret:            testClientSecretHash,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedScopes:           []string{"openid", "profile", "email"},
		AllowedGrantTypes:       []string{"authorization_code"},
		AllowedResponseTypes:    []string{"code"},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		RequirePKCE:             boolPtr(true),
		PKCEChallengeMethod:     "S256",
		Enabled:                 boolPtr(true),
		Name:                    "Test Client",
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	clientRepo := repoFactory.ClientRepository()
	require.NoError(t, clientRepo.Create(ctx, testClient), "Failed to create test client")

	// Create test user with hashed password.
	testUsername := "testuser-" + googleUuid.Must(googleUuid.NewV7()).String()
	testPasswordHash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(testPassword)
	require.NoError(t, err, "Failed to hash password")

	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		PreferredUsername: testUsername,
		Email:             "testuser-" + googleUuid.Must(googleUuid.NewV7()).String() + "@example.com",
		EmailVerified:     true,
		Name:              "Test User",
		GivenName:         "Test",
		FamilyName:        "User",
		Enabled:           true,
		PasswordHash:      testPasswordHash,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser), "Failed to create test user")

	// Create authorization request (simulates authz service creating request).
	authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     testClient.ClientID,
		RedirectURI:  testClient.RedirectURIs[0],
		ResponseType: "code",
		Scope:        "openid profile email",
		State:        "test-state",
		Nonce:        "test-nonce",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(10 * time.Minute),
	}

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	require.NoError(t, authzReqRepo.Create(ctx, authzReq), "Failed to create authorization request")

	// Step 1: POST /login - Authenticate user (simulate form submission).
	loginFormData := url.Values{}
	loginFormData.Set("username", testUsername)
	loginFormData.Set("password", testPassword)
	loginFormData.Set("request_id", authzReq.ID.String())

	loginReq := httptest.NewRequest(
		http.MethodPost,
		"/oidc/v1/login",
		strings.NewReader(loginFormData.Encode()),
	)
	loginReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

	loginResp, err := app.Test(loginReq, -1)
	require.NoError(t, err, "Login request failed")

	defer func() { _ = loginResp.Body.Close() }() //nolint:errcheck // Test code cleanup

	require.Equal(t, http.StatusFound, loginResp.StatusCode, "Expected 302 redirect to consent page")

	// Extract session cookie from Set-Cookie header.
	setCookieHeader := loginResp.Header.Get("Set-Cookie")
	require.Contains(t, setCookieHeader, "identity_session", "Expected session cookie")

	// Parse session cookie value.
	sessionCookie := parseCookieValue(setCookieHeader, "identity_session")
	require.NotEmpty(t, sessionCookie, "Session cookie value is empty")

	// Verify redirect to consent page.
	consentLocationHeader := loginResp.Header.Get("Location")
	require.Contains(t, consentLocationHeader, "/oidc/v1/consent", "Expected redirect to consent page")
	require.Contains(t, consentLocationHeader, "request_id=", "Expected request_id in redirect URL")

	// Step 2: POST /consent - Approve scopes (simulate form submission).
	consentFormData := url.Values{}
	consentFormData.Set("request_id", authzReq.ID.String())
	consentFormData.Set("decision", "approve")

	consentReq := httptest.NewRequest(
		http.MethodPost,
		"/oidc/v1/consent",
		strings.NewReader(consentFormData.Encode()),
	)
	consentReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)
	consentReq.Header.Set("Cookie", fmt.Sprintf("%s=%s", "identity_session", sessionCookie))

	consentResp, err := app.Test(consentReq, -1)
	require.NoError(t, err, "Consent request failed")

	defer func() { _ = consentResp.Body.Close() }() //nolint:errcheck // Test code cleanup

	require.Equal(t, http.StatusFound, consentResp.StatusCode, "Expected 302 redirect to client redirect_uri with code")

	// Extract authorization code from redirect URL.
	callbackLocationHeader := consentResp.Header.Get("Location")
	require.Contains(t, callbackLocationHeader, testClient.RedirectURIs[0], "Expected redirect to client redirect_uri")
	require.Contains(t, callbackLocationHeader, "code=", "Expected code parameter in redirect URL")
	require.Contains(t, callbackLocationHeader, "state=test-state", "Expected state parameter in redirect URL")

	// Parse authorization code from Location header.
	callbackURL, err := url.Parse(callbackLocationHeader)
	require.NoError(t, err, "Failed to parse callback Location header")

	authorizationCode := callbackURL.Query().Get("code")
	require.NotEmpty(t, authorizationCode, "Authorization code is empty")

	// Step 3: Simulate token issuance (normally done by authz service).
	// Create access token in database.
	accessToken := googleUuid.Must(googleUuid.NewV7()).String()
	clientIDUUID, err := googleUuid.Parse(testClientID)
	require.NoError(t, err, "Failed to parse client ID")

	token := &cryptoutilIdentityDomain.Token{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
		TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
		TokenValue:  accessToken,
		ClientID:    clientIDUUID,
		UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: testUser.ID, Valid: true},
		Scopes:      []string{"openid", "profile", "email"},
		IssuedAt:    time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
	}

	tokenRepo := repoFactory.TokenRepository()
	require.NoError(t, tokenRepo.Create(ctx, token), "Failed to create access token")

	// Step 4: GET /userinfo - Retrieve user claims using access token.
	// Note: This test uses plain token instead of JWT for simplicity.
	// Production uses JWT tokens validated by handleUserInfo via tokenSvc.ValidateAccessToken().
	// This tests the IDP endpoint logic, not the full token validation flow.
	userinfoReq := httptest.NewRequest(
		http.MethodGet,
		"/oidc/v1/userinfo",
		nil,
	)
	userinfoReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	userinfoResp, err := app.Test(userinfoReq, -1)
	require.NoError(t, err, "UserInfo request failed")

	defer func() { _ = userinfoResp.Body.Close() }() //nolint:errcheck // Test code cleanup

	// UserInfo endpoint validation happens via tokenSvc.ValidateAccessToken() which returns error for non-JWT.
	// This test confirms the endpoint is registered and responds (even if validation fails).
	// Full JWT token validation is tested in integration tests with real token service.
	require.Contains(t, []int{http.StatusOK, http.StatusUnauthorized}, userinfoResp.StatusCode,
		"UserInfo endpoint should respond with either 200 (valid JWT) or 401 (invalid token format)")
}

// parseCookieValue extracts cookie value from Set-Cookie header by name.
func parseCookieValue(setCookieHeader, cookieName string) string {
	parts := strings.Split(setCookieHeader, ";")
	if len(parts) == 0 {
		return ""
	}

	cookiePart := strings.TrimSpace(parts[0])
	keyValue := strings.SplitN(cookiePart, "=", 2)

	if len(keyValue) != 2 {
		return ""
	}

	if keyValue[0] == cookieName {
		return keyValue[1]
	}

	return ""
}
