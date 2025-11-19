// Copyright (c) 2025 Justin Cranford
//
//

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	cryptoutilIdentityPKCE "cryptoutil/internal/identity/authz/pkce"
	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
	cryptoutilIdentityServer "cryptoutil/internal/identity/server"

	googleUuid "github.com/google/uuid"
	testify "github.com/stretchr/testify/require"
)

const (
	testTimeout       = 30 * time.Second
	serverStartDelay  = 500 * time.Millisecond
	httpClientTimeout = 5 * time.Second
	shutdownTimeout   = 3 * time.Second
	testAuthZPort     = 18080
	testIDPPort       = 18081
	testRSPort        = 18082
	testAuthZBaseURL  = "http://127.0.0.1:18080"
	testIDPBaseURL    = "http://127.0.0.1:18081"
	testRSBaseURL     = "http://127.0.0.1:18082"
	testRedirectURI   = "http://127.0.0.1:18083/callback"
	testClientID      = "test-client"
	testClientSecret  = "test-secret" // pragma: allowlist secret
	testScope         = "read:resource write:resource"
	testAdminScope    = "admin"
	testUserID        = "test-user"
	testUsername      = "testuser"
	testPassword      = "testpass123" // pragma: allowlist secret
	authorizationCode = "test-auth-code"
	bearerTokenPrefix = "Bearer "
)

// testServers holds all three identity servers for integration testing.
type testServers struct {
	authzServer *cryptoutilIdentityServer.AuthZServer
	idpServer   *cryptoutilIdentityServer.IDPServer
	rsServer    *cryptoutilIdentityServer.RSServer
	httpClient  *http.Client
}

// setupTestServers creates and starts all three identity servers.
func setupTestServers(t *testing.T) (*testServers, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)

	// Create unique database name for this test.
	// Use simple :memory: for SQLite to avoid query parameter issues.
	testDBName := ":memory:"

	// Configure all three servers.
	authzConfig := &cryptoutilIdentityConfig.Config{
		AuthZ: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-authz",
			BindAddress: "127.0.0.1",
			Port:        testAuthZPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: "jws",
			Issuer:            testAuthZBaseURL,
		},
	}

	idpConfig := &cryptoutilIdentityConfig.Config{
		IDP: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-idp",
			BindAddress: "127.0.0.1",
			Port:        testIDPPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			AccessTokenFormat: "jws",
			Issuer:            testIDPBaseURL,
		},
	}

	rsConfig := &cryptoutilIdentityConfig.Config{
		RS: &cryptoutilIdentityConfig.ServerConfig{
			Name:        "test-rs",
			BindAddress: "127.0.0.1",
			Port:        testRSPort,
			TLSEnabled:  false,
		},
		Database: &cryptoutilIdentityConfig.DatabaseConfig{
			Type: "sqlite",
			DSN:  testDBName,
		},
	}

	// Initialize repository factory (shared across all servers).
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, authzConfig.Database)
	testify.NoError(t, err, "Failed to initialize repository factory")

	// Run database migrations.
	err = repoFactory.AutoMigrate(ctx)
	testify.NoError(t, err, "Failed to run database migrations")

	// Create logger for tests.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Reduce noise in test output.
	}))

	// Create token service for AuthZ and IdP.
	jwsIssuer := &cryptoutilIdentityIssuer.JWSIssuer{}
	jweIssuer := &cryptoutilIdentityIssuer.JWEIssuer{}
	uuidIssuer := &cryptoutilIdentityIssuer.UUIDIssuer{}
	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, authzConfig.Tokens)

	// Create all three servers.
	authzServer := cryptoutilIdentityServer.NewAuthZServer(authzConfig, repoFactory, tokenSvc)
	idpServer := cryptoutilIdentityServer.NewIDPServer(idpConfig, repoFactory, tokenSvc)
	rsServer, err := cryptoutilIdentityServer.NewRSServer(ctx, rsConfig, logger, tokenSvc)
	testify.NoError(t, err, "Failed to create RS server")

	// Create HTTP client for testing.
	httpClient := &http.Client{
		Timeout: httpClientTimeout,
	}

	servers := &testServers{
		authzServer: authzServer,
		idpServer:   idpServer,
		rsServer:    rsServer,
		httpClient:  httpClient,
	}

	// Start all servers in background.
	go func() {
		if err := authzServer.Start(ctx); err != nil {
			t.Logf("AuthZ server error: %v", err)
		}
	}()

	go func() {
		if err := idpServer.Start(ctx); err != nil {
			t.Logf("IdP server error: %v", err)
		}
	}()

	go func() {
		if err := rsServer.Start(ctx); err != nil {
			t.Logf("RS server error: %v", err)
		}
	}()

	// Wait for servers to start.
	time.Sleep(serverStartDelay)

	// Seed test data: create test client.
	seedTestData(t, ctx, repoFactory)

	return servers, cancel
}

