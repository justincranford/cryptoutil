package go_update_direct_dependencies

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockGitHubServer creates a test HTTP server that simulates GitHub API responses.
// It handles release and tag queries with configurable responses for testing.
type MockGitHubServer struct {
	server        *httptest.Server
	releaseData   map[string]string // owner/repo -> release JSON
	tagData       map[string]string // owner/repo -> tags JSON array
	errorResponse map[string]int    // owner/repo -> HTTP status code for errors
}

// NewMockGitHubServer creates a new mock GitHub API server for testing.
func NewMockGitHubServer() *MockGitHubServer {
	mock := &MockGitHubServer{
		releaseData:   make(map[string]string),
		tagData:       make(map[string]string),
		errorResponse: make(map[string]int),
	}

	// Create HTTP test server with handler.
	mock.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.handleRequest(w, r)
	}))

	return mock
}

// URL returns the base URL of the mock server.
func (m *MockGitHubServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server.
func (m *MockGitHubServer) Close() {
	m.server.Close()
}

// SetReleaseData configures the mock to return specific release data for an owner/repo.
// Format: owner/repo -> release JSON string.
func (m *MockGitHubServer) SetReleaseData(ownerRepo, jsonData string) {
	m.releaseData[ownerRepo] = jsonData
}

// SetTagData configures the mock to return specific tag data for an owner/repo.
// Format: owner/repo -> tags JSON array string.
func (m *MockGitHubServer) SetTagData(ownerRepo, jsonData string) {
	m.tagData[ownerRepo] = jsonData
}

// SetErrorResponse configures the mock to return an error status code for an owner/repo.
// Common codes: 403 (rate limit), 404 (not found), 500 (server error).
func (m *MockGitHubServer) SetErrorResponse(ownerRepo string, statusCode int) {
	m.errorResponse[ownerRepo] = statusCode
}

// handleRequest processes incoming HTTP requests and returns mock responses.
func (m *MockGitHubServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract owner/repo from path.
	// Expected paths:
	// - /repos/{owner}/{repo}/releases/latest
	// - /repos/{owner}/{repo}/tags
	parts := strings.Split(strings.TrimPrefix(path, "/repos/"), "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid path", http.StatusBadRequest)

		return
	}

	ownerRepo := fmt.Sprintf("%s/%s", parts[0], parts[1])

	// Check for error response configuration.
	if statusCode, ok := m.errorResponse[ownerRepo]; ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		// Return appropriate error JSON based on status code.
		switch statusCode {
		case http.StatusForbidden:
			_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
		case http.StatusNotFound:
			_, _ = w.Write([]byte(`{"message": "Not Found"}`))
		default:
			_, _ = w.Write([]byte(`{"message": "Server error"}`))
		}

		return
	}

	// Handle release endpoint.
	if strings.HasSuffix(path, "/releases/latest") {
		if data, ok := m.releaseData[ownerRepo]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(data))

			return
		}

		http.Error(w, "Release not found", http.StatusNotFound)

		return
	}

	// Handle tags endpoint.
	if strings.HasSuffix(path, "/tags") {
		if data, ok := m.tagData[ownerRepo]; ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(data))

			return
		}

		http.Error(w, "Tags not found", http.StatusNotFound)

		return
	}

	http.Error(w, "Unknown endpoint", http.StatusNotFound)
}

// SampleReleaseJSON returns sample release JSON for testing.
func SampleReleaseJSON(tagName string) string {
	return fmt.Sprintf(`{
		"tag_name": "%s",
		"name": "Release %s",
		"published_at": "2024-01-15T10:00:00Z"
	}`, tagName, tagName)
}

// SampleTagsJSON returns sample tags JSON array for testing.
func SampleTagsJSON(tags ...string) string {
	var tagObjects []string
	for _, tag := range tags {
		tagObjects = append(tagObjects, fmt.Sprintf(`{"name": "%s"}`, tag))
	}

	return "[" + strings.Join(tagObjects, ",") + "]"
}
