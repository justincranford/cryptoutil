package main

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
	"os/exec"
	"strings"
	"time"
)

// MockAuthZServer simulates OAuth 2.1 Authorization Server
func startAuthZServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful authorization - redirect with code
		code := generateRandomString(32)
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
		response := map[string]interface{}{
			"access_token":  generateRandomString(32),
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": generateRandomString(32),
			"id_token":      generateRandomString(64),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "authz"})
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
		if err != nil && err != http.ErrServerClosed {
			log.Printf("AuthZ Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("AuthZ Server on port %d shutting down", port)
		server.Shutdown(context.Background())
	}
}

// MockIdPServer simulates OIDC Identity Provider
func startIdPServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful authentication for any method
		response := map[string]interface{}{
			"success":    true,
			"user_id":    "test_user",
			"session_id": generateRandomString(16),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "idp"})
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
		if err != nil && err != http.ErrServerClosed {
			log.Printf("IdP Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("IdP Server on port %d shutting down", port)
		server.Shutdown(context.Background())
	}
}

// MockResourceServer simulates protected API
func startResourceServer(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/protected", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		response := map[string]interface{}{
			"data":    "protected resource",
			"user_id": "test_user",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "resource"})
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
		if err != nil && err != http.ErrServerClosed {
			log.Printf("Resource Server on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("Resource Server on port %d shutting down", port)
		server.Shutdown(context.Background())
	}
}

// MockSPARP simulates SPA Relying Party
func startSPARP(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		response := map[string]interface{}{
			"code":        code,
			"state":       state,
			"received_at": time.Now().Format(time.RFC3339),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "spa-rp"})
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
		if err != nil && err != http.ErrServerClosed {
			log.Printf("SPA RP on port %d failed: %v", port, err)
		}
	case <-ctx.Done():
		log.Printf("SPA RP on port %d shutting down", port)
		server.Shutdown(context.Background())
	}
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func main() {
	log.Println("Starting mock identity services...")

	// Start all mock services in goroutines
	go startAuthZServer(context.Background(), 8080)
	go startIdPServer(context.Background(), 8081)
	go startResourceServer(context.Background(), 8082)
	go startSPARP(context.Background(), 8083)

	// Wait for services to be ready
	log.Println("Waiting for services to be ready...")
	time.Sleep(3 * time.Second) // Give services time to start

	// Health check all services
	if !waitForService("https://127.0.0.1:8080/health", "AuthZ") ||
		!waitForService("https://127.0.0.1:8081/health", "IdP") ||
		!waitForService("https://127.0.0.1:8082/health", "Resource") ||
		!waitForService("https://127.0.0.1:8083/health", "SPA RP") {
		log.Fatal("One or more services failed to start")
	}

	log.Println("All mock identity services started successfully")

	// Run the tests
	log.Println("Running E2E tests...")
	runTests()
}

func waitForService(url, serviceName string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for i := 0; i < 10; i++ {
		resp, err := client.Get(url)
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

func runTests() {
	// Run go test command
	cmd := exec.Command("go", "test", "-cover", "-coverprofile=coverage.out", "./internal/identity/test/e2e/...", "-timeout=60s", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Printf("Tests failed: %v", err)
		os.Exit(1)
	}

	log.Println("Tests completed successfully")
}
