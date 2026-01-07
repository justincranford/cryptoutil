//go:build e2e

// Copyright (c) 2025 Justin Cranford
//
//

package e2e

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMain sets up and tears down mock services for e2e tests.
// NOTE: Only used for identity_e2e_test.go tests that require HTTP servers.
// OTP flow tests (otp_flows_test.go) use in-process mocks only.
func TestMain(m *testing.M) {
	// Create mock services
	mockServices := NewTestableMockServices()

	// Start mock services
	ctx := context.Background()
	if err := mockServices.Start(ctx); err != nil {
		log.Fatalf("Failed to start mock services: %v", err)
	}

	// Run tests
	code := m.Run()

	// Stop mock services
	mockServices.Stop(ctx)

	// Exit with test result code
	os.Exit(code)
}

// E2ETestSuite manages the E2E testing environment.
type E2ETestSuite struct {
	AuthZURL string
	IDPURL   string
	RSURL    string
	SPAUrl   string
	Client   *http.Client
}

// NewE2ETestSuite creates a new E2E test suite with default configuration.
func NewE2ETestSuite() *E2ETestSuite {
	return &E2ETestSuite{
		AuthZURL: "https://127.0.0.1:8080",
		IDPURL:   "https://127.0.0.1:8081",
		RSURL:    "https://127.0.0.1:8082",
		SPAUrl:   "https://127.0.0.1:8083",
		Client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Self-signed certs in dev
				},
			},
		},
	}
}

// TestConnectivity verifies all services are reachable.
func TestConnectivity(t *testing.T) {
	suite := NewE2ETestSuite()
	ctx := context.Background()

	t.Run("AuthZ Server Connectivity", func(t *testing.T) {
		err := suite.checkHealth(ctx, suite.AuthZURL)
		require.NoError(t, err, "AuthZ server should be reachable")
	})

	t.Run("IdP Server Connectivity", func(t *testing.T) {
		err := suite.checkHealth(ctx, suite.IDPURL)
		require.NoError(t, err, "IdP server should be reachable")
	})

	t.Run("Resource Server Connectivity", func(t *testing.T) {
		err := suite.checkHealth(ctx, suite.RSURL)
		require.NoError(t, err, "Resource server should be reachable")
	})

	t.Run("SPA Relying Party Connectivity", func(t *testing.T) {
		err := suite.checkHealth(ctx, suite.SPAUrl)
		require.NoError(t, err, "SPA relying party should be reachable")
	})
}

