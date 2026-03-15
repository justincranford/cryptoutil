//go:build e2e

// Copyright (c) 2025 Justin Cranford
//
//

package e2e

import (
	"context"
	json "encoding/json"
	"fmt"
	"log"
	http "net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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
		authZPort:  cryptoutilSharedMagic.TestAuthZServerPort,
		idpPort:    cryptoutilSharedMagic.TestIDPServerPort,
		rsPort:     cryptoutilSharedMagic.TestResourceServerPort,
		spaRPPort:  cryptoutilSharedMagic.TestSPARPServerPort,
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
		redirectURI := r.URL.Query().Get(cryptoutilSharedMagic.ParamRedirectURI)
		if redirectURI == "" {
			http.Error(w, "redirect_uri is required", http.StatusBadRequest)

			return
		}

		// Generate authorization code and get state
		code := generateRandomString(cryptoutilSharedMagic.TestRandomStringLength16)
		state := r.URL.Query().Get(cryptoutilSharedMagic.ParamState)

		// Build redirect URL with authorization code and state
		redirectURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)

		// Perform HTTP redirect
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}))

	mux.HandleFunc("/oauth2/v1/token", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token exchange
		response := map[string]any{
			cryptoutilSharedMagic.TokenTypeAccessToken:  generateRandomString(cryptoutilSharedMagic.TestRandomStringLength32),
			cryptoutilSharedMagic.ParamTokenType:        cryptoutilSharedMagic.AuthorizationBearer,
			cryptoutilSharedMagic.ParamExpiresIn:        tokenExpirySeconds,
			cryptoutilSharedMagic.ParamIDToken:          generateRandomString(cryptoutilSharedMagic.TestRandomStringLength64),
			cryptoutilSharedMagic.GrantTypeRefreshToken: generateRandomString(cryptoutilSharedMagic.TestRandomStringLength32),
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
			"active":                             true,
			cryptoutilSharedMagic.ClaimClientID:  "spa-client",
			cryptoutilSharedMagic.ParamTokenType: cryptoutilSharedMagic.AuthorizationBearer,
			cryptoutilSharedMagic.ClaimExp:       time.Now().UTC().Add(time.Hour).Unix(),
			cryptoutilSharedMagic.ClaimIat:       time.Now().UTC().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Printf("Failed to encode introspect response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc(cryptoutilSharedMagic.PathAuthorize, corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate OAuth 2.1 authorization code flow
		response := map[string]any{
			cryptoutilSharedMagic.ResponseTypeCode: generateRandomString(cryptoutilSharedMagic.TestRandomStringLength16),
			cryptoutilSharedMagic.ParamState:       r.URL.Query().Get(cryptoutilSharedMagic.ParamState),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode authorize response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	}))

	mux.HandleFunc(cryptoutilSharedMagic.PathToken, corsHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token exchange
		response := map[string]any{
			cryptoutilSharedMagic.TokenTypeAccessToken:  generateRandomString(cryptoutilSharedMagic.TestRandomStringLength32),
			cryptoutilSharedMagic.ParamTokenType:        cryptoutilSharedMagic.AuthorizationBearer,
			cryptoutilSharedMagic.ParamExpiresIn:        tokenExpirySeconds,
			cryptoutilSharedMagic.ParamIDToken:          generateRandomString(cryptoutilSharedMagic.TestRandomStringLength64),
			cryptoutilSharedMagic.GrantTypeRefreshToken: generateRandomString(cryptoutilSharedMagic.TestRandomStringLength32),
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

		if err := json.NewEncoder(w).Encode(map[string]string{cryptoutilSharedMagic.StringStatus: "ok", "service": cryptoutilSharedMagic.AuthzServiceName}); err != nil {
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
