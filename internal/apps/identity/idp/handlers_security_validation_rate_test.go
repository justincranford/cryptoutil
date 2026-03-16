// Copyright (c) 2025 Justin Cranford
//
//

//nolint:goconst // Linter suggests extracting testPassword to magic constant - intentionally inline for test clarity
package idp_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

	cryptoutilIdentityClientAuth "cryptoutil/internal/apps/identity/authz/clientauth"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
)

func TestSecurityValidation_RateLimiting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create in-memory database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
	}

	// Create repository factory.
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err, "Failed to create repository factory")

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

	// Create default auth profile.
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
			AccessTokenLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
		},
		Sessions: &cryptoutilIdentityConfig.SessionConfig{
			CookieName:      "identity_session",
			SessionLifetime: cryptoutilSharedMagic.IMDefaultSessionTimeout * time.Second,
			CookieHTTPOnly:  true,
		},
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			TLSEnabled: true,
		},
	}

	// Create token service.
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, config.Tokens)

	// Create IDP service.
	service := cryptoutilIdentityIdp.NewService(config, repoFactory, tokenSvc)

	// Start service to initialize auth profiles.
	err = service.Start(ctx)
	require.NoError(t, err, "Failed to start IDP service")

	// Create Fiber app and register IDP routes.
	app := fiber.New()
	service.RegisterRoutes(app)

	// Create test client.
	testClientSecret := "test-client-secret-" + googleUuid.Must(googleUuid.NewV7()).String() // pragma: allowlist secret
	testClientSecretHash, err := cryptoutilIdentityClientAuth.HashLowEntropyNonDeterministic(testClientSecret)
	require.NoError(t, err, "Failed to hash client secret")

	testClientID := googleUuid.Must(googleUuid.NewV7()).String()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                testClientID,
		ClientSecret:            testClientSecretHash,
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		RedirectURIs:            []string{cryptoutilSharedMagic.DemoRedirectURI},
		AllowedScopes:           []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile, cryptoutilSharedMagic.ClaimEmail},
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilSharedMagic.ResponseTypeCode},
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretPost,
		RequirePKCE:             boolPtr(true),
		PKCEChallengeMethod:     cryptoutilSharedMagic.PKCEMethodS256,
		Enabled:                 boolPtr(true),
		Name:                    "Test Client",
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}

	clientRepo := repoFactory.ClientRepository()
	require.NoError(t, clientRepo.Create(ctx, testClient), "Failed to create test client")

	// Create test user.
	testUsername := "testuser-" + googleUuid.Must(googleUuid.NewV7()).String()
	testPassword := "TestPassword123!" // pragma: allowlist secret
	testPasswordHash, err := cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(testPassword)
	require.NoError(t, err, "Failed to hash test password")

	testUser := &cryptoutilIdentityDomain.User{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
		PreferredUsername: testUsername,
		Email:             fmt.Sprintf("%s@example.com", testUsername),
		PasswordHash:      testPasswordHash,
		Enabled:           true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	userRepo := repoFactory.UserRepository()
	require.NoError(t, userRepo.Create(ctx, testUser), "Failed to create test user")

	// Create authorization request.
	authzReq := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:           googleUuid.Must(googleUuid.NewV7()),
		ClientID:     testClient.ClientID,
		RedirectURI:  testClient.RedirectURIs[0],
		ResponseType: cryptoutilSharedMagic.ResponseTypeCode,
		Scope:        "openid profile email",
		State:        "test-state",
		Nonce:        "test-nonce",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
	}

	authzReqRepo := repoFactory.AuthorizationRequestRepository()
	require.NoError(t, authzReqRepo.Create(ctx, authzReq), "Failed to create authorization request")

	// Note: Rate limiting implementation is deferred (MEDIUM priority TODO).
	// This test documents expected behavior for when rate limiting is implemented.
	//
	// Expected implementation:
	// 1. Track failed login attempts per username/IP in cache or database
	// 2. Implement exponential backoff or fixed threshold (e.g., 5 attempts per 15 minutes)
	// 3. Return HTTP 429 Too Many Requests after threshold exceeded
	// 4. Include Retry-After header indicating lockout duration
	//
	// For now, verify that multiple failed attempts are handled correctly (without rate limiting).

	const maxAttempts = 10

	failedAttempts := 0

	for i := 0; i < maxAttempts; i++ {
		formData := url.Values{}
		formData.Set("username", testUsername)
		formData.Set("password", "WrongPassword123!") // Incorrect password
		formData.Set("request_id", authzReq.ID.String())

		req := httptest.NewRequest(
			http.MethodPost,
			"/oidc/v1/login",
			strings.NewReader(formData.Encode()),
		)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationForm)

		resp, err := app.Test(req, -1)
		require.NoError(t, err, "Login request failed")

		defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup

		if resp.StatusCode == http.StatusUnauthorized {
			failedAttempts++
		}
	}

	// Without rate limiting, all attempts should fail with 401 Unauthorized.
	require.Equal(t, maxAttempts, failedAttempts, "All failed attempts should return 401 without rate limiting")
	// TODO: When rate limiting is implemented, update this test to expect:
	// - HTTP 429 Too Many Requests after threshold exceeded
	// - Retry-After header present
	// - Subsequent attempts blocked until lockout expires
}
