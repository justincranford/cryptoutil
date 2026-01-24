// Copyright (c) 2025 Justin Cranford

package demo

import (
	"bytes"
	"context"
	json "encoding/json"
	"fmt"
	"io"
	http "net/http"
	"os"
	"time"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilJoseServer "cryptoutil/internal/jose/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// JOSE demo step counts.
const joseStepCount = 9

// runJOSEDemo runs the JOSE Authority Server demonstration.
// Demonstrates JWK generation, JWS signing/verification, and JWT operations.
func runJOSEDemo(ctx context.Context, config *Config) int {
	progress := NewProgressDisplay(config)
	errors := NewErrorAggregator("jose")
	startTime := time.Now()
	passedSteps := 0

	progress.Info("Starting JOSE Authority Demo")
	progress.Info("============================")
	progress.SetTotalSteps(joseStepCount)

	// Step 1: Start JOSE Authority Server.
	progress.StartStep("Starting JOSE Authority Server")

	settings := cryptoutilAppsTemplateServiceConfig.NewForJOSEServer(
		cryptoutilSharedMagic.IPv4Loopback,
		0, // Dynamic port.
		true,
	)

	// Create TLS configuration for JOSE server.
	tlsCfg, err := cryptoutilAppsTemplateServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{"localhost", "jose-server"},
		[]string{"127.0.0.1", "::1"},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	if err != nil {
		progress.FailStep("Generating TLS config", fmt.Errorf("TLS generation failed: %w", err))
		errors.Add("tls", "failed to generate TLS config", err)

		return exitJOSEDemo(progress, errors, passedSteps, startTime)
	}

	server, err := cryptoutilJoseServer.NewServer(ctx, settings, tlsCfg)
	if err != nil {
		progress.FailStep("Starting server", fmt.Errorf("create failed: %w", err))
		errors.Add("server", "failed to create server", err)

		return exitJOSEDemo(progress, errors, passedSteps, startTime)
	}

	if err := server.StartNonBlocking(); err != nil {
		progress.FailStep("Starting server", fmt.Errorf("start failed: %w", err))
		errors.Add("server", "failed to start server", err)

		return exitJOSEDemo(progress, errors, passedSteps, startTime)
	}

	defer func() {
		if shutdownErr := server.Shutdown(); shutdownErr != nil {
			progress.Error("Server shutdown error: %v", shutdownErr)
		}
	}()

	// Wait for server to be ready.
	time.Sleep(cryptoutilSharedMagic.ServerStartupWait)

	baseURL := fmt.Sprintf("http://%s:%d", cryptoutilSharedMagic.IPv4Loopback, server.ActualPort())
	client := &http.Client{Timeout: cryptoutilSharedMagic.DefaultDemoTimeout}

	progress.CompleteStep(fmt.Sprintf("Server started at %s", baseURL))

	passedSteps++

	// Step 2: Check health.
	progress.StartStep("Checking server health")

	resp, err := doJOSEGet(ctx, client, baseURL+"/health")
	if err != nil {
		progress.FailStep("Health check", err)
		errors.Add("health", "health check failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}

		if resp.StatusCode != http.StatusOK {
			progress.FailStep("Health check", fmt.Errorf("status %d", resp.StatusCode))

			if !config.ContinueOnError {
				return exitJOSEDemo(progress, errors, passedSteps, startTime)
			}
		} else {
			progress.CompleteStep("Health check passed")

			passedSteps++
		}
	}

	// Step 3: Generate EC key.
	progress.StartStep("Generating EC P-384 signing key")

	keyResp, err := generateJWK(ctx, client, baseURL, "EC/P384", "sig")
	if err != nil {
		progress.FailStep("Key generation", err)
		errors.Add("keygen", "key generation failed", err)

		return exitJOSEDemo(progress, errors, passedSteps, startTime)
	}

	progress.CompleteStep(fmt.Sprintf("Generated key: kid=%s", keyResp.KID))

	passedSteps++

	// Step 4: Sign a message.
	progress.StartStep("Signing message with JWS")

	jws, err := signJWS(ctx, client, baseURL, keyResp.KID, "Hello, JOSE Authority!")
	if err != nil {
		progress.FailStep("JWS signing", err)
		errors.Add("sign", "JWS signing failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("Created JWS (length: %d)", len(jws)))

		passedSteps++
	}

	// Step 5: Verify the signature.
	progress.StartStep("Verifying JWS signature")

	valid, payload, err := verifyJWS(ctx, client, baseURL, keyResp.KID, jws)
	if err != nil {
		progress.FailStep("JWS verification", err)
		errors.Add("verify", "JWS verification failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else if !valid {
		progress.FailStep("JWS verification", fmt.Errorf("signature invalid"))

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("JWS verified: %s", payload))

		passedSteps++
	}

	// Step 6: Create JWT.
	progress.StartStep("Creating JWT")

	claims := map[string]any{
		"sub":  "demo-user",
		"name": "JOSE Demo User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	}

	jwt, err := createJWT(ctx, client, baseURL, keyResp.KID, claims)
	if err != nil {
		progress.FailStep("JWT creation", err)
		errors.Add("jwt-create", "JWT creation failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("Created JWT (length: %d)", len(jwt)))

		passedSteps++
	}

	// Step 7: Verify JWT.
	progress.StartStep("Verifying JWT")

	jwtValid, jwtClaims, err := verifyJWT(ctx, client, baseURL, keyResp.KID, jwt)
	if err != nil {
		progress.FailStep("JWT verification", err)
		errors.Add("jwt-verify", "JWT verification failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else if !jwtValid {
		progress.FailStep("JWT verification", fmt.Errorf("JWT invalid"))

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("JWT verified, sub=%v", jwtClaims["sub"]))

		passedSteps++
	}

	// Step 8: List keys.
	progress.StartStep("Listing all keys")

	keyCount, err := listJWKs(ctx, client, baseURL)
	if err != nil {
		progress.FailStep("Key listing", err)
		errors.Add("list", "key listing failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("Found %d keys in store", keyCount))

		passedSteps++
	}

	// Step 9: Get JWKS.
	progress.StartStep("Fetching JWKS endpoint")

	jwksCount, err := getJWKS(ctx, client, baseURL)
	if err != nil {
		progress.FailStep("JWKS fetch", err)
		errors.Add("jwks", "JWKS fetch failed", err)

		if !config.ContinueOnError {
			return exitJOSEDemo(progress, errors, passedSteps, startTime)
		}
	} else {
		progress.CompleteStep(fmt.Sprintf("JWKS contains %d public keys", jwksCount))

		passedSteps++
	}

	return exitJOSEDemo(progress, errors, passedSteps, startTime)
}

// exitJOSEDemo prints summary and returns exit code.
func exitJOSEDemo(progress *ProgressDisplay, errors *ErrorAggregator, passedSteps int, startTime time.Time) int {
	result := errors.ToResult(passedSteps, joseStepCount-passedSteps-len(errors.errors))
	result.DurationMS = time.Since(startTime).Milliseconds()
	progress.PrintSummary(result)

	return result.ExitCode()
}

// jwkResponse holds the JWK generation response.
type jwkResponse struct {
	KID       string `json:"kid"`
	Algorithm string `json:"algorithm"`
	KeyType   string `json:"key_type"`
	Use       string `json:"use"`
}

// generateJWK generates a new JWK.
func generateJWK(ctx context.Context, client *http.Client, baseURL, algorithm, use string) (*jwkResponse, error) {
	reqBody, err := json.Marshal(map[string]string{
		"algorithm": algorithm,
		"use":       use,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := doJOSEPost(ctx, client, baseURL+"/jose/v1/jwk/generate", reqBody)
	if err != nil {
		return nil, fmt.Errorf("POST request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result jwkResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// signJWS signs a payload with JWS.
func signJWS(ctx context.Context, client *http.Client, baseURL, kid, payload string) (string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"kid":     kid,
		"payload": payload,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := doJOSEPost(ctx, client, baseURL+"/jose/v1/jws/sign", reqBody)
	if err != nil {
		return "", fmt.Errorf("POST request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result["jws"], nil
}

// verifyJWS verifies a JWS.
func verifyJWS(ctx context.Context, client *http.Client, baseURL, kid, jws string) (bool, string, error) {
	reqBody, err := json.Marshal(map[string]string{
		"kid": kid,
		"jws": jws,
	})
	if err != nil {
		return false, "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := doJOSEPost(ctx, client, baseURL+"/jose/v1/jws/verify", reqBody)
	if err != nil {
		return false, "", fmt.Errorf("POST request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return false, "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Valid   bool   `json:"valid"`
		Payload string `json:"payload"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", fmt.Errorf("decode response: %w", err)
	}

	return result.Valid, result.Payload, nil
}

// createJWT creates a JWT.
func createJWT(ctx context.Context, client *http.Client, baseURL, kid string, claims map[string]any) (string, error) {
	reqBody, err := json.Marshal(map[string]any{
		"kid":    kid,
		"claims": claims,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	resp, err := doJOSEPost(ctx, client, baseURL+"/jose/v1/jwt/sign", reqBody)
	if err != nil {
		return "", fmt.Errorf("POST request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result["jwt"], nil
}

// verifyJWT verifies a JWT.
func verifyJWT(ctx context.Context, client *http.Client, baseURL, kid, jwt string) (bool, map[string]any, error) {
	reqBody, err := json.Marshal(map[string]string{
		"kid": kid,
		"jwt": jwt,
	})
	if err != nil {
		return false, nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := doJOSEPost(ctx, client, baseURL+"/jose/v1/jwt/verify", reqBody)
	if err != nil {
		return false, nil, fmt.Errorf("POST request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return false, nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Valid  bool           `json:"valid"`
		Claims map[string]any `json:"claims"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Valid, result.Claims, nil
}

// listJWKs lists all JWKs.
func listJWKs(ctx context.Context, client *http.Client, baseURL string) (int, error) {
	resp, err := doJOSEGet(ctx, client, baseURL+"/jose/v1/jwk")
	if err != nil {
		return 0, fmt.Errorf("GET request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Count int `json:"count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return result.Count, nil
}

// getJWKS fetches the JWKS endpoint.
func getJWKS(ctx context.Context, client *http.Client, baseURL string) (int, error) {
	resp, err := doJOSEGet(ctx, client, baseURL+"/.well-known/jwks.json")
	if err != nil {
		return 0, fmt.Errorf("GET request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Keys []any `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	return len(result.Keys), nil
}

// doJOSEGet performs a GET request with context.
func doJOSEGet(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute GET request: %w", err)
	}

	return resp, nil
}

// doJOSEPost performs a POST request with context.
func doJOSEPost(ctx context.Context, client *http.Client, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute POST request: %w", err)
	}

	return resp, nil
}
