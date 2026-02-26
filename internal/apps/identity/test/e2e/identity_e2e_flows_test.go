//go:build e2e

// Copyright (c) 2025 Justin Cranford
//
//

package e2e

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	crand "crypto/rand"
	sha256 "crypto/sha256"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"strings"
)

func (s *E2ETestSuite) exchangeCodeForTokens(ctx context.Context, code string, clientAuth ClientAuthMethod) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/token", s.AuthZURL)

	// Build token request parameters
	params := url.Values{}
	params.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeAuthorizationCode)
	params.Set(cryptoutilSharedMagic.ResponseTypeCode, code)
	params.Set(cryptoutilSharedMagic.ParamRedirectURI, "https://127.0.0.1:8083/callback")

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
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
		params.Set(cryptoutilSharedMagic.ParamClientSecret, "test_client_secret")

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

		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
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

		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthTLS:
		// tls_client_auth: Mutual TLS with client certificate
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Note: In a real implementation, the client certificate would be configured in the HTTP transport

	case ClientAuthSelfSignedTLS:
		// self_signed_tls_client_auth: Mutual TLS with self-signed certificate
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")

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
	params.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeRefreshToken)
	params.Set(cryptoutilSharedMagic.GrantTypeRefreshToken, refreshToken)

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
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
		params.Set(cryptoutilSharedMagic.ParamClientSecret, "test_client_secret")

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

		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
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

		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")
		params.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
		params.Set("client_assertion", clientAssertion)

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthTLS:
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")

		req, err = http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(params.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create refresh request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	case ClientAuthSelfSignedTLS:
		params.Set(cryptoutilSharedMagic.ClaimClientID, "test_client_id")

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
	verifierBytes := make([]byte, cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes)
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
	stateBytes := make([]byte, cryptoutilSharedMagic.RealmMinTokenLengthBytes)
	if _, err := crand.Read(stateBytes); err != nil {
		panic(fmt.Sprintf("failed to generate state: %v", err))
	}

	// Base64URL encode the state.
	return base64.RawURLEncoding.EncodeToString(stateBytes)
}
