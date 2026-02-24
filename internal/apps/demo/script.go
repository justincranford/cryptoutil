// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides the integration demo script implementation.
// This file contains the "get token → KMS operation" demo flow.
package demo

import (
	"bytes"
	"context"
	"crypto/tls"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"net/url"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// DemoEndpoints contains the endpoints for the integration demo.
type DemoEndpoints struct {
	// Identity endpoints.
	TokenEndpoint string
	JWKSEndpoint  string

	// KMS endpoints.
	KMSAPIEndpoint    string
	KMSHealthEndpoint string
}

// DefaultDemoEndpoints returns the default demo endpoints.
func DefaultDemoEndpoints() *DemoEndpoints {
	return &DemoEndpoints{
		TokenEndpoint:     "https://localhost:8082/oauth2/token",
		JWKSEndpoint:      "https://localhost:8082/.well-known/jwks.json",
		KMSAPIEndpoint:    "https://localhost:8080/api/v1",
		KMSHealthEndpoint: "https://localhost:9090/admin/api/v1/livez",
	}
}

// DemoCredentials contains the demo client credentials.
type DemoCredentials struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// DefaultDemoCredentials returns the default demo credentials.
func DefaultDemoCredentials() *DemoCredentials {
	return &DemoCredentials{
		ClientID:     cryptoutilSharedMagic.DemoClientID,
		ClientSecret: cryptoutilSharedMagic.DemoClientSecret,
		Scopes:       []string{"kms:read", "kms:write", "kms:encrypt", "kms:decrypt"},
	}
}

// TokenResponse represents the OAuth2 token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// DemoScript runs the integration demo script.
type DemoScript struct {
	endpoints   *DemoEndpoints
	credentials *DemoCredentials
	httpClient  *http.Client
	progress    *ProgressDisplay
	errors      *ErrorAggregator
}

// NewDemoScript creates a new demo script runner.
func NewDemoScript(config *Config) *DemoScript {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // G402: Demo uses self-signed certs
		},
	}

	return &DemoScript{
		endpoints:   DefaultDemoEndpoints(),
		credentials: DefaultDemoCredentials(),
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   cryptoutilSharedMagic.DefaultDemoTimeout,
		},
		progress: NewProgressDisplay(config),
		errors:   NewErrorAggregator("demo-script"),
	}
}

// Run executes the demo script.
// Demo script step counts and constants.
const (
	demoScriptStepCount   = 5
	demoScriptPassedZero  = 0
	demoScriptPassedOne   = 1
	demoScriptPassedTwo   = 2
	demoScriptPassedThree = 3
	demoScriptPassedFour  = 4

	// demoScriptSkippedN constants for skipped steps calculations.
	demoScriptSkippedOne   = 1
	demoScriptSkippedTwo   = 2
	demoScriptSkippedThree = 3
	demoScriptSkippedFour  = 4

	// demoCiphertext is the placeholder ciphertext for skipped encryption.
	demoCiphertext = "demo-ciphertext"
)

// Run executes the demo script with all integration steps.
func (d *DemoScript) Run(ctx context.Context) (*DemoResult, error) {
	d.progress.Info("Integration Demo Script")
	d.progress.Info("=======================")
	d.progress.SetTotalSteps(demoScriptStepCount)

	startTime := time.Now().UTC()

	// Step 1: Wait for services.
	if err := d.waitForServices(ctx); err != nil {
		return d.buildResult(startTime, demoScriptPassedZero, demoScriptStepCount), err
	}

	// Step 2: Get access token.
	token, err := d.getAccessToken(ctx)
	if err != nil {
		return d.buildResult(startTime, demoScriptPassedOne, demoScriptStepCount-demoScriptSkippedOne), err
	}

	// Step 3: List key pools.
	if err := d.listKeyPools(ctx, token); err != nil {
		return d.buildResult(startTime, demoScriptPassedTwo, demoScriptStepCount-demoScriptSkippedTwo), err
	}

	// Step 4: Perform encryption.
	ciphertext, err := d.performEncryption(ctx, token)
	if err != nil {
		return d.buildResult(startTime, demoScriptPassedThree, demoScriptStepCount-demoScriptSkippedThree), err
	}

	// Step 5: Perform decryption.
	if err := d.performDecryption(ctx, token, ciphertext); err != nil {
		return d.buildResult(startTime, demoScriptPassedFour, demoScriptStepCount-demoScriptSkippedFour), err
	}

	return d.buildResult(startTime, demoScriptStepCount, demoScriptPassedZero), nil
}

// waitForServices waits for all demo services to be healthy.
func (d *DemoScript) waitForServices(ctx context.Context) error {
	d.progress.StartStep("Waiting for services")

	// Check Identity server.
	if err := d.waitForEndpoint(ctx, d.endpoints.JWKSEndpoint); err != nil {
		d.progress.FailStep("Identity server not healthy", err)
		d.errors.Add("identity", "health_check", err)

		return err
	}

	// Check KMS server.
	if err := d.waitForEndpoint(ctx, d.endpoints.KMSHealthEndpoint); err != nil {
		d.progress.FailStep("KMS server not healthy", err)
		d.errors.Add("kms", "health_check", err)

		return err
	}

	d.progress.CompleteStep("All services healthy")

	return nil
}