// shutdownTestServers gracefully shuts down all test servers.
func shutdownTestServers(t *testing.T, servers *testServers) {
	t.Helper()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := servers.authzServer.Stop(shutdownCtx); err != nil {
		t.Logf("AuthZ server shutdown error: %v", err)
	}

	if err := servers.idpServer.Stop(shutdownCtx); err != nil {
		t.Logf("IdP server shutdown error: %v", err)
	}

	if err := servers.rsServer.Stop(shutdownCtx); err != nil {
		t.Logf("RS server shutdown error: %v", err)
	}
}

// seedTestData seeds the database with test client.
func seedTestData(t *testing.T, ctx context.Context, repoFactory *cryptoutilIdentityRepository.RepositoryFactory) {
	t.Helper()

	// Use GORM AutoMigrate now that gorm.Model duplication is fixed.
	db := repoFactory.DB()

	// Create test client using repository.
	clientRepo := repoFactory.ClientRepository()
	testClientUUID := googleUuid.Must(googleUuid.NewV7())
	now := time.Now()

	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      testClientUUID,
		ClientID:                testClientID,
		ClientSecret:            testClientSecret, // pragma: allowlist secret
		ClientType:              "confidential",
		Name:                    "Test Client",
		Description:             "Test client for integration tests",
		RedirectURIs:            []string{testRedirectURI},
		AllowedGrantTypes:       []string{"authorization_code", "client_credentials", "refresh_token"},
		AllowedResponseTypes:    []string{"code"},
		AllowedScopes:           []string{"read:resource", "write:resource", "delete:resource", "admin"},
		TokenEndpointAuthMethod: "client_secret_post",
		RequirePKCE:             true,
		PKCEChallengeMethod:     "S256",
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		IDTokenLifetime:         3600,
		Enabled:                 true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	err := clientRepo.Create(ctx, testClient)
	testify.NoError(t, err, "Failed to create test client")

	// Verify client was created.
	var clientCount int64

	err = db.Table("clients").Count(&clientCount).Error
	testify.NoError(t, err, "Failed to count clients")
	testify.Equal(t, int64(1), clientCount, "Should have exactly 1 client")
}

// TestHealthCheckEndpoints verifies all servers respond to health checks.
func TestHealthCheckEndpoints(t *testing.T) {
	servers, cancel := setupTestServers(t)
	defer cancel()
	defer shutdownTestServers(t, servers)

	tests := []struct {
		name     string
		endpoint string
	}{
		{"AuthZ Health Check", testAuthZBaseURL + "/health"},
		{"IdP Health Check", testIDPBaseURL + "/health"},
		{"RS Health Check", testRSBaseURL + "/api/v1/public/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.endpoint, nil)
			testify.NoError(t, err, "Failed to create request")

			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Health check request failed")

			defer func() {
				_ = resp.Body.Close()
			}()

			testify.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200 OK")
		})
	}
}

