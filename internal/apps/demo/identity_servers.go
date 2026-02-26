// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides Identity demo implementation.
package demo

import (
	"bytes"
	"context"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"strings"
	"time"

	cryptoutilSharedUtilPoll "cryptoutil/internal/shared/util/poll"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func stopIdentityServer(demoServer *identityDemoServer) {
	if demoServer == nil {
		return
	}

	// Cancel context to trigger shutdown.
	if demoServer.cancelFunc != nil {
		demoServer.cancelFunc()
	}

	// Stop server gracefully.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), identityShutdownTimeout)
	defer shutdownCancel()

	if demoServer.server != nil {
		_ = demoServer.server.Stop(shutdownCtx)
	}

	// Close repository.
	if demoServer.repoFactory != nil {
		_ = demoServer.repoFactory.Close()
	}
}

// waitForIdentityHealth waits for Identity server health checks to pass.
func waitForIdentityHealth(ctx context.Context, demoServer *identityDemoServer, timeout time.Duration) error {
	healthURL := demoServer.baseURL + "/health"
	client := &http.Client{Timeout: identityHTTPClientTimeout}

	if err := cryptoutilSharedUtilPoll.Until(ctx, timeout, identityHealthInterval, func(pollCtx context.Context) (bool, error) {
		return isHTTPHealthy(pollCtx, client, healthURL), nil
	}); err != nil {
		return fmt.Errorf("identity health check failed: %w", err)
	}

	return nil
}

// verifyOpenIDConfiguration verifies the OpenID configuration endpoint.
func verifyOpenIDConfiguration(ctx context.Context, demoServer *identityDemoServer, progress *ProgressDisplay) error {
	configURL := demoServer.baseURL + cryptoutilSharedMagic.PathDiscovery

	client := &http.Client{Timeout: identityHTTPLongTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch OpenID configuration: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("failed to parse OpenID configuration: %w", err)
	}

	// Verify required fields.
	requiredFields := []string{"issuer", "token_endpoint", "jwks_uri"}
	for _, field := range requiredFields {
		if _, ok := config[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	progress.Debug(fmt.Sprintf("OpenID Configuration: issuer=%s", config["issuer"]))
	progress.Debug(fmt.Sprintf("  token_endpoint: %s", config["token_endpoint"]))
	progress.Debug(fmt.Sprintf("  jwks_uri: %s", config["jwks_uri"]))

	return nil
}

// demonstrateClientCredentialsFlow demonstrates the OAuth 2.1 client_credentials flow.
func demonstrateClientCredentialsFlow(ctx context.Context, demoServer *identityDemoServer, progress *ProgressDisplay) error {
	tokenURL := demoServer.baseURL + "/oauth2/v1/token"

	// Build token request using demo client credentials.
	clientID := cryptoutilSharedMagic.DemoClientID
	clientSecret := cryptoutilSharedMagic.DemoClientSecret

	form := url.Values{}
	form.Set(cryptoutilSharedMagic.ParamGrantType, cryptoutilSharedMagic.GrantTypeClientCredentials)
	form.Set(cryptoutilSharedMagic.ClaimScope, "read write")

	client := &http.Client{Timeout: identityHTTPLongTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic auth for client authentication.
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	progress.Debug(fmt.Sprintf("Requesting token from: %s", tokenURL))
	progress.Debug(fmt.Sprintf("  client_id: %s", clientID))
	progress.Debug("  grant_type: client_credentials")
	progress.Debug("  scope: read write")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp map[string]any
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Verify token response.
	accessToken, ok := tokenResp[cryptoutilSharedMagic.TokenTypeAccessToken].(string)
	if !ok || accessToken == "" {
		return fmt.Errorf("missing access_token in response")
	}

	tokenType, _ := tokenResp[cryptoutilSharedMagic.ParamTokenType].(string)
	expiresIn, _ := tokenResp[cryptoutilSharedMagic.ParamExpiresIn].(float64)

	progress.Debug("Token response received successfully:")
	progress.Debug(fmt.Sprintf("  token_type: %s", tokenType))
	progress.Debug(fmt.Sprintf("  expires_in: %.0f seconds", expiresIn))
	progress.Debug(fmt.Sprintf("  access_token: %s... (truncated)", accessToken[:min(identityTokenTruncateLen, len(accessToken))]))

	// Decode and display JWT claims (for demo purposes).
	parts := strings.Split(accessToken, ".")
	if len(parts) == identityJWTPartCount {
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var claims map[string]any
			if json.Unmarshal(payload, &claims) == nil {
				var prettyJSON bytes.Buffer
				if json.Indent(&prettyJSON, payload, "    ", "  ") == nil {
					progress.Debug("  JWT payload:")

					for _, line := range strings.Split(prettyJSON.String(), "\n") {
						progress.Debug("    " + line)
					}
				}
			}
		}
	}

	return nil
}
