// Copyright (c) 2025 Justin Cranford
//

// Package assertions provides reusable HTTP response assertion helpers for cryptoutil service tests.
// Each helper reads and closes the response body; do not call Body.Close() separately.
package assertions

import (
	json "encoding/json"
	"io"
	http "net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// AssertHealthy asserts a response is 200 OK with body {"status":"healthy"}.
// The response body is read and closed by this function.
func AssertHealthy(t testing.TB, resp *http.Response) {
	t.Helper()

	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var result map[string]string

	require.NoError(t, json.Unmarshal(body, &result))
	assert.Equal(t, cryptoutilSharedMagic.DockerServiceHealthHealthy, result[cryptoutilSharedMagic.StringStatus], "health status should be healthy")
}

// AssertErrorResponse asserts a response has the expected HTTP status code and a valid JSON error body.
// The response body is read and closed by this function.
func AssertErrorResponse(t testing.TB, resp *http.Response, code int) {
	t.Helper()

	require.NotNil(t, resp)
	assert.Equal(t, code, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var result map[string]any

	require.NoError(t, json.Unmarshal(body, &result), "error response body must be valid JSON")
}

// AssertTraceID asserts the response includes a non-empty X-Request-Id header.
func AssertTraceID(t testing.TB, resp *http.Response) {
	t.Helper()

	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Header.Get("X-Request-Id"), "X-Request-Id header should be present")
}

// AssertJSONContentType asserts the response Content-Type contains "application/json".
func AssertJSONContentType(t testing.TB, resp *http.Response) {
	t.Helper()

	require.NotNil(t, resp)
	contentType := resp.Header.Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "application/json"), "Content-Type should contain application/json, got: %s", contentType)
}
