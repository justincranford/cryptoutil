//go:build e2e

// Copyright (c) 2025 Justin Cranford
//
//

package e2e

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

const (
	mockCertFile           = "mock_cert.pem"
	mockKeyFile            = "mock_key.pem"
	httpMethodOptions      = "OPTIONS"
	tokenExpirySeconds     = 3600
	shutdownTimeout        = 10 * time.Second
	serviceReadyMaxRetries = 30
	serviceReadyRetryDelay = 1 * time.Second
	alphabetSize           = 26 // Number of letters in English alphabet (A-Z)
)

// TestableMockServices provides mock identity services that can be started and stopped for testing.
type TestableMockServices struct {
	authZServer *http.Server
	idpServer   *http.Server
	rsServer    *http.Server
	spaRPServer *http.Server
	authZPort   int
	idpPort     int
	rsPort      int
	spaRPPort   int
	wg          sync.WaitGroup
	shutdownCh  chan struct{}
}

// NewTestableMockServices creates a new testable mock services instance.
func NewTestableMockServices() *TestableMockServices {
	return &TestableMockServices{
		authZPort:  cryptoutilMagic.TestAuthZServerPort,
		idpPort:    cryptoutilMagic.TestIDPServerPort,
		rsPort:     cryptoutilMagic.TestResourceServerPort,
		spaRPPort:  cryptoutilMagic.TestSPARPServerPort,
		shutdownCh: make(chan struct{}),
	}
}

// Start starts all mock services and waits for them to be ready.
func (tms *TestableMockServices) Start(ctx context.Context) error {
	log.Println("Starting testable mock identity services...")

	// Start all servers
	if err := tms.startAuthZServer(ctx); err != nil {
		return fmt.Errorf("failed to start AuthZ server: %w", err)
	}

	if err := tms.startIDPServer(ctx); err != nil {
		return fmt.Errorf("failed to start IdP server: %w", err)
	}

	if err := tms.startResourceServer(ctx); err != nil {
		return fmt.Errorf("failed to start resource server: %w", err)
	}

	if err := tms.startSPARPServer(ctx); err != nil {
		return fmt.Errorf("failed to start SPA RP server: %w", err)
	}

	// Wait for services to be ready
	if err := tms.waitForServicesReady(ctx); err != nil {
		tms.Stop(ctx)

		return fmt.Errorf("services failed to become ready: %w", err)
	}

	log.Println("All testable mock identity services started successfully")
	log.Printf("Services are running on:")
	log.Printf("  AuthZ Server: https://127.0.0.1:%d", tms.authZPort)
	log.Printf("  IdP Server:   https://127.0.0.1:%d", tms.idpPort)
	log.Printf("  Resource:     https://127.0.0.1:%d", tms.rsPort)
	log.Printf("  SPA RP:       https://127.0.0.1:%d", tms.spaRPPort)

	return nil
}

// Stop stops all mock services gracefully.
func (tms *TestableMockServices) Stop(ctx context.Context) {
	log.Println("Stopping testable mock identity services...")

	// Signal shutdown
	close(tms.shutdownCh)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// Shutdown servers
	if tms.authZServer != nil {
		if err := tms.authZServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("AuthZ server shutdown error: %v", err)
		}
	}

	if tms.idpServer != nil {
		if err := tms.idpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("IdP server shutdown error: %v", err)
		}
	}

	if tms.rsServer != nil {
		if err := tms.rsServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Resource server shutdown error: %v", err)
		}
	}

	if tms.spaRPServer != nil {
		if err := tms.spaRPServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("SPA RP server shutdown error: %v", err)
		}
	}

	// Wait for all goroutines to finish
	done := make(chan struct{})

	go func() {
		tms.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All testable mock identity services stopped")
	case <-shutdownCtx.Done():
		log.Println("Timeout waiting for services to stop")
	}
}

func (tms *TestableMockServices) getCertPaths() (string, string) {
	// Look for cert files in the project root directory
	// This works for both development and CI environments
	projectRoot := "c:\\Dev\\Projects\\cryptoutil" // Absolute path for Windows development
	certFile := filepath.Join(projectRoot, mockCertFile)
	keyFile := filepath.Join(projectRoot, mockKeyFile)

	// Check if files exist in project root
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		// Fallback: try current working directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Printf("Failed to get current working directory: %v", err)

			return mockCertFile, mockKeyFile
		}

		certFile = filepath.Join(cwd, mockCertFile)
		keyFile = filepath.Join(cwd, mockKeyFile)

		// If still not found, use relative paths as final fallback
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			log.Printf("Certificate files not found in project root or CWD, using relative paths")

			return mockCertFile, mockKeyFile
		}
	}

	return certFile, keyFile
}