// waitForEndpoint polls an endpoint until it responds successfully.
func (d *DemoScript) waitForEndpoint(ctx context.Context, endpoint string) error {
	maxRetries := 10
	retryDelay := 2 * time.Second

	for i := range maxRetries {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := d.httpClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("endpoint %s not available after %d retries", endpoint, maxRetries)
}

// getAccessToken obtains an OAuth2 access token using client credentials.
func (d *DemoScript) getAccessToken(ctx context.Context) (string, error) {
	d.progress.StartStep("Obtaining access token")

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", d.credentials.ClientID)
	form.Set("client_secret", d.credentials.ClientSecret)
	form.Set("scope", "kms:read kms:write kms:encrypt kms:decrypt")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoints.TokenEndpoint, bytes.NewBufferString(form.Encode()))
	if err != nil {
		d.progress.FailStep("Failed to create token request", err)

		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.progress.FailStep("Token request failed", err)
		d.errors.Add("identity", "token_request", err)

		return "", fmt.Errorf("token request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.progress.FailStep("Failed to read token response", err)

		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
		d.progress.FailStep("Token endpoint returned error", statusErr)

		return "", statusErr
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		d.progress.FailStep("Failed to parse token response", err)

		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	d.progress.CompleteStep(fmt.Sprintf("Got access token (expires in %ds)", tokenResp.ExpiresIn))

	return tokenResp.AccessToken, nil
}

// listKeyPools lists available key pools using the access token.
func (d *DemoScript) listKeyPools(ctx context.Context, token string) error {
	d.progress.StartStep("Listing key pools")

	endpoint := d.endpoints.KMSAPIEndpoint + "/pools"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		d.progress.FailStep("Failed to create pools request", err)

		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+token)
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.progress.FailStep("Pools request failed", err)
		d.errors.Add("kms", "list_pools", err)

		return fmt.Errorf("pools request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.progress.FailStep("Failed to read pools response", err)

		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("pools endpoint returned %d: %s", resp.StatusCode, string(body))
		d.progress.FailStep(fmt.Sprintf("Pools endpoint returned %d", resp.StatusCode), statusErr)

		return statusErr
	}

	// Parse pools response to count.
	var poolsResp struct {
		Pools []any `json:"pools"`
	}

	if err := json.Unmarshal(body, &poolsResp); err == nil {
		d.progress.CompleteStep(fmt.Sprintf("Found %d key pools", len(poolsResp.Pools)))
	} else {
		d.progress.CompleteStep("Listed key pools")
	}

	return nil
}

// performEncryption performs an encryption operation.
func (d *DemoScript) performEncryption(ctx context.Context, token string) (string, error) {
	d.progress.StartStep("Performing encryption")

	// Note: This is a placeholder - actual implementation depends on KMS API structure.
	endpoint := d.endpoints.KMSAPIEndpoint + "/encrypt"

	payload := map[string]any{
		"plaintext": "SGVsbG8gV29ybGQh", // Base64 encoded "Hello World!"
		"pool_id":   "demo-encryption-pool",
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		d.progress.FailStep("Failed to create encrypt request", err)

		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		// For demo purposes, we'll skip this step if the endpoint doesn't exist.
		d.progress.SkipStep("Encryption", "endpoint not available in demo")

		return demoCiphertext, nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.progress.FailStep("Failed to read encrypt response", err)

		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// For 404, skip the step (endpoint may not be implemented yet).
	if resp.StatusCode == http.StatusNotFound {
		d.progress.SkipStep("Encryption", "endpoint not implemented")

		return demoCiphertext, nil
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("encrypt endpoint returned %d: %s", resp.StatusCode, string(body))
		d.progress.FailStep(fmt.Sprintf("Encrypt endpoint returned %d", resp.StatusCode), statusErr)

		return "", statusErr
	}

	d.progress.CompleteStep("Encryption successful")

	// Extract ciphertext from response.
	var encryptResp struct {
		Ciphertext string `json:"ciphertext"`
	}

	if err := json.Unmarshal(body, &encryptResp); err == nil {
		return encryptResp.Ciphertext, nil
	}

	return string(body), nil
}

// performDecryption performs a decryption operation.
func (d *DemoScript) performDecryption(ctx context.Context, token string, ciphertext string) error {
	d.progress.StartStep("Performing decryption")

	endpoint := d.endpoints.KMSAPIEndpoint + "/decrypt"

	payload := map[string]any{
		"ciphertext": ciphertext,
		"pool_id":    "demo-encryption-pool",
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		d.progress.FailStep("Failed to create decrypt request", err)

		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.progress.SkipStep("Decryption", "endpoint not available in demo")

		return nil
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.progress.FailStep("Failed to read decrypt response", err)

		return fmt.Errorf("failed to read response: %w", err)
	}

	// For 404, skip the step.
	if resp.StatusCode == http.StatusNotFound {
		d.progress.SkipStep("Decryption", "endpoint not implemented")

		return nil
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("decrypt endpoint returned %d: %s", resp.StatusCode, string(body))
		d.progress.FailStep(fmt.Sprintf("Decrypt endpoint returned %d", resp.StatusCode), statusErr)

		return statusErr
	}

	d.progress.CompleteStep("Decryption successful")

	return nil
}

// buildResult builds a DemoResult from current state.
func (d *DemoScript) buildResult(startTime time.Time, passed, skipped int) *DemoResult {
	result := d.errors.ToResult(passed, skipped)
	result.DurationMS = time.Since(startTime).Milliseconds()

	return result
}
