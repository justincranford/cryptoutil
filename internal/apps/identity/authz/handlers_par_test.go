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
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.DemoRedirectURI},
		cryptoutilSharedMagic.ParamScope:               []string{"openid profile"},
		cryptoutilSharedMagic.ParamState:               []string{"random-state-value"},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-code-challenge-value-xxxxxxxxxxxxxxxxx"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
		cryptoutilSharedMagic.ClaimNonce: []string{"random-nonce-value"},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusCreated, resp.StatusCode, "Should return 201 Created")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	// Validate response fields (RFC 9126 Section 2.1).
	require.Contains(t, result, cryptoutilSharedMagic.ParamRequestURI, "Response should include request_uri")
	require.Contains(t, result, cryptoutilSharedMagic.ParamExpiresIn, "Response should include expires_in")

	// Validate field types and values.
	requestURI, ok := result[cryptoutilSharedMagic.ParamRequestURI].(string)
	require.True(t, ok, "request_uri should be string")
	require.True(t, strings.HasPrefix(requestURI, cryptoutilSharedMagic.RequestURIPrefix), "request_uri should start with URN prefix")
	require.GreaterOrEqual(t, len(requestURI), len(cryptoutilSharedMagic.RequestURIPrefix)+cryptoutilSharedMagic.DefaultCodeChallengeLength, "request_uri should be at least 43 chars")

	expiresIn, ok := result[cryptoutilSharedMagic.ParamExpiresIn].(float64)
	require.True(t, ok, "expires_in should be number")
	require.Equal(t, float64(cryptoutilSharedMagic.StrictCertificateMaxAgeDays), expiresIn, "expires_in should be 90 seconds")

	// Verify PAR stored in database.
	parRepo := repoFactory.PushedAuthorizationRequestRepository()
	storedPAR, err := parRepo.GetByRequestURI(ctx, requestURI)
	require.NoError(t, err, "Should retrieve stored PAR from database")
	require.NotNil(t, storedPAR, "Stored PAR should not be nil")
	require.Equal(t, testClient.ID, storedPAR.ClientID, "ClientID should match")
	require.Equal(t, cryptoutilSharedMagic.ResponseTypeCode, storedPAR.ResponseType, "ResponseType should match")
	require.Equal(t, cryptoutilSharedMagic.DemoRedirectURI, storedPAR.RedirectURI, "RedirectURI should match")
	require.Equal(t, "openid profile", storedPAR.Scope, "Scope should match")
	require.Equal(t, "random-state-value", storedPAR.State, "State should match")
	require.Equal(t, "test-code-challenge-value-xxxxxxxxxxxxxxxxx", storedPAR.CodeChallenge, "CodeChallenge should match")
	require.Equal(t, cryptoutilSharedMagic.PKCEMethodS256, storedPAR.CodeChallengeMethod, "CodeChallengeMethod should match")
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
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.DemoRedirectURI},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result[cryptoutilSharedMagic.StringError].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, cryptoutilSharedMagic.ClaimClientID, "Error description should mention client_id")
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
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.DemoRedirectURI},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result[cryptoutilSharedMagic.StringError].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, cryptoutilSharedMagic.ParamResponseType, "Error description should mention response_type")
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
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamCodeChallenge:       []string{"test-code-challenge"},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result[cryptoutilSharedMagic.StringError].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, cryptoutilSharedMagic.ParamRedirectURI, "Error description should mention redirect_uri")
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
		cryptoutilSharedMagic.ParamClientID:            []string{testClient.ClientID},
		cryptoutilSharedMagic.ParamResponseType:        []string{cryptoutilSharedMagic.ResponseTypeCode},
		cryptoutilSharedMagic.ParamRedirectURI:         []string{cryptoutilSharedMagic.DemoRedirectURI},
		cryptoutilSharedMagic.ParamCodeChallengeMethod: []string{cryptoutilSharedMagic.PKCEMethodS256},
	}

	req := httptest.NewRequest("POST", "/oauth2/v1/par", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req, -1)
	require.NoError(t, err, "Request should succeed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Should return 400 Bad Request")

	var result map[string]any

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Should decode JSON response")

	errorCode, ok := result[cryptoutilSharedMagic.StringError].(string)
	require.True(t, ok, "error should be string")
	require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, errorCode, "Should return invalid_request error")

	errorDescription, ok := result["error_description"].(string)
	require.True(t, ok, "error_description should be string")
	require.Contains(t, errorDescription, cryptoutilSharedMagic.ParamCodeChallenge, "Error description should mention code_challenge")
}

// TestHandlePAR_InvalidClient validates error when client_id is invalid.