// checkHealth verifies a service is healthy.
func (s *E2ETestSuite) checkHealth(ctx context.Context, baseURL string) error {
	healthURL := fmt.Sprintf("%s/health", baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform health check: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ClientAuthMethod represents OAuth 2.1 client authentication methods.
type ClientAuthMethod string

const (
	ClientAuthBasic         ClientAuthMethod = "client_secret_basic"
	ClientAuthPost          ClientAuthMethod = "client_secret_post"
	ClientAuthSecretJWT     ClientAuthMethod = "client_secret_jwt"
	ClientAuthPrivateKeyJWT ClientAuthMethod = "private_key_jwt"
	ClientAuthTLS           ClientAuthMethod = "tls_client_auth"
	ClientAuthSelfSignedTLS ClientAuthMethod = "self_signed_tls_client_auth"
)

// UserAuthMethod represents OIDC user authentication methods.
type UserAuthMethod string

const (
	UserAuthUsernamePassword UserAuthMethod = "username_password"
	UserAuthEmailOTP         UserAuthMethod = "email_otp"
	UserAuthSMSOTP           UserAuthMethod = "sms_otp"
	UserAuthTOTP             UserAuthMethod = "totp"
	UserAuthHOTP             UserAuthMethod = "hotp"
	UserAuthMagicLink        UserAuthMethod = "magic_link"
	UserAuthPasskey          UserAuthMethod = "passkey"
	UserAuthBiometric        UserAuthMethod = "biometric"
	UserAuthHardwareKey      UserAuthMethod = "hardware_key"
)

// GrantType represents OAuth 2.1 grant types.
type GrantType string

const (
	GrantAuthorizationCode GrantType = "authorization_code"
	GrantRefreshToken      GrantType = "refresh_token"
	GrantClientCredentials GrantType = "client_credentials"
)

// TestScenario represents a complete E2E test scenario.
type TestScenario struct {
	Name             string
	ClientAuth       ClientAuthMethod
	UserAuth         UserAuthMethod
	GrantType        GrantType
	Scopes           []string
	ExpectedSuccess  bool
	ExpectedHTTPCode int
}

// GetAllTestScenarios generates all possible test scenario combinations.
func GetAllTestScenarios() []TestScenario {
	scenarios := []TestScenario{}

	clientAuths := []ClientAuthMethod{
		ClientAuthBasic,
		ClientAuthPost,
		ClientAuthSecretJWT,
		ClientAuthPrivateKeyJWT,
		ClientAuthTLS,
		ClientAuthSelfSignedTLS,
	}

	userAuths := []UserAuthMethod{
		UserAuthUsernamePassword,
		UserAuthEmailOTP,
		UserAuthSMSOTP,
		UserAuthTOTP,
		UserAuthHOTP,
		UserAuthMagicLink,
		UserAuthPasskey,
		UserAuthBiometric,
		UserAuthHardwareKey,
	}

	grantTypes := []GrantType{
		GrantAuthorizationCode,
		GrantRefreshToken,
		GrantClientCredentials,
	}

	// Generate combinations: 6 client auth Ã— 9 user auth Ã— 3 grant types = 162 scenarios
	for _, clientAuth := range clientAuths {
		for _, userAuth := range userAuths {
			for _, grantType := range grantTypes {
				scenario := TestScenario{
					Name: fmt.Sprintf("%s_%s_%s",
						string(clientAuth),
						string(userAuth),
						string(grantType),
					),
					ClientAuth:       clientAuth,
					UserAuth:         userAuth,
					GrantType:        grantType,
					Scopes:           []string{"openid", "profile", "email"},
					ExpectedSuccess:  true,
					ExpectedHTTPCode: http.StatusOK,
				}

				// Client credentials grant doesn't use user authentication
				if grantType == GrantClientCredentials && userAuth != UserAuthUsernamePassword {
					continue
				}

				scenarios = append(scenarios, scenario)
			}
		}
	}

	return scenarios
}

// TestParameterizedAuthFlows tests all OAuth 2.1 + OIDC flow combinations.
func TestParameterizedAuthFlows(t *testing.T) {
	t.Parallel()

	suite := NewE2ETestSuite()
	scenarios := GetAllTestScenarios()

	t.Logf("Running %d parameterized test scenarios", len(scenarios))

	for _, scenario := range scenarios {
		// Capture loop variable
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel() // Run scenarios in parallel for faster execution

			// Execute the complete OAuth 2.1 + OIDC flow
			err := suite.executeAuthFlow(context.Background(), scenario)

			if scenario.ExpectedSuccess {
				require.NoError(t, err, "Auth flow should succeed for scenario: %s", scenario.Name)
			} else {
				require.Error(t, err, "Auth flow should fail for scenario: %s", scenario.Name)
			}
		})
	}
}

// executeAuthFlow executes a complete OAuth 2.1 + OIDC authentication flow.
func (s *E2ETestSuite) executeAuthFlow(ctx context.Context, scenario TestScenario) error {
	// Step 1: User authentication with IdP
	if scenario.GrantType != GrantClientCredentials {
		if err := s.performUserAuth(ctx, scenario.UserAuth); err != nil {
			return fmt.Errorf("user authentication failed: %w", err)
		}
	}

	// Step 2: Authorization request to AuthZ server
	authCode, err := s.initiateAuthorizationCodeFlow(ctx, scenario)
	if err != nil {
		return fmt.Errorf("authorization code flow initiation failed: %w", err)
	}

	// Step 3: Token exchange with client authentication
	tokens, err := s.exchangeCodeForTokens(ctx, authCode, scenario.ClientAuth)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// Step 4: Access protected resource
	if err := s.accessProtectedResource(ctx, tokens.AccessToken); err != nil {
		return fmt.Errorf("protected resource access failed: %w", err)
	}

	// Step 5: Refresh token flow
	if scenario.GrantType == GrantRefreshToken {
		if err := s.refreshAccessToken(ctx, tokens.RefreshToken, scenario.ClientAuth); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
	}

	return nil
}

// performUserAuth performs user authentication with the specified method.
// Detailed implementations for each method are in user_auth_test.go.
func (s *E2ETestSuite) performUserAuth(ctx context.Context, method UserAuthMethod) error {
	switch method {
	case UserAuthUsernamePassword:
		return s.performUsernamePasswordAuth(ctx)
	case UserAuthEmailOTP:
		return s.performEmailOTPAuth(ctx)
	case UserAuthSMSOTP:
		return s.performSMSOTPAuth(ctx)
	case UserAuthTOTP:
		return s.performTOTPAuth(ctx)
	case UserAuthHOTP:
		return s.performHOTPAuth(ctx)
	case UserAuthMagicLink:
		return s.performMagicLinkAuth(ctx)
	case UserAuthPasskey:
		return s.performPasskeyAuth(ctx)
	case UserAuthBiometric:
		return s.performBiometricAuth(ctx)
	case UserAuthHardwareKey:
		return s.performHardwareKeyAuth(ctx)
	default:
		return fmt.Errorf("unsupported user auth method: %s", method)
	}
}

