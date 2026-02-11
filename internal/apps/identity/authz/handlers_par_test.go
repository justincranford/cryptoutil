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
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestHandlePAR_HappyPath validates successful PAR request.
func TestHandlePAR_HappyPath(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamScope:               []string{"openid profile"},
		cryptoutilIdentityMagic.ParamState:               []string{"random-state-value"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge-value-xxxxxxxxxxxxxxxxx"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
		"nonce": []string{"random-nonce-value"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode, "Should return 201 Created")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	// Validate response fields (RFC 9126 Section 2.1).
	require.Contains(t, result, "request_uri", "Response should include request_uri")
	require.Contains(t, result, "expires_in", "Response should include expires_in")

	// Validate field types and values.
	requestURI, ok := result["request_uri"].(string)
	require.True(t, ok, "request_uri should be string")
	require.True(t, strings.HasPrefix(requestURI, cryptoutilIdentityMagic.RequestURIPrefix), "request_uri should start with URN prefix")
	require.GreaterOrEqual(t, len(requestURI), len(cryptoutilIdentityMagic.RequestURIPrefix)+43, "request_uri should be at least 43 chars")

	expiresIn, ok := result["expires_in"].(float64)
	require.True(t, ok, "expires_in should be number")
	require.Equal(t, float64(90), expiresIn, "expires_in should be 90 seconds")

	// Verify PAR stored in database.
	parRepo := repoFactory.PushedAuthorizationRequestRepository()
	storedPAR, err := parRepo.GetByRequestURI(ctx, requestURI)
	require.NoError(t, err, "Should retrieve stored PAR from database")
	require.NotNil(t, storedPAR, "Stored PAR should not be nil")
	require.Equal(t, testClient.ID, storedPAR.ClientID, "ClientID should match")
	require.Equal(t, cryptoutilIdentityMagic.ResponseTypeCode, storedPAR.ResponseType, "ResponseType should match")
	require.Equal(t, "https://example.com/callback", storedPAR.RedirectURI, "RedirectURI should match")
	require.Equal(t, "openid profile", storedPAR.Scope, "Scope should match")
	require.Equal(t, "random-state-value", storedPAR.State, "State should match")
	require.Equal(t, "test-code-challenge-value-xxxxxxxxxxxxxxxxx", storedPAR.CodeChallenge, "CodeChallenge should match")
	require.Equal(t, cryptoutilIdentityMagic.PKCEMethodS256, storedPAR.CodeChallengeMethod, "CodeChallengeMethod should match")
	require.Equal(t, "random-nonce-value", storedPAR.Nonce, "Nonce should match")
	require.False(t, storedPAR.Used, "PAR should not be marked as used initially")
}

// TestHandlePAR_MissingClientID validates error when client_id is missing.
func TestHandlePAR_MissingClientID(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "client_id", "Error description should mention client_id")
}

// TestHandlePAR_MissingResponseType validates error when response_type is missing.
func TestHandlePAR_MissingResponseType(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "response_type", "Error description should mention response_type")
}

// TestHandlePAR_MissingRedirectURI validates error when redirect_uri is missing.
func TestHandlePAR_MissingRedirectURI(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "redirect_uri", "Error description should mention redirect_uri")
}

// TestHandlePAR_MissingCodeChallenge validates error when code_challenge is missing.
func TestHandlePAR_MissingCodeChallenge(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "code_challenge", "Error description should mention code_challenge")
}

// TestHandlePAR_InvalidClient validates error when client_id is invalid.
func TestHandlePAR_InvalidClient(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{"nonexistent-client-id"},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidClient, errorCode, "Should return invalid_client error")
}

// TestHandlePAR_InvalidRedirectURI validates error when redirect_uri is not registered.
func TestHandlePAR_InvalidRedirectURI(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://malicious.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "redirect_uri", "Error description should mention redirect_uri")
}

// TestHandlePAR_UnsupportedResponseType validates error for unsupported response_type.
func TestHandlePAR_UnsupportedResponseType(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{"token"}, // Only "code" is supported.
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{cryptoutilIdentityMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorUnsupportedResponseType, errorCode, "Should return unsupported_response_type error")
}

// TestHandlePAR_UnsupportedCodeChallengeMethod validates error for unsupported code_challenge_method.
func TestHandlePAR_UnsupportedCodeChallengeMethod(t *testing.T) {
	t.Parallel()

	config, repoFactory := createPARTestDependencies(t)

	ctx := context.Background()
	testClient := createTestClientForPAR(ctx, t, repoFactory)

	svc := cryptoutilIdentityAuthz.NewService(config, repoFactory, nil)
	require.NotNil(t, svc, "Service should not be nil")

	app := fiber.New()
	svc.RegisterRoutes(app)

	formData := url.Values{
		cryptoutilIdentityMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilIdentityMagic.ParamResponseType:        []string{cryptoutilIdentityMagic.ResponseTypeCode},
		cryptoutilIdentityMagic.ParamRedirectURI:         []string{"https://example.com/callback"},
		cryptoutilIdentityMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilIdentityMagic.ParamCodeChallengeMethod: []string{"plain"}, // Only S256 is supported.
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result["error"].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilIdentityMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, "code_challenge_method", "Error description should mention code_challenge_method")
}

// createPARTestDependencies creates test dependencies for PAR tests.
func createPARTestDependencies(t *testing.T) (*cryptoutilIdentityConfig.Config, *cryptoutilIdentityRepository.RepositoryFactory) {
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

	// Auto-migrate all required tables for PAR tests.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.PushedAuthorizationRequest{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.ClientSecretHistory{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, err, "Failed to auto-migrate database tables")

	return config, repoFactory
}

// createTestClientForPAR creates a test client for PAR tests.
func createTestClientForPAR(ctx context.Context, t *testing.T, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) *cryptoutilIdentityDomain.Client {
	t.Helper()

	enabled := true
	requirePKCE := true

	client := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.Must(googleUuid.NewV7()),
		ClientID:                "test-par-client-" + googleUuid.NewString()[:8],
		ClientSecret:            "$2a$10$examplehashedvalue",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		Name:                    "Test PAR Client",
		RedirectURIs:            []string{"https://example.com/callback"},
		AllowedScopes:           []string{"openid", "profile", "email"},
		AllowedGrantTypes:       []string{cryptoutilIdentityMagic.GrantTypeAuthorizationCode},
		AllowedResponseTypes:    []string{cryptoutilIdentityMagic.ResponseTypeCode},
		TokenEndpointAuthMethod: cryptoutilIdentityMagic.ClientAuthMethodSecretPost,
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
