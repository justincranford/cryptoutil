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