// TestOAuth2AuthorizationCodeFlow tests the complete OAuth 2.0 authorization code flow.
func TestOAuth2AuthorizationCodeFlow(t *testing.T) {
	servers, cancel := setupTestServers(t)
	defer cancel()
	defer shutdownTestServers(t, servers)

	// Generate PKCE code verifier and challenge (OAuth 2.1 requirement).
	codeVerifier, err := cryptoutilIdentityPKCE.GenerateCodeVerifier()
	testify.NoError(t, err, "Failed to generate code verifier")

	codeChallenge := cryptoutilIdentityPKCE.GenerateCodeChallenge(codeVerifier, "S256")

	// Step 1: Request authorization code with PKCE.
	authorizeURL := fmt.Sprintf("%s/oauth2/v1/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=test-state&code_challenge=%s&code_challenge_method=S256",
		testAuthZBaseURL, testClientID, url.QueryEscape(testRedirectURI), url.QueryEscape(testScope), codeChallenge)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, authorizeURL, nil)
	testify.NoError(t, err, "Failed to create authorize request")

	// Don't follow redirects - we need to inspect the redirect location.
	client := &http.Client{
		Timeout: httpClientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	testify.NoError(t, err, "Authorization request failed")

	defer func() {
		_ = resp.Body.Close()
	}()

	// Should get 302 redirect.
	testify.Equal(t, http.StatusFound, resp.StatusCode, "Should redirect to callback")

	// Extract authorization code from redirect location.
	location := resp.Header.Get("Location")
	testify.NotEmpty(t, location, "Redirect location should be set")

	redirectURL, err := url.Parse(location)
	testify.NoError(t, err, "Invalid redirect URL")

	code := redirectURL.Query().Get("code")
	testify.NotEmpty(t, code, "Authorization code should be present")

	state := redirectURL.Query().Get("state")
	testify.Equal(t, "test-state", state, "State parameter should match")

	// Step 2: Exchange authorization code for tokens.
	tokenURL := testAuthZBaseURL + "/oauth2/v1/token"
	tokenData := url.Values{}
	tokenData.Set("grant_type", "authorization_code")
	tokenData.Set("code", code)
	tokenData.Set("redirect_uri", testRedirectURI)
	tokenData.Set("client_id", testClientID)
	tokenData.Set("client_secret", testClientSecret)
	tokenData.Set("code_verifier", codeVerifier) // OAuth 2.1 PKCE requirement

	tokenReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, strings.NewReader(tokenData.Encode()))
	testify.NoError(t, err, "Failed to create token request")

	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := servers.httpClient.Do(tokenReq)
	testify.NoError(t, err, "Token request failed")

	defer func() {
		_ = tokenResp.Body.Close()
	}()

	testify.Equal(t, http.StatusOK, tokenResp.StatusCode, "Token request should succeed")

	var tokenResponse map[string]any

	err = json.NewDecoder(tokenResp.Body).Decode(&tokenResponse)
	testify.NoError(t, err, "Failed to decode token response")

	accessToken, ok := tokenResponse["access_token"].(string)
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
		_ = resourceResp.Body.Close()
	}()

	testify.Equal(t, http.StatusOK, resourceResp.StatusCode, "Protected resource access should succeed")
}

// TestResourceServerScopeEnforcement verifies scope-based access control.
func TestResourceServerScopeEnforcement(t *testing.T) {
	servers, cancel := setupTestServers(t)
	defer cancel()
	defer shutdownTestServers(t, servers)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		requiredScopes []string
		expectedStatus int
	}{
		{
			name:           "GET Protected Resource",
			endpoint:       "/api/v1/protected/resource",
			method:         http.MethodGet,
			requiredScopes: []string{"read:resource"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST Protected Resource",
			endpoint:       "/api/v1/protected/resource",
			method:         http.MethodPost,
			requiredScopes: []string{"write:resource"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DELETE Protected Resource",
			endpoint:       "/api/v1/protected/resource",
			method:         http.MethodDelete,
			requiredScopes: []string{"delete:resource"},
			expectedStatus: http.StatusForbidden, // Will fail if token doesn't have delete scope.
		},
		{
			name:           "Admin Users",
			endpoint:       "/api/v1/admin/users",
			method:         http.MethodGet,
			requiredScopes: []string{testAdminScope},
			expectedStatus: http.StatusForbidden, // Will fail if token doesn't have admin scope.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get access token with specified scopes.
			accessToken := getTestAccessToken(t, servers, strings.Join(tt.requiredScopes, " "))

			// Make request to protected resource.
			resourceURL := testRSBaseURL + tt.endpoint
			req, err := http.NewRequestWithContext(context.Background(), tt.method, resourceURL, nil)
			testify.NoError(t, err, "Failed to create resource request")

			req.Header.Set("Authorization", bearerTokenPrefix+accessToken)

			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Resource request failed")

			defer func() {
				_ = resp.Body.Close()
			}()

			testify.Equal(t, tt.expectedStatus, resp.StatusCode, "Unexpected status code for %s", tt.endpoint)
		})
	}
}