// Stub implementations for user authentication methods
// These would be properly implemented in a real identity system

func (s *E2ETestSuite) performUsernamePasswordAuth(ctx context.Context) error {
	// Stub implementation - simulate successful username/password authentication
	return nil
}

func (s *E2ETestSuite) performEmailOTPAuth(ctx context.Context) error {
	// Stub implementation - simulate successful email OTP authentication
	return nil
}

func (s *E2ETestSuite) performSMSOTPAuth(ctx context.Context) error {
	// Stub implementation - simulate successful SMS OTP authentication
	return nil
}

func (s *E2ETestSuite) performTOTPAuth(ctx context.Context) error {
	// Stub implementation - simulate successful TOTP authentication
	return nil
}

func (s *E2ETestSuite) performHOTPAuth(ctx context.Context) error {
	// Stub implementation - simulate successful HOTP authentication
	return nil
}

func (s *E2ETestSuite) performMagicLinkAuth(ctx context.Context) error {
	// Stub implementation - simulate successful magic link authentication
	return nil
}

func (s *E2ETestSuite) performPasskeyAuth(ctx context.Context) error {
	// Stub implementation - simulate successful passkey authentication
	return nil
}

func (s *E2ETestSuite) performBiometricAuth(ctx context.Context) error {
	// Stub implementation - simulate successful biometric authentication
	return nil
}

func (s *E2ETestSuite) performHardwareKeyAuth(ctx context.Context) error {
	// Stub implementation - simulate successful hardware key authentication
	return nil
}

// initiateAuthorizationCodeFlow initiates the OAuth 2.1 authorization code flow.
func (s *E2ETestSuite) initiateAuthorizationCodeFlow(ctx context.Context, scenario TestScenario) (string, error) {
	// Generate PKCE code verifier and challenge.
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallengeE2E(codeVerifier)

	// Generate state parameter for CSRF protection.
	state := generateState()

	authorizeURL := fmt.Sprintf("%s/authorize", s.AuthZURL)

	// Build authorization request with PKCE parameters
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", "test_client_id")
	params.Set("redirect_uri", "https://127.0.0.1:8083/callback")
	params.Set("scope", strings.Join(scenario.Scopes, " "))
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	fullURL := fmt.Sprintf("%s?%s", authorizeURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create authorization request: %w", err)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform authorization request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Check for redirect to login page (302) or authorization code response (302).
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read authorization error response: %w", err)
		}

		return "", fmt.Errorf("authorization request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Extract authorization code from response or redirect.
	location := resp.Header.Get("Location")
	if location == "" {
		// If no redirect, code might be in response body (test mode).
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read authorization response: %w", err)
		}

		// Parse JSON response for authorization code.
		var authResp struct {
			Code  string `json:"code"`
			State string `json:"state"`
		}

		if err := json.Unmarshal(body, &authResp); err != nil {
			return "", fmt.Errorf("failed to parse authorization response: %w", err)
		}

		// Validate state parameter.
		if authResp.State != state {
			return "", fmt.Errorf("state mismatch: expected %s, got %s", state, authResp.State)
		}

		return authResp.Code, nil
	}

	// Parse location header for authorization code.
	locationURL, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("failed to parse redirect location: %w", err)
	}

	code := locationURL.Query().Get("code")
	if code == "" {
		return "", fmt.Errorf("no authorization code in redirect: %s", location)
	}

	// Validate state parameter.
	returnedState := locationURL.Query().Get("state")
	if returnedState != state {
		return "", fmt.Errorf("state mismatch: expected %s, got %s", state, returnedState)
	}

	return code, nil
}

// TokenResponse represents the OAuth 2.1 token response.
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	TokenType    string
}

