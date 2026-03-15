// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"context"
	http "net/http"
)

// isHTTPHealthy checks if an HTTP health endpoint returns 200 OK.
// Returns false when the server is unreachable or returns non-200 — not an error,
// just "not ready yet" for polling purposes.
func isHTTPHealthy(ctx context.Context, client *http.Client, healthURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	_ = resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// isKMSHealthy checks if a KMS admin endpoint passes both liveness and readiness checks.
// Returns false when the server is not yet ready — not an error, just "not ready yet"
// for polling purposes.
func isKMSHealthy(ctx context.Context, client *http.Client, adminBaseURL string) bool {
	return isHTTPHealthy(ctx, client, adminBaseURL+"/admin/api/v1/livez") &&
		isHTTPHealthy(ctx, client, adminBaseURL+"/admin/api/v1/readyz")
}
