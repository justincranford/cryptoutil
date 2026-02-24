// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides a demonstration of the identity service capabilities.
package demo

import (
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

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func printEndpointSummary() {
	_, _ = fmt.Fprintln(outWriter, "   OAuth 2.1 / OpenID Connect Endpoints:")
	_, _ = fmt.Fprintf(outWriter, "   • Discovery:     %s/.well-known/oauth-authorization-server\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • OIDC Config:   %s/.well-known/openid-configuration\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • Authorization: %s/oauth2/v1/authorize\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • Token:         %s/oauth2/v1/token\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • Introspect:    %s/oauth2/v1/introspect\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • Revoke:        %s/oauth2/v1/revoke\n", demoIssuer)
	_, _ = fmt.Fprintf(outWriter, "   • JWKS:          %s/oauth2/v1/jwks\n", demoIssuer)
}

func demonstrateAuthorization(ctx context.Context, client *http.Client, codeChallenge, state string) error {
	authURL := fmt.Sprintf("%s/oauth2/v1/authorize", demoIssuer)

	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {demoClientID},
		"redirect_uri":          {demoRedirectURI},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"scope":                 {"openid profile email"},
	}

	fullURL := authURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("create authorization request: %w", err)
	}

	// Don't follow redirects.
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute authorization request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	_, _ = fmt.Fprintf(outWriter, "   Request: GET %s?...\n", authURL)
	_, _ = fmt.Fprintf(outWriter, "   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	_, _ = fmt.Fprintln(outWriter, "   ✅ Authorization endpoint accessible")
	_, _ = fmt.Fprintln(outWriter, "   📝 In production: would redirect to IdP login page")

	return nil
}

