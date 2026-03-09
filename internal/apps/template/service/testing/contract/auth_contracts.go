// Copyright (c) 2025 Justin Cranford
//

package contract

import (
	"fmt"
	http "net/http"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// AuthContractServer is an optional interface that services implement to opt
// into authentication contract tests. Services expose the protected endpoint
// paths so the contract suite can verify 401 rejection without knowledge of
// domain-specific routes.
//
// Services that do NOT implement this interface silently skip auth contracts.
type AuthContractServer interface {
	// ProtectedServicePaths returns GET-accessible paths behind service session
	// middleware (e.g., ["/service/api/v1/messages/rx"]).
	ProtectedServicePaths() []string

	// ProtectedBrowserPaths returns GET-accessible paths behind browser session
	// middleware (e.g., ["/browser/api/v1/messages/rx"]).
	ProtectedBrowserPaths() []string
}

// RunAuthContracts verifies that protected endpoints reject unauthenticated
// and invalid-auth requests with HTTP 401.
//
// Tests 4 contract categories (per path):
//  1. Missing Authorization header returns 401.
//  2. Invalid Bearer token returns 401.
//  3. Malformed Authorization header returns 401.
//  4. Empty Bearer token returns 401.
func RunAuthContracts(t *testing.T, publicBaseURL string, authPaths AuthContractServer) {
	t.Helper()

	client := newTLSHTTPClient(t)

	servicePaths := authPaths.ProtectedServicePaths()
	browserPaths := authPaths.ProtectedBrowserPaths()

	for _, path := range servicePaths {
		runAuthPathTests(t, client, publicBaseURL, "service", path)
	}

	for _, path := range browserPaths {
		runAuthPathTests(t, client, publicBaseURL, "browser", path)
	}
}

// runAuthPathTests runs the four auth rejection tests for a single path.
func runAuthPathTests(t *testing.T, client *http.Client, publicBaseURL, pathType, path string) {
	t.Helper()

	url := publicBaseURL + path

	tests := []struct {
		name       string
		setupAuth  func(req *http.Request)
		wantStatus int
	}{
		{
			name:       fmt.Sprintf("%s_no_auth_returns_401_%s", pathType, path),
			setupAuth:  func(_ *http.Request) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: fmt.Sprintf("%s_invalid_bearer_returns_401_%s", pathType, path),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix+"invalid-token-for-contract-test")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: fmt.Sprintf("%s_malformed_auth_returns_401_%s", pathType, path),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "NotBearer some-value")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: fmt.Sprintf("%s_empty_bearer_returns_401_%s", pathType, path),
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", cryptoutilSharedMagic.HTTPAuthorizationBearerPrefix)
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := newContractRequest(t, url)
			tc.setupAuth(req)

			resp, err := client.Do(req)
			require.NoError(t, err, "auth contract request should not fail")

			defer func() { require.NoError(t, resp.Body.Close()) }()

			require.Equal(t, tc.wantStatus, resp.StatusCode,
				"expected %d for %s %s, got %d", tc.wantStatus, pathType, path, resp.StatusCode)
		})
	}
}
