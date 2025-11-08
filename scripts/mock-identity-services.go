package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

const (
	alphabetSize = 26 // Number of letters in English alphabet
)

// MockAuthZServer simulates OAuth 2.1 Authorization Server.
func startAuthZServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful authorization - redirect with code
		code := generateRandomString(cryptoutilMagic.TestRandomStringLength32)
		redirectURI := r.URL.Query().Get("redirect_uri")
		state := r.URL.Query().Get("state")

		if redirectURI == "" {
			http.Error(w, "missing redirect_uri", http.StatusBadRequest)

			return
		}

		location := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, code, state)
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusFound)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// Simulate token exchange
		response := map[string]any{
			"access_token":  generateRandomString(cryptoutilMagic.TestRandomStringLength32),
			"token_type":    "Bearer",
			"expires_in":    cryptoutilMagic.TestTokenExpirationSeconds,
			"refresh_token": generateRandomString(cryptoutilMagic.TestRandomStringLength32),
			"id_token":      generateRandomString(cryptoutilMagic.TestRandomStringLength64),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode token response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "authz"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	server := &http.Server{
		Addr:    "127.0.0.1:" + fmt.Sprintf("%d", port),
		Handler: mux,
	}

	log.Printf("AuthZ Server starting on port %d", port)

	// Channel to signal when server is done
	done := make(chan error, 1)
	go func() {
		done <- server.ListenAndServeTLS("mock_cert.pem", "mock_key.pem")
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("AuthZ Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("AuthZ Server on port %d shutting down", port)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("AuthZ Server shutdown error: %v", err)
		}
	}
}

// MockIdPServer simulates OIDC Identity Provider.
func startIDPServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "idp"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)

			return
		}
	})

	server := &http.Server{
		Addr:    "127.0.0.1:" + fmt.Sprintf("%d", port),
		Handler: mux,
	}

	log.Printf("IdP Server starting on port %d", port)

	// Channel to signal when server is done
	done := make(chan error, 1)
	go func() {
		done <- server.ListenAndServeTLS("mock_cert.pem", "mock_key.pem")
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("IdP Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("IdP Server on port %d shutting down", port)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("IdP Server shutdown error: %v", err)
		}
	}
}

// MockResourceServer simulates protected API.
func startResourceServer(ctx context.Context, port int) {
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

	server := &http.Server{
		Addr:    "127.0.0.1:" + fmt.Sprintf("%d", port),
		Handler: mux,
	}

	log.Printf("Resource Server starting on port %d", port)

	// Channel to signal when server is done
	done := make(chan error, 1)
	go func() {
		done <- server.ListenAndServeTLS("mock_cert.pem", "mock_key.pem")
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Resource Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("Resource Server on port %d shutting down", port)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("Resource Server shutdown error: %v", err)
		}
	}
}

// MockSPARP simulates SPA Relying Party.
func startSPARP(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		response := map[string]any{
			"code":        code,
			"state":       state,
			"received_at": time.Now().Format(time.RFC3339),
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

	server := &http.Server{
		Addr:    "127.0.0.1:" + fmt.Sprintf("%d", port),
		Handler: mux,
	}

	log.Printf("SPA RP starting on port %d", port)

	// Channel to signal when server is done
	done := make(chan error, 1)
	go func() {
		done <- server.ListenAndServeTLS("mock_cert.pem", "mock_key.pem")
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("SPA RP on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("SPA RP on port %d shutting down", port)

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("SPA RP shutdown error: %v", err)
		}
	}
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Failed to generate random string: %v", err)
		// Fallback to a simple pattern for testing
		for i := range bytes {
			bytes[i] = 'A' + byte(i%alphabetSize) // A-Z pattern
		}
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func main() {
	log.Println("Starting mock identity services...")

	// Start all mock services in goroutines
	go startAuthZServer(context.Background(), cryptoutilMagic.TestAuthZServerPort)
	go startIDPServer(context.Background(), cryptoutilMagic.TestIDPServerPort)
	go startResourceServer(context.Background(), cryptoutilMagic.TestResourceServerPort)
	go startSPARP(context.Background(), cryptoutilMagic.TestSPARPServerPort)

	// Wait for services to be ready
	log.Println("Waiting for services to be ready...")
	time.Sleep(cryptoutilMagic.TestServiceStartupDelaySeconds * time.Second) // Give services time to start

	// Health check all services
	if !waitForService("https://127.0.0.1:8080/health", "AuthZ") ||
		!waitForService("https://127.0.0.1:8081/health", "IdP") ||
		!waitForService("https://127.0.0.1:8082/health", "Resource") ||
		!waitForService("https://127.0.0.1:8083/health", "SPA RP") {
		log.Fatal("One or more services failed to start")
	}

	log.Println("All mock identity services started successfully")

	// Keep services running indefinitely for UI access
	log.Println("Services are running. Press Ctrl+C to stop.")
	select {} // Block forever
}

func waitForService(url, serviceName string) bool {
	client := &http.Client{
		Timeout: cryptoutilMagic.TestHTTPHealthTimeoutSeconds * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for i := 0; i < 10; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			time.Sleep(1 * time.Second)

			continue
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				log.Printf("%s service is ready", serviceName)

				return true
			}
		}

		log.Printf("Waiting for %s service... (%d/10)", serviceName, i+1)
		time.Sleep(1 * time.Second)
	}

	log.Printf("%s service failed to start", serviceName)

	return false
}
