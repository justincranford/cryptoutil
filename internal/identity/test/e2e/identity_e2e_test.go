package e2e

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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
		AuthZURL: "https://localhost:8080",
		IDPURL:   "https://localhost:8081",
		RSURL:    "https://localhost:8082",
		SPAUrl:   "https://localhost:8083",
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
	defer resp.Body.Close()

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

// initiateAuthorizationCodeFlow initiates the OAuth 2.1 authorization code flow.
func (s *E2ETestSuite) initiateAuthorizationCodeFlow(ctx context.Context, scenario TestScenario) (string, error) {
	// TODO: Implement authorization code flow initiation
	// This will interact with the AuthZ server to obtain authorization code
	return "auth_code_placeholder", nil
}

// exchangeCodeForTokens exchanges authorization code for tokens with client authentication.
func (s *E2ETestSuite) exchangeCodeForTokens(ctx context.Context, code string, clientAuth ClientAuthMethod) (*TokenResponse, error) {
	// TODO: Implement token exchange with specified client authentication method
	return &TokenResponse{
		AccessToken:  "access_token_placeholder",
		RefreshToken: "refresh_token_placeholder",
		IDToken:      "id_token_placeholder",
	}, nil
}

// accessProtectedResource accesses a protected resource using the access token.
func (s *E2ETestSuite) accessProtectedResource(ctx context.Context, accessToken string) error {
	// TODO: Implement protected resource access
	return nil
}

// refreshAccessToken refreshes the access token using refresh token.
func (s *E2ETestSuite) refreshAccessToken(ctx context.Context, refreshToken string, clientAuth ClientAuthMethod) error {
	// TODO: Implement token refresh flow
	return nil
}

// TokenResponse represents the OAuth 2.1 token response.
type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	IDToken      string
	ExpiresIn    int
	TokenType    string
}
