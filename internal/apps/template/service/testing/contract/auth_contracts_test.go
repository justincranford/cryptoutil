// Copyright (c) 2025 Justin Cranford
//
// Tests for RunAuthContracts using a mock TLS server with auth-protected routes.
package contract

import (
	"crypto/tls"
	http "net/http"
	httptest "net/http/httptest"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// mockAuthContractServer implements AuthContractServer for testing.
type mockAuthContractServer struct {
	servicePaths []string
	browserPaths []string
}

func (m *mockAuthContractServer) ProtectedServicePaths() []string { return m.servicePaths }
func (m *mockAuthContractServer) ProtectedBrowserPaths() []string { return m.browserPaths }

// newAuthTestServer creates a TLS test server that returns 401 for protected
// paths when auth is missing or invalid, and 200 for unprotected paths.
func newAuthTestServer(t *testing.T, protectedPaths []string) *httptest.Server {
	t.Helper()

	protected := make(map[string]bool, len(protectedPaths))
	for _, p := range protectedPaths {
		protected[p] = true
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !protected[r.URL.Path] {
			w.WriteHeader(http.StatusOK)

			return
		}

		authHeader := r.Header.Get("Authorization")

		switch {
		case authHeader == "":
			w.WriteHeader(http.StatusUnauthorized)
		case !strings.HasPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix):
			w.WriteHeader(http.StatusUnauthorized)
		case strings.TrimPrefix(authHeader, cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix) == "":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			// Any non-empty Bearer token returns 401 (invalid session).
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	return server
}

func TestRunAuthContracts(t *testing.T) {
	t.Parallel()

	servicePaths := []string{"/service/api/v1/protected"}
	browserPaths := []string{"/browser/api/v1/protected"}
	allPaths := append(servicePaths, browserPaths...)

	server := newAuthTestServer(t, allPaths)

	authPaths := &mockAuthContractServer{
		servicePaths: servicePaths,
		browserPaths: browserPaths,
	}

	RunAuthContracts(t, server.URL, authPaths)
}

func TestRunAuthContracts_EmptyPaths(t *testing.T) {
	t.Parallel()

	server := newAuthTestServer(t, nil)

	authPaths := &mockAuthContractServer{
		servicePaths: nil,
		browserPaths: nil,
	}

	// Should complete without error when no paths are provided.
	RunAuthContracts(t, server.URL, authPaths)
}

func TestRunAuthContracts_MultiplePaths(t *testing.T) {
	t.Parallel()

	servicePaths := []string{"/service/api/v1/messages/rx", "/service/api/v1/keys"}
	browserPaths := []string{"/browser/api/v1/messages/rx", "/browser/api/v1/keys"}
	allPaths := append(servicePaths, browserPaths...)

	server := newAuthTestServer(t, allPaths)

	authPaths := &mockAuthContractServer{
		servicePaths: servicePaths,
		browserPaths: browserPaths,
	}

	RunAuthContracts(t, server.URL, authPaths)
}

func TestRunAuthContracts_DirectHTTPValidation(t *testing.T) {
	t.Parallel()

	protectedPath := "/service/api/v1/protected"
	server := newAuthTestServer(t, []string{protectedPath})

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test-only: mock TLS server
			DisableKeepAlives: true,
		},
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"no_auth_header", "", http.StatusUnauthorized},
		{"invalid_bearer", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix + "bad-token", http.StatusUnauthorized},
		{"malformed_auth", "Basic dXNlcjpwYXNz", http.StatusUnauthorized},
		{"empty_bearer", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix, http.StatusUnauthorized},
		{"unprotected_path_ok", "", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url := server.URL + protectedPath
			if tc.name == "unprotected_path_ok" {
				url = server.URL + "/unprotected"
			}

			req, err := http.NewRequest(http.MethodGet, url, nil) //nolint:noctx // test-only: no context needed
			require.NoError(t, err, "failed to create request")

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			resp, err := client.Do(req)
			require.NoError(t, err, "request failed")

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tc.wantStatus, resp.StatusCode, "expected status %d, got %d", tc.wantStatus, resp.StatusCode)
		})
	}
}