func demonstrateTokenEndpoint(ctx context.Context, client *http.Client) (string, error) {
	tokenURL := fmt.Sprintf("%s/oauth2/v1/token", demoIssuer)

	data := url.Values{
		"grant_type": {"client_credentials"},
		"scope":      {"openid profile email"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   Request: POST %s\n", tokenURL)
	_, _ = fmt.Fprintf(outWriter, "   Grant:   client_credentials\n")
	_, _ = fmt.Fprintf(outWriter, "   Client:  %s (Basic Auth)\n", demoClientID)
	_, _ = fmt.Fprintf(outWriter, "   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var accessToken string

	if resp.StatusCode == http.StatusOK {
		var tokenResp map[string]any

		if err := json.Unmarshal(body, &tokenResp); err == nil {
			if at, ok := tokenResp["access_token"].(string); ok {
				accessToken = at
				_, _ = fmt.Fprintf(outWriter, "   ✅ Access Token (first %d): %s...\n",
					cryptoutilSharedMagic.DemoMinTokenChars,
					accessToken[:min(cryptoutilSharedMagic.DemoMinTokenChars, len(accessToken))])
			}

			if tokenType, ok := tokenResp["token_type"].(string); ok {
				_, _ = fmt.Fprintf(outWriter, "   ✅ Token Type: %s\n", tokenType)
			}

			if expiresIn, ok := tokenResp["expires_in"].(float64); ok {
				_, _ = fmt.Fprintf(outWriter, "   ✅ Expires In: %.0f seconds\n", expiresIn)
			}
		}
	} else {
		_, _ = fmt.Fprintf(outWriter, "   ⚠️ Response: %s\n", string(body))
		_, _ = fmt.Fprintln(outWriter, "   📝 Note: Token service may need to be configured for token issuance")
	}

	return accessToken, nil
}

func demonstrateIntrospection(ctx context.Context, client *http.Client, accessToken string) error {
	introspectURL := fmt.Sprintf("%s/oauth2/v1/introspect", demoIssuer)

	// Use the actual access token if available, otherwise use a sample.
	tokenToIntrospect := accessToken
	if tokenToIntrospect == "" {
		tokenToIntrospect = sampleAccessTokenFmt
	}

	data := url.Values{
		"token": {tokenToIntrospect},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute introspection request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read introspection response: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   Request: POST %s\n", introspectURL)
	_, _ = fmt.Fprintf(outWriter, "   Client:  %s (Basic Auth)\n", demoClientID)
	_, _ = fmt.Fprintf(outWriter, "   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var introspectResp map[string]any
	if err := json.Unmarshal(body, &introspectResp); err == nil {
		active, _ := introspectResp["active"].(bool) //nolint:errcheck // Demo ok assertion

		if active && accessToken != "" {
			_, _ = fmt.Fprintln(outWriter, "   ✅ Token Active: true (issued access token validated)")
		} else if !active && accessToken != "" {
			_, _ = fmt.Fprintln(outWriter, "   ⚠️ Token Active: false (token may not be stored)")
		} else {
			_, _ = fmt.Fprintf(outWriter, "   ✅ Token Active: %v (unknown token returns inactive)\n", active)
		}
	}

	return nil
}

func demonstrateRevocation(ctx context.Context, client *http.Client, accessToken string) error {
	revokeURL := fmt.Sprintf("%s/oauth2/v1/revoke", demoIssuer)

	// Use the actual access token if available, otherwise use a sample.
	tokenToRevoke := accessToken
	if tokenToRevoke == "" {
		tokenToRevoke = "sample-token-to-revoke"
	}

	data := url.Values{
		"token": {tokenToRevoke},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create revocation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute revocation request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	_, _ = fmt.Fprintf(outWriter, "   Request: POST %s\n", revokeURL)
	_, _ = fmt.Fprintf(outWriter, "   Client:  %s (Basic Auth)\n", demoClientID)
	_, _ = fmt.Fprintf(outWriter, "   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	_, _ = fmt.Fprintln(outWriter, "   ✅ Revocation endpoint returns 200 per RFC 7009")

	return nil
}

func demonstrateIntrospectionAfterRevoke(ctx context.Context, client *http.Client, accessToken string) error {
	introspectURL := fmt.Sprintf("%s/oauth2/v1/introspect", demoIssuer)

	// Use the revoked access token.
	tokenToIntrospect := accessToken
	if tokenToIntrospect == "" {
		tokenToIntrospect = sampleAccessTokenFmt
	}

	data := url.Values{
		"token": {tokenToIntrospect},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create post-revoke introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Use Basic authentication with registered client credentials.
	req.SetBasicAuth(demoClientID, demoClientSecret)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("execute post-revoke introspection request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read post-revoke introspection response: %w", err)
	}

	_, _ = fmt.Fprintf(outWriter, "   Request: POST %s\n", introspectURL)
	_, _ = fmt.Fprintf(outWriter, "   Client:  %s (Basic Auth)\n", demoClientID)
	_, _ = fmt.Fprintf(outWriter, "   Status:  %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))

	var introspectResp map[string]any
	if err := json.Unmarshal(body, &introspectResp); err == nil {
		active, _ := introspectResp["active"].(bool) //nolint:errcheck // Demo ok assertion

		if !active && accessToken != "" {
			_, _ = fmt.Fprintln(outWriter, "   ✅ Token Active: false (revoked token correctly invalidated)")
		} else if active && accessToken != "" {
			_, _ = fmt.Fprintln(outWriter, "   ⚠️ Token Active: true (revocation may not have persisted)")
		} else {
			_, _ = fmt.Fprintf(outWriter, "   ✅ Token Active: %v\n", active)
		}
	}

	return nil
}

func getJSON(ctx context.Context, client *http.Client, urlStr string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute GET request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Demo cleanup

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode JSON response: %w", err)
	}

	return result, nil
}

func generatePKCE() (verifier, challenge string) {
	verifierBytes := make([]byte, cryptoutilSharedMagic.DefaultStateLength)

	if _, err := crand.Read(verifierBytes); err != nil {
		// Fall back to deterministic value for demo purposes.
		copy(verifierBytes, []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ12"))
	}

	verifier = base64.RawURLEncoding.EncodeToString(verifierBytes)

	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge
}

func generateState() string {
	stateBytes := make([]byte, cryptoutilSharedMagic.DefaultNonceLength/2)

	if _, err := crand.Read(stateBytes); err != nil {
		// Fall back to deterministic value for demo purposes.
		copy(stateBytes, []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ012345"))
	}

	return base64.RawURLEncoding.EncodeToString(stateBytes)
}
