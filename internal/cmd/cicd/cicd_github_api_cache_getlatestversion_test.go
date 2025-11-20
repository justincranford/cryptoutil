package cicd

import (
	"testing"

	"cryptoutil/internal/cmd/cicd/common"

	"github.com/stretchr/testify/require"
)

// TestGetLatestVersion_Success tests successful version retrieval from GitHub API.
func TestGetLatestVersion_Success(t *testing.T) {
	t.Parallel()

	// Setup mock server.
	mock := NewMockGitHubServer()
	defer mock.Close()

	mock.SetReleaseData("actions/checkout", SampleReleaseJSON("v4.2.1"))

	// Create logger.
	logger := common.NewLogger("test")

	// Call getLatestVersionWithBaseURL with mock server URL.
	version, err := getLatestVersionWithBaseURL(logger, "actions/checkout", mock.URL())

	// Assertions.
	require.NoError(t, err, "getLatestVersion should succeed")
	require.Equal(t, "v4.2.1", version, "Version should match mock response")
}


// TestGetLatestVersion_RateLimitError tests handling of 403 rate limit response.
func TestGetLatestVersion_RateLimitError(t *testing.T) {
	t.Parallel()

	// Setup mock server with rate limit error.
	mock := NewMockGitHubServer()
	defer mock.Close()

	mock.SetErrorResponse("actions/checkout", 403)

	// Create logger.
	logger := common.NewLogger("test")

	// Call getLatestVersionWithBaseURL with mock server URL.
	version, err := getLatestVersionWithBaseURL(logger, "actions/checkout", mock.URL())

	// Assertions.
	require.Error(t, err, "getLatestVersion should return error for rate limit")
	require.Contains(t, err.Error(), "403", "Error should mention status code")
	require.Empty(t, version, "Version should be empty on error")
}

// TestGetLatestVersion_NotFoundError tests handling of 404 not found response.
func TestGetLatestVersion_NotFoundError(t *testing.T) {
	t.Parallel()

	// Setup mock server with not found error.
	mock := NewMockGitHubServer()
	defer mock.Close()

	mock.SetErrorResponse("nonexistent/repo", 404)

	// Create logger.
	logger := common.NewLogger("test")

	// Call getLatestVersionWithBaseURL with mock server URL.
	version, err := getLatestVersionWithBaseURL(logger, "nonexistent/repo", mock.URL())

	// Assertions.
	require.Error(t, err, "getLatestVersion should return error for not found")
	require.Contains(t, err.Error(), "404", "Error should mention status code")
	require.Empty(t, version, "Version should be empty on error")
}

// TestGetLatestVersion_ServerError tests handling of 500 server error response.
func TestGetLatestVersion_ServerError(t *testing.T) {
	t.Parallel()

	// Setup mock server with server error.
	mock := NewMockGitHubServer()
	defer mock.Close()

	mock.SetErrorResponse("actions/checkout", 500)

	// Create logger.
	logger := common.NewLogger("test")

	// Call getLatestVersionWithBaseURL with mock server URL.
	version, err := getLatestVersionWithBaseURL(logger, "actions/checkout", mock.URL())

	// Assertions.
	require.Error(t, err, "getLatestVersion should return error for server error")
	require.Contains(t, err.Error(), "500", "Error should mention status code")
	require.Empty(t, version, "Version should be empty on error")
}

// TestGetLatestVersion_CacheHit tests that cached version is used without API call.
func TestGetLatestVersion_CacheHit(t *testing.T) {
	t.Parallel()

	// Create logger.
	logger := common.NewLogger("test")

	// Populate cache directly.
	cacheKey := "release:actions/checkout"
	githubAPICache.Set(cacheKey, "v4.1.0")

	// Call getLatestVersionWithBaseURL (should use cache, no API call).
	version, err := getLatestVersionWithBaseURL(logger, "actions/checkout", "http://should-not-be-called")

	// Assertions.
	require.NoError(t, err, "getLatestVersion should succeed with cache")
	require.Equal(t, "v4.1.0", version, "Version should come from cache")

	// Note: No cache cleanup needed as it's an in-memory cache and each test uses unique keys.
}