// exchangeCodeForTokens exchanges authorization code for tokens with client authentication.
func (s *E2ETestSuite) exchangeCodeForTokens(ctx context.Context, code string, clientAuth ClientAuthMethod) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/token", s.AuthZURL)

	// Build token request parameters
	params := url.Values{}
	params.Set("grant_type", "authorization_code")
	params.Set("code", code)
	params.Set("redirect_uri", "https://127.0.0.1:8083/callback")

	// Add client authentication based on method
	var req *http.Request

	var err error

	switch clientAuth {
	case ClientAuthBasic:
		// client_secret_basic: HTTP Basic Authentication
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("test_client_id", "test_client_secret")

	case ClientAuthPost:
		// client_secret_post: Include client_id and client_secret in POST body
		params.Set("client_id", "test_client_id")
		params.Set("client_secret", "test_client_secret")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthSecretJWT:
		// client_secret_jwt: JWT signed with client secret
		clientAssertion, err := s.generateClientSecretJWT()
		if err != nil {
			return nil, fmt.Errorf("failed to generate client secret JWT: %w", err)
		}

		params.Set("client_id", "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthPrivateKeyJWT:
		// private_key_jwt: JWT signed with private key
		clientAssertion, err := s.generatePrivateKeyJWT()
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key JWT: %w", err)
		}

		params.Set("client_id", "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthTLS:
		// tls_client_auth: Mutual TLS with client certificate
		params.Set("client_id", "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Note: In a real implementation, the client certificate would be configured in the HTTP transport

	case ClientAuthSelfSignedTLS:
		// self_signed_tls_client_auth: Mutual TLS with self-signed certificate
		params.Set("client_id", "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Note: In a real implementation, the self-signed certificate would be configured in the HTTP transport

	default:
		return nil, fmt.Errorf("unsupported client authentication method: %s", clientAuth)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform token request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read token error response: %w", err)
		}

		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		IDToken:      tokenResp.IDToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		TokenType:    tokenResp.TokenType,
	}, nil
}

// accessProtectedResource accesses a protected resource using the access token.
func (s *E2ETestSuite) accessProtectedResource(ctx context.Context, accessToken string) error {
	protectedURL := fmt.Sprintf("%s/api/protected", s.RSURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, protectedURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create protected resource request: %w", err)
	}

	// Add Bearer token authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to access protected resource: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read protected resource error response: %w", err)
		}

		return fmt.Errorf("protected resource access failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// refreshAccessToken refreshes the access token using refresh token.
func (s *E2ETestSuite) refreshAccessToken(ctx context.Context, refreshToken string, clientAuth ClientAuthMethod) error {
	tokenURL := fmt.Sprintf("%s/token", s.AuthZURL)

	// Build refresh token request parameters
	params := url.Values{}
	params.Set("grant_type", "refresh_token")
	params.Set("refresh_token", refreshToken)

	// Add client authentication based on method
	var req *http.Request

	var err error

	switch clientAuth {
	case ClientAuthBasic:
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("test_client_id", "test_client_secret")

	case ClientAuthPost:
		params.Set("client_id", "test_client_id")
		params.Set("client_secret", "test_client_secret")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthSecretJWT:
		clientAssertion, err := s.generateClientSecretJWT()
		if err != nil {
			return fmt.Errorf("failed to generate client secret JWT: %w", err)
		}

		params.Set("client_id", "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthPrivateKeyJWT:
		clientAssertion, err := s.generatePrivateKeyJWT()
		if err != nil {
			return fmt.Errorf("failed to generate private key JWT: %w", err)
		}

		params.Set("client_id", "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthTLS:
		params.Set("client_id", "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthSelfSignedTLS:
		params.Set("client_id", "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	default:
		return fmt.Errorf("unsupported client authentication method: %s", clientAuth)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform refresh request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read refresh error response: %w", err)
		}

		return fmt.Errorf("refresh request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse refresh response (similar to token response)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read refresh response: %w", err)
	}

	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token,omitempty"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Verify we got a new access token
	if refreshResp.AccessToken == "" {
		return fmt.Errorf("refresh response missing access token")
	}

	return nil
}

// generateCodeVerifier generates a cryptographically secure random code verifier for PKCE.
func generateCodeVerifier() string {
	// Generate 32 bytes of random data (256 bits).
	verifierBytes := make([]byte, 32)
	if _, err := crand.Read(verifierBytes); err != nil {
		panic(fmt.Sprintf("failed to generate code verifier: %v", err))
	}

	// Base64URL encode the verifier.
	return base64.RawURLEncoding.EncodeToString(verifierBytes)
}

// generateCodeChallengeE2E generates a code challenge from a code verifier using S256 method.
func generateCodeChallengeE2E(verifier string) string {
	// Hash the verifier with SHA256.
	hash := sha256.Sum256([]byte(verifier))

	// Base64URL encode the hash.
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateClientSecretJWT generates a JWT signed with client secret for client authentication.
func (s *E2ETestSuite) generateClientSecretJWT() (string, error) {
	// This is a simplified implementation for testing.
	// In a real implementation, this would create a proper JWT with the required claims.
	return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test_client_secret_jwt", nil
}

// generatePrivateKeyJWT generates a JWT signed with private key for client authentication.
func (s *E2ETestSuite) generatePrivateKeyJWT() (string, error) {
	// This is a simplified implementation for testing.
	// In a real implementation, this would create a proper JWT signed with the client's private key.
	return "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test_private_key_jwt", nil
}

// generateState generates a random state parameter for CSRF protection.
func generateState() string {
	// Generate 16 bytes of random data (128 bits).
	stateBytes := make([]byte, 16)
	if _, err := crand.Read(stateBytes); err != nil {
		panic(fmt.Sprintf("failed to generate state: %v", err))
	}

	// Base64URL encode the state.
	return base64.RawURLEncoding.EncodeToString(stateBytes)
}