// TestUnauthorizedAccess verifies that requests without tokens are rejected.
func TestUnauthorizedAccess(t *testing.T) {
	servers, cancel := setupTestServers(t)
	defer cancel()
	defer shutdownTestServers(t, servers)

	endpoints := []string{
		"/api/v1/protected/resource",
		"/api/v1/admin/users",
		"/api/v1/admin/metrics",
	}

	for _, endpoint := range endpoints {
		t.Run("Unauthorized_"+endpoint, func(t *testing.T) {
			resourceURL := testRSBaseURL + endpoint
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, resourceURL, nil)
			testify.NoError(t, err, "Failed to create resource request")

			// Don't set Authorization header.
			resp, err := servers.httpClient.Do(req)
			testify.NoError(t, err, "Resource request failed")

			defer func() {
				_ = resp.Body.Close()
			}()

			testify.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return 401 Unauthorized")
		})
	}
}

// TestGracefulShutdown verifies servers shut down cleanly.
func TestGracefulShutdown(t *testing.T) {
	servers, cancel := setupTestServers(t)

	// Start servers normally.
	time.Sleep(serverStartDelay)

	// Verify servers are running.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testAuthZBaseURL+"/health", nil)
	testify.NoError(t, err, "Failed to create health check request")

	resp, err := servers.httpClient.Do(req)
	testify.NoError(t, err, "Health check failed before shutdown")

	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	testify.Equal(t, http.StatusOK, resp.StatusCode, "Server should be healthy before shutdown")

	// Cancel context to trigger shutdown.
	cancel()

	// Shutdown servers gracefully.
	shutdownTestServers(t, servers)

	// Verify servers stopped (connection should fail).
	time.Sleep(100 * time.Millisecond)

	req2, err := http.NewRequestWithContext(context.Background(), http.MethodGet, testAuthZBaseURL+"/health", nil)
	testify.NoError(t, err, "Failed to create health check request")

	resp2, err := servers.httpClient.Do(req2)
	if resp2 != nil && resp2.Body != nil {
		_ = resp2.Body.Close()
	}

	testify.Error(t, err, "Connection should fail after shutdown")
}

// getTestAccessToken helper function to obtain an access token for testing.
func getTestAccessToken(t *testing.T, servers *testServers, scope string) string {
	t.Helper()

	// Simplified token generation for testing - in real integration tests,
	// this would go through the full OAuth flow.
	tokenURL := testAuthZBaseURL + "/oauth2/v1/token"
	tokenData := url.Values{}
	tokenData.Set("grant_type", "client_credentials")
	tokenData.Set("client_id", testClientID)
	tokenData.Set("client_secret", testClientSecret)
	tokenData.Set("scope", scope)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tokenURL, strings.NewReader(tokenData.Encode()))
	testify.NoError(t, err, "Failed to create token request")

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := servers.httpClient.Do(req)
	testify.NoError(t, err, "Token request failed")

	defer func() {
		_ = resp.Body.Close()
	}()

	testify.Equal(t, http.StatusOK, resp.StatusCode, "Token request should succeed")

	body, err := io.ReadAll(resp.Body)
	testify.NoError(t, err, "Failed to read token response")

	var tokenResponse map[string]any

	err = json.Unmarshal(body, &tokenResponse)
	testify.NoError(t, err, "Failed to decode token response")

	accessToken, ok := tokenResponse["access_token"].(string)
	testify.True(t, ok, "Access token should be present")
	testify.NotEmpty(t, accessToken, "Access token should not be empty")

	return accessToken
}
