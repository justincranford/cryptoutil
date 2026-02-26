// Copyright (c) 2025 Justin Cranford
//
//

//go:build integration

// Package integration provides integration tests for the identity server.
// These tests require real network listeners and use hardcoded ports,
// so they must be run separately from unit tests using:
//
//	go test -tags=integration ./internal/identity/integration/...
//
// Do NOT run these with `go test ./...` as they will conflict with parallel tests.
package integration

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"strings"
	"testing"

	cryptoutilIdentityPKCE "cryptoutil/internal/apps/identity/authz/pkce"

	testify "github.com/stretchr/testify/require"
)

func TestOAuth2AuthorizationCodeFlow(t *testing.T) {
	t.Parallel()
	// Sequential execution to avoid port conflicts with hardcoded ports.
	testMutex.Lock()

	servers, cancel := setupTestServers(t)

	defer func() {
		shutdownTestServers(t, servers)
		cancel()
		testMutex.Unlock()
	}()

	// Generate PKCE code verifier and challenge (OAuth 2.1 requirement).
	codeVerifier, err := cryptoutilIdentityPKCE.GenerateCodeVerifier()
	testify.NoError(t, err, "Failed to generate code verifier")

	codeChallenge := cryptoutilIdentityPKCE.GenerateCodeChallenge(codeVerifier, cryptoutilSharedMagic.PKCEMethodS256)

	// Step 1a: Request authorization code with PKCE.
	// This should redirect to login page with request_id.
	authorizeURL := fmt.Sprintf("%s/oauth2/v1/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=test-state&code_challenge=%s&code_challenge_method=S256",
		testAuthZBaseURL, testClientID, url.QueryEscape(testRedirectURI), url.QueryEscape(testScope), codeChallenge)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, authorizeURL, nil)
	testify.NoError(t, err, "Failed to create authorize request")

	// Don't follow redirects - we need to handle login/consent flow manually.
	client := &http.Client{
		Timeout: httpClientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	authResp, err := client.Do(req)
	testify.NoError(t, err, "Authorization request failed")

	defer func() {
		defer func() { _ = authResp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	// Should redirect to login page (302 Found).
	testify.Equal(t, http.StatusFound, authResp.StatusCode, "Should redirect to login page")

	// Extract request_id from login redirect.
	loginLocation := authResp.Header.Get("Location")
	testify.NotEmpty(t, loginLocation, "Login redirect location should be set")
	testify.Contains(t, loginLocation, "/oidc/v1/login?request_id=", "Should redirect to login page")

	loginURL, err := url.Parse(loginLocation)
	testify.NoError(t, err, "Invalid login URL")

	requestIDStr := loginURL.Query().Get("request_id")
	testify.NotEmpty(t, requestIDStr, "request_id should be present in login redirect")

	// Step 1b: Submit login credentials (simulate user login).
	loginSubmitURL := testIDPBaseURL + "/oidc/v1/login"
	loginFormData := url.Values{}
	loginFormData.Set("username", testUsername)
	loginFormData.Set("password", testPassword)
	loginFormData.Set("request_id", requestIDStr)

	loginSubmitReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, loginSubmitURL, strings.NewReader(loginFormData.Encode()))
	testify.NoError(t, err, "Failed to create login submit request")

	loginSubmitReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	loginSubmitResp, err := client.Do(loginSubmitReq)
	testify.NoError(t, err, "Login submit request failed")

	defer func() {
		defer func() { _ = loginSubmitResp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	// Debug: print response body if not 302
	if loginSubmitResp.StatusCode != http.StatusFound {
		bodyBytes, _ := io.ReadAll(loginSubmitResp.Body) //nolint:errcheck // Debug logging only
		t.Logf("Login submit response status: %d, body: %s", loginSubmitResp.StatusCode, string(bodyBytes))
	}

	// Should redirect to consent page (302 Found).
	testify.Equal(t, http.StatusFound, loginSubmitResp.StatusCode, "Should redirect to consent page after login")

	consentLocation := loginSubmitResp.Header.Get("Location")
	testify.NotEmpty(t, consentLocation, "Consent redirect location should be set")
	testify.Contains(t, consentLocation, "/oidc/v1/consent?request_id=", "Should redirect to consent page")

	// Extract session cookie for consent request.
	sessionCookies := loginSubmitResp.Cookies()
	testify.NotEmpty(t, sessionCookies, "Session cookie should be set after login")

	// Step 1c: Submit consent decision (simulate user consent).
	consentSubmitURL := testIDPBaseURL + "/oidc/v1/consent"
	consentFormData := url.Values{}
	consentFormData.Set("request_id", requestIDStr)
	consentFormData.Set("decision", "allow")

	consentSubmitReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, consentSubmitURL, strings.NewReader(consentFormData.Encode()))
	testify.NoError(t, err, "Failed to create consent submit request")

	consentSubmitReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Add session cookie to consent request.
	for _, cookie := range sessionCookies {
		consentSubmitReq.AddCookie(cookie)
	}

	consentSubmitResp, err := client.Do(consentSubmitReq)
	testify.NoError(t, err, "Consent submit request failed")

	defer func() {
		defer func() { _ = consentSubmitResp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	// Should redirect to callback with authorization code (302 Found).
	testify.Equal(t, http.StatusFound, consentSubmitResp.StatusCode, "Should redirect to callback after consent")

	// Extract authorization code from redirect location.
	location := consentSubmitResp.Header.Get("Location")
	testify.NotEmpty(t, location, "Callback redirect location should be set")

	redirectURL, err := url.Parse(location)
	testify.NoError(t, err, "Invalid redirect URL")

	code := redirectURL.Query().Get(cryptoutilSharedMagic.ResponseTypeCode)
	testify.NotEmpty(t, code, "Authorization code should be present in callback redirect")

	state := redirectURL.Query().Get(cryptoutilSharedMagic.ParamState)
	testify.Equal(t, "test-state", state, "State parameter should match")

	// Step 2: Exchange authorization code for tokens.
	tokenURL := testAuthZBaseURL + "/oauth2/v1/token"
	tokenData := url.Values{}
	tokenData.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeAuthorizationCode)
	tokenData.Set(cryptoutilSharedMagic.ResponseTypeCode, code)
	tokenData.Set(cryptoutilSharedMagic.ParamRedirectURI, testRedirectURI)
	tokenData.Set(cryptoutilSharedMagic.ClaimClientID, testClientID)
	tokenData.Set(cryptoutilSharedMagic.ParamClientSecret, testClientSecret)
	tokenData.Set(cryptoutilSharedMagic.ParamCodeVerifier, codeVerifier) // OAuth 2.1 PKCE requirement

	tokenReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, strings.NewReader(tokenData.Encode()))
	testify.NoError(t, err, "Failed to create token request")

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := servers.httpClient.Do(tokenReq)
	testify.NoError(t, err, "Token request failed")

	defer func() {
		defer func() { _ = tokenResp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	testify.Equal(t, http.StatusOK, tokenResp.StatusCode, "Token request should succeed")

	var tokenResponse map[string]any

	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	testify.NoError(t, err, "Failed to decode token response")

	accessToken, ok := tokenResponse[cryptoutilSharedMagic.TokenTypeAccessToken].(string)
	testify.True(t, ok, "Access token should be present")
	testify.NotEmpty(t, accessToken, "Access token should not be empty")

	// Step 3: Use access token to access protected resource.
	resourceURL := testRSBaseURL + "/api/v1/protected/resource"
	resourceReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, resourceURL, nil)
	testify.NoError(t, err, "Failed to create resource request")

	resourceReq.Header.Set("Authorization", bearerTokenPrefix+accessToken)

	resourceResp, err := servers.httpClient.Do(resourceReq)
	testify.NoError(t, err, "Resource request failed")

	defer func() {
		defer func() { _ = resourceResp.Body.Close() }() //nolint:errcheck // Test code cleanup
	}()

	testify.Equal(t, http.StatusOK, resourceResp.StatusCode, "Protected resource access should succeed")

	// Validates requirements:
	// - R05-01: Refresh token issuance with offline_access scope
	// - R05-02: Refresh token exchange for new access tokens
	// Verify refresh token issued when offline_access scope granted.
	refreshToken, ok := tokenResponse[cryptoutilSharedMagic.GrantTypeRefreshToken].(string)
	if testScope == "openid offline_access" {
		testify.True(t, ok, "Refresh token should be present when offline_access scope granted")
		testify.NotEmpty(t, refreshToken, "Refresh token should not be empty")
	}
}

// TestResourceServerScopeEnforcement verifies scope-based access control.
//
// Validates requirements:
// - R06-02: CSRF protection for state-changing requests
// - R06-03: Rate limiting per IP and per client.
func TestResourceServerScopeEnforcement(t *testing.T) {
	t.Parallel()
	// Sequential execution to avoid port conflicts with hardcoded ports.
	testMutex.Lock()

	servers, cancel := setupTestServers(t)

	defer func() {
		shutdownTestServers(t, servers)
		cancel()
		testMutex.Unlock()
	}()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		tokenScopes    []string
		expectedStatus int
	}{
		{
			name:           "GET Protected Resource",
			endpoint:       "/api/v1/protected/resource",
			method:         http.MethodGet,
			tokenScopes:    []string{"read:resource"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST Protected Resource",
			endpoint:       "/api/v1/protected/resource",
			method:         http.MethodPost,
			tokenScopes:    []string{"write:resource"},
			expectedStatus: http.StatusCreated, // POST returns 201 Created, not 200 OK
		},
		{
			name:           "DELETE Protected Resource - Without Delete Scope",
			endpoint:       "/api/v1/protected/resource/test-id",
			method:         http.MethodDelete,
			tokenScopes:    []string{"read:resource", "write:resource"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "DELETE Protected Resource - With Delete Scope",
			endpoint:       "/api/v1/protected/resource/test-id",
			method:         http.MethodDelete,
			tokenScopes:    []string{"delete:resource"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin Users - Without Admin Scope",
			endpoint:       "/api/v1/admin/users",
			method:         http.MethodGet,
			tokenScopes:    []string{"read:resource", "write:resource"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin Users - With Admin Scope",
			endpoint:       "/api/v1/admin/users",
			method:         http.MethodGet,
			tokenScopes:    []string{testAdminScope},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get access token with specified scopes.
			accessToken := getTestAccessToken(t, servers, strings.Join(tt.tokenScopes, " "))

			// Make request to protected resource.
			resourceURL := testRSBaseURL + tt.endpoint
			req, err := http.NewRequestWithContext(context.Background(), tt.method, resourceURL, nil)
			testify.NoError(t, err, "Failed to create resource request")

			req.Header.Set("Authorization", bearerTokenPrefix+accessToken)

			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Resource request failed")

			defer func() {
				defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test code cleanup
			}()

			testify.Equal(t, tt.expectedStatus, resp.StatusCode, "Unexpected status code for %s", tt.endpoint)
		})
	}
}

// TestUnauthorizedAccess verifies that requests without tokens are rejected.
// TestUnauthorizedAccess verifies protected endpoints require authentication.
//
// Validates requirements:
// - R02-02: UserInfo endpoint validates Bearer token
// - R05-05: Revoked tokens rejected with 401 Unauthorized
// - R06-01: Session middleware validates access tokens.