func (tms *TestableMockServices) startAuthZServer(ctx context.Context) error {
	mux := http.NewServeMux()

	// CORS middleware
	corsHandler := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8446")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == httpMethodOptions {
				w.WriteHeader(http.StatusOK)

				return
			}

			handler(w, r)
		}
	}

	mux.HandleFunc("/oauth2/v1/authorize", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate OAuth 2.1 authorization code flow
		// Parse the redirect_uri from query parameters
		redirectURI := r.URL.Query().Get("redirect_uri")
		if redirectURI == "" {
			http.Error(w, "redirect_uri is required", http.StatusBadRequest)

			return
		}

		// Generate authorization code and get state
		code := generateRandomString(cryptoutilMagic.TestRandomStringLength16)
		state := r.URL.Query().Get("state")

		// Build redirect URL with authorization code and state
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)

		// Perform HTTP redirect
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}))

	mux.HandleFunc("/oauth2/v1/token", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token exchange
		response := map[string]any{
			"access_token":  generateRandomString(cryptoutilMagic.TestRandomStringLength32),
			"token_type":    "Bearer",
			"expires_in":    tokenExpirySeconds,
			"id_token":      generateRandomString(cryptoutilMagic.TestRandomStringLength64),
			"refresh_token": generateRandomString(cryptoutilMagic.TestRandomStringLength32),
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Failed to encode token response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/oauth2/v1/introspect", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token introspection
		response := map[string]any{
			"active":     true,
			"client_id":  "spa-client",
			"token_type": "Bearer",
			"exp":        time.Now().UTC().Add(time.Hour).Unix(),
			"iat":        time.Now().UTC().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Failed to encode introspect response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/authorize", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate OAuth 2.1 authorization code flow
		response := map[string]any{
			"code":  generateRandomString(cryptoutilMagic.TestRandomStringLength16),
			"state": r.URL.Query().Get("state"),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode authorize response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/token", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token exchange
		response := map[string]any{
			"access_token":  generateRandomString(cryptoutilMagic.TestRandomStringLength32),
			"token_type":    "Bearer",
			"expires_in":    tokenExpirySeconds,
			"id_token":      generateRandomString(cryptoutilMagic.TestRandomStringLength64),
			"refresh_token": generateRandomString(cryptoutilMagic.TestRandomStringLength32),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode token response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "authz"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	tms.authZServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", tms.authZPort),
		Handler: mux,
	}

	tms.wg.Add(1)

	go func() {
		defer tms.wg.Done()

		certFile, keyFile := tms.getCertPaths()

		if err := tms.authZServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Printf("AuthZ server error: %v", err)
		}
	}()

	return nil
}

func (tms *TestableMockServices) startIDPServer(ctx context.Context) error {
	mux := http.NewServeMux()

	// CORS middleware
	corsHandler := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:8446")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)

				return
			}

			handler(w, r)
		}
	}

	mux.HandleFunc("/oidc/v1/userinfo", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate OIDC UserInfo endpoint
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		response := map[string]any{
			"sub":            "test_user",
			"name":           "Test User",
			"email":          "test@example.com",
			"email_verified": true,
			"profile":        "https://example.com/profile/test_user",
			"picture":        "https://example.com/avatar/test_user.jpg",
			"updated_at":     time.Now().UTC().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode userinfo response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/login", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful authentication for any method
		response := map[string]any{
			"success":    true,
			"user_id":    "test_user",
			"session_id": generateRandomString(cryptoutilMagic.TestRandomStringLength16),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode login response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "idp"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	tms.idpServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", tms.idpPort),
		Handler: mux,
	}

	tms.wg.Add(1)

	go func() {
		defer tms.wg.Done()

		certFile, keyFile := tms.getCertPaths()

		if err := tms.idpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Printf("IdP server error: %v", err)
		}
	}()

	return nil
}

func (tms *TestableMockServices) startResourceServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/protected", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		response := map[string]any{
			"data":    "protected resource",
			"user_id": "test_user",
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode protected resource response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "resource"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	tms.rsServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", tms.rsPort),
		Handler: mux,
	}

	tms.wg.Add(1)

	go func() {
		defer tms.wg.Done()

		certFile, keyFile := tms.getCertPaths()

		if err := tms.rsServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Printf("Resource server error: %v", err)
		}
	}()

	return nil
}

func (tms *TestableMockServices) startSPARPServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		response := map[string]any{
			"code":        code,
			"state":       state,
			"received_at": time.Now().UTC().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode callback response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "spa-rp"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	tms.spaRPServer = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", tms.spaRPPort),
		Handler: mux,
	}

	tms.wg.Add(1)

	go func() {
		defer tms.wg.Done()

		certFile, keyFile := tms.getCertPaths()

		if err := tms.spaRPServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			log.Printf("SPA RP server error: %v", err)
		}
	}()

	return nil
}

func (tms *TestableMockServices) waitForServicesReady(ctx context.Context) error {
	client := &http.Client{
		Timeout: healthCheckTimeoutService,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	endpoints := []struct {
		url      string
		expected string
	}{
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.authZPort), "authz"},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.idpPort), "idp"},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.rsPort), "resource"},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.spaRPPort), "spa-rp"},
	}

	maxRetries := 30 // 30 seconds max wait
	for i := 0; i < maxRetries; i++ {
		allReady := true

		for _, ep := range endpoints {
			req, err := http.NewRequestWithContext(ctx, "GET", ep.url, nil)
			if err != nil {
				return fmt.Errorf("failed to create health check request for %s: %w", ep.url, err)
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Health check attempt %d failed for %s: %v", i+1, ep.url, err)

				allReady = false

				break
			}

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			var result map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				log.Printf("Health check attempt %d failed for %s: invalid JSON", i+1, ep.url)

				allReady = false

				break
			}

			if result["status"] != "ok" || result["service"] != ep.expected {
				log.Printf("Health check attempt %d failed for %s: unexpected response %+v", i+1, ep.url, result)

				allReady = false

				break
			}
		}

		if allReady {
			log.Printf("All services ready after %d attempts", i+1)

			return nil
		}

		// Wait before retrying
		select {
		case <-time.After(serviceReadyRetryDelay):
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for services: %w", ctx.Err())
		}
	}

	return fmt.Errorf("services failed to become ready after %d attempts", maxRetries)
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate random string: %v", err)
		// Fallback to a simple pattern for testing.
		for i := range bytes {
			bytes[i] = 'A' + byte(i%alphabetSize) // A-Z pattern
		}
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
