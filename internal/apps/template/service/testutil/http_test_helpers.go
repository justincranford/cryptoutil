// Copyright (c) 2025 Justin Cranford

package testutil

import (
"crypto/tls"
"net/http"
"net/http/httptest"
"time"
)

// NewInsecureTLSClient creates an HTTP client that accepts self-signed certificates.
// Used for testing HTTPS endpoints with self-signed TLS certificates.
func NewInsecureTLSClient(timeout time.Duration) *http.Client {
return &http.Client{
Transport: &http.Transport{
TLSClientConfig: &tls.Config{
InsecureSkipVerify: true, //nolint:gosec // Test client for self-signed certs.
},
},
Timeout: timeout,
}
}

// NewMockServerOK creates a mock HTTP server that always returns 200 OK.
// Used for testing health check endpoints.
func NewMockServerOK() *httptest.Server {
return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))
}

// NewMockServerError creates a mock HTTP server that always returns 503 Service Unavailable.
// Used for testing error handling in health check endpoints.
func NewMockServerError() *httptest.Server {
return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusServiceUnavailable)
}))
}

// NewMockServerSlow creates a mock HTTP server with configurable delay before responding.
// Used for testing timeout handling in health check endpoints.
func NewMockServerSlow(delay time.Duration) *httptest.Server {
return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
time.Sleep(delay)
w.WriteHeader(http.StatusOK)
}))
}
