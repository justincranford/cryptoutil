//go:build e2e

// Copyright (c) 2025 Justin Cranford
//
//

package e2e

import (
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/base64"
	json "encoding/json"
	"fmt"
	"log"
	http "net/http"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

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
		if !strings.HasPrefix(auth, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		response := map[string]any{
			cryptoutilSharedMagic.ClaimSub:            "test_user",
			cryptoutilSharedMagic.ClaimName:           "Test User",
			cryptoutilSharedMagic.ClaimEmail:          "test@example.com",
			cryptoutilSharedMagic.ClaimEmailVerified: true,
			cryptoutilSharedMagic.ClaimProfile:        "https://example.com/profile/test_user",
			cryptoutilSharedMagic.ClaimPicture:        "https://example.com/avatar/test_user.jpg",
			cryptoutilSharedMagic.ClaimUpdatedAt:     time.Now().UTC().Unix(),
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
			"session_id": generateRandomString(cryptoutilSharedMagic.TestRandomStringLength16),
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

		if err := json.NewEncoder(w).Encode(map[string]string{cryptoutilSharedMagic.StringStatus: "ok", "service": cryptoutilSharedMagic.IDPServiceName}); err != nil {
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
		if !strings.HasPrefix(auth, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) {
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

		if err := json.NewEncoder(w).Encode(map[string]string{cryptoutilSharedMagic.StringStatus: "ok", "service": "resource"}); err != nil {
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
		code := r.URL.Query().Get(cryptoutilSharedMagic.ResponseTypeCode)
		state := r.URL.Query().Get(cryptoutilSharedMagic.ParamState)

		response := map[string]any{
			cryptoutilSharedMagic.ResponseTypeCode:        code,
			cryptoutilSharedMagic.ParamState:       state,
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

		if err := json.NewEncoder(w).Encode(map[string]string{cryptoutilSharedMagic.StringStatus: "ok", "service": "spa-rp"}); err != nil {
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
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // G402: E2E testing with self-signed certs
		},
	}

	endpoints := []struct {
		url      string
		expected string
	}{
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.authZPort), cryptoutilSharedMagic.AuthzServiceName},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.idpPort), cryptoutilSharedMagic.IDPServiceName},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.rsPort), "resource"},
		{fmt.Sprintf("https://127.0.0.1:%d/health", tms.spaRPPort), "spa-rp"},
	}

	maxRetries := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days // 30 seconds max wait
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

			if result[cryptoutilSharedMagic.StringStatus] != "ok" || result["service"] != ep.expected {
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
	if _, err := crand.Read(bytes); err != nil {
		log.Printf("Failed to generate random string: %v", err)
		// Fallback to a simple pattern for testing.
		for i := range bytes {
			bytes[i] = 'A' + byte(i%alphabetSize) // A-Z pattern
		}
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
